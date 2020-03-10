package xmpp

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gosrc.io/xmpp/stanza"
)

// Tests are ran in parallel, so each test creating a server must use a different port so we do not get any
// conflict. Using iota for this should do the trick.
const (
	defaultChannelTimeout = 5 * time.Second
)

func TestHandshake(t *testing.T) {
	opts := ComponentOptions{
		Domain: "test.localhost",
		Secret: "mypass",
	}
	c := Component{ComponentOptions: opts}

	streamID := "1263952298440005243"
	expected := "c77e2ef0109fbbc5161e83b51629cd1353495332"

	result := c.handshake(streamID)
	if result != expected {
		t.Errorf("incorrect handshake calculation '%s' != '%s'", result, expected)
	}
}

// Tests connection process with a handshake exchange
// Tests multiple session IDs. All serverConnections should generate a unique stream ID
func TestGenerateHandshakeId(t *testing.T) {
	clientDone := make(chan struct{})
	serverDone := make(chan struct{})
	// Using this array with a channel to make a queue of values to test
	// These are stream IDs that will be used to test the connection process, mixing them with the "secret" to generate
	// some handshake value
	var uuidsArray = [5]string{}
	for i := 1; i < len(uuidsArray); i++ {
		id, _ := uuid.NewRandom()
		uuidsArray[i] = id.String()
	}

	// Channel to pass stream IDs as a queue
	var uchan = make(chan string, len(uuidsArray))
	// Populate test channel
	for _, elt := range uuidsArray {
		uchan <- elt
	}

	// Performs a Component connection with a handshake. It expects to have an ID sent its way through the "uchan"
	// channel of this file. Otherwise it will hang for ever.
	h := func(t *testing.T, sc *ServerConn) {
		checkOpenStreamHandshakeID(t, sc, <-uchan)
		readHandshakeComponent(t, sc.decoder)
		sc.connection.Write([]byte("<handshake/>")) // That's all the server needs to return (see xep-0114)
		serverDone <- struct{}{}
	}

	// Init mock server
	testComponentAddess := fmt.Sprintf("%s:%d", testComponentDomain, testHandshakePort)
	mock := ServerMock{}
	mock.Start(t, testComponentAddess, h)

	// Init component
	opts := ComponentOptions{
		TransportConfiguration: TransportConfiguration{
			Address: testComponentAddess,
			Domain:  "localhost",
		},
		Domain:   testComponentDomain,
		Secret:   "mypass",
		Name:     "Test Component",
		Category: "gateway",
		Type:     "service",
	}
	router := NewRouter()
	c, err := NewComponent(opts, router, componentDefaultErrorHandler)
	if err != nil {
		t.Errorf("%+v", err)
	}
	c.transport, err = NewComponentTransport(c.ComponentOptions.TransportConfiguration)
	if err != nil {
		t.Errorf("%+v", err)
	}

	// Try connecting, and storing the resulting streamID in a map.
	go func() {
		m := make(map[string]bool)
		for range uuidsArray {
			idChan := make(chan string)
			go func() {
				streamId, err := c.transport.Connect()
				if err != nil {
					t.Fatalf("failed to mock component connection to get a handshake: %s", err)
				}
				idChan <- streamId
			}()

			var streamId string
			select {
			case streamId = <-idChan:
			case <-time.After(defaultTimeout):
				t.Fatalf("test timed out")
			}

			hs := stanza.Handshake{
				Value: c.handshake(streamId),
			}
			m[hs.Value] = true
			hsRaw, err := xml.Marshal(hs)
			if err != nil {
				t.Fatalf("could not marshal handshake: %s", err)
			}
			c.SendRaw(string(hsRaw))
			waitForEntity(t, serverDone)
			c.transport.Close()
		}
		if len(uuidsArray) != len(m) {
			t.Errorf("Handshake does not produce a unique id. Expected: %d unique ids, got: %d", len(uuidsArray), len(m))
		}
		clientDone <- struct{}{}
	}()

	waitForEntity(t, clientDone)
	mock.Stop()
}

// Test that NewStreamManager can accept a Component.
//
// This validates that Component conforms to StreamClient interface.
func TestStreamManager(t *testing.T) {
	NewStreamManager(&Component{}, nil)
}

// Tests that the decoder is properly initialized when connecting a component to a server.
// The decoder is expected to be built after a valid connection
// Based on the xmpp_component example.
func TestDecoder(t *testing.T) {
	c, _ := mockComponentConnection(t, testDecoderPort, handlerForComponentHandshakeDefaultID)
	if c.transport.GetDecoder() == nil {
		t.Errorf("Failed to initialize decoder. Decoder is nil.")
	}
}

// Tests sending an IQ to the server, and getting the response
func TestSendIq(t *testing.T) {
	serverDone := make(chan struct{})
	clientDone := make(chan struct{})
	h := func(t *testing.T, sc *ServerConn) {
		handlerForComponentIQSend(t, sc)
		serverDone <- struct{}{}
	}

	//Connecting to a mock server, initialized with given port and handler function
	c, m := mockComponentConnection(t, testSendIqPort, h)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	iqReq, err := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, From: "test1@localhost/mremond-mbp", To: defaultServerName, Id: defaultStreamID, Lang: "en"})
	if err != nil {
		t.Fatalf("failed to create IQ request: %v", err)
	}
	disco := iqReq.DiscoInfo()
	iqReq.Payload = disco

	// Handle a possible error
	errChan := make(chan error)
	errorHandler := func(err error) {
		errChan <- err
	}
	c.ErrorHandler = errorHandler

	go func() {
		var res chan stanza.IQ
		res, _ = c.SendIQ(ctx, iqReq)

		select {
		case <-res:
		case err := <-errChan:
			t.Fatalf(err.Error())
		}
		clientDone <- struct{}{}
	}()

	waitForEntity(t, clientDone)
	waitForEntity(t, serverDone)

	cancel()
	m.Stop()
}

// Checking that error handling is done properly client side when an invalid IQ is sent and the server responds in kind.
func TestSendIqFail(t *testing.T) {
	done := make(chan struct{})
	h := func(t *testing.T, sc *ServerConn) {
		handlerForComponentIQSend(t, sc)
		done <- struct{}{}
	}
	//Connecting to a mock server, initialized with given port and handler function
	c, m := mockComponentConnection(t, testSendIqFailPort, h)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	iqReq, err := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, From: "test1@localhost/mremond-mbp", To: defaultServerName, Id: defaultStreamID, Lang: "en"})
	if err != nil {
		t.Fatalf("failed to create IQ request: %v", err)
	}

	// Removing the id to make the stanza invalid. The IQ constructor makes a random one if none is specified
	// so we need to overwrite it.
	iqReq.Id = ""
	disco := iqReq.DiscoInfo()
	iqReq.Payload = disco

	errChan := make(chan error)
	errorHandler := func(err error) {
		errChan <- err
	}
	c.ErrorHandler = errorHandler

	var res chan stanza.IQ
	res, _ = c.SendIQ(ctx, iqReq)

	select {
	case r := <-res: // Do we get an IQ response from the server ?
		t.Errorf("We should not be getting an IQ response here : this should fail !")
		fmt.Println(r)
	case <-errChan: // Do we get a stream error from the server ?
		// If we get an error from the server, the test passes.
	case <-time.After(defaultChannelTimeout): // Timeout ?
		t.Errorf("Failed to receive response, to sent IQ, from mock server")
	}

	select {
	case <-done:
		m.Stop()
	case <-time.After(defaultChannelTimeout):
		t.Errorf("The mock server failed to finish its job !")
	}
	cancel()
}

// Tests sending raw xml to the mock server.
// Right now, the server response is not checked and an err is passed in a channel if the test is supposed to err.
// In this test, we use IQs
func TestSendRaw(t *testing.T) {
	done := make(chan struct{})
	// Handler for the mock server
	h := func(t *testing.T, sc *ServerConn) {
		// Completes the connection by exchanging handshakes
		handlerForComponentHandshakeDefaultID(t, sc)
		respondToIQ(t, sc)
		done <- struct{}{}
	}

	type testCase struct {
		req       string
		shouldErr bool
		port      int
	}
	testRequests := make(map[string]testCase)
	// Sending a correct IQ of type get. Not supposed to err
	testRequests["Correct IQ"] = testCase{
		req:       `<iq type="get" id="91bd0bba-012f-4d92-bb17-5fc41e6fe545" from="test1@localhost/mremond-mbp" to="testServer" lang="en"><query xmlns="http://jabber.org/protocol/disco#info"></query></iq>`,
		shouldErr: false,
		port:      testSendRawPort + 100,
	}
	// Sending an IQ with a missing ID. Should err
	testRequests["IQ with missing ID"] = testCase{
		req:       `<iq type="get" from="test1@localhost/mremond-mbp" to="testServer" lang="en"><query xmlns="http://jabber.org/protocol/disco#info"></query></iq>`,
		shouldErr: true,
		port:      testSendRawPort + 200,
	}

	// A handler for the component.
	// In the failing test, the server returns a stream error, which triggers this handler, component side.
	errChan := make(chan error)
	errHandler := func(err error) {
		errChan <- err
	}

	// Tests for all the IQs
	for name, tcase := range testRequests {
		t.Run(name, func(st *testing.T) {
			//Connecting to a mock server, initialized with given port and handler function
			c, m := mockComponentConnection(t, tcase.port, h)
			c.ErrorHandler = errHandler
			// Sending raw xml from test case
			err := c.SendRaw(tcase.req)
			if err != nil {
				t.Errorf("Error sending Raw string")
			}
			// Just wait a little so the message has time to arrive
			select {
			// We don't use the default "long" timeout here because waiting it out means passing the test.
			case <-time.After(200 * time.Millisecond):
			case err = <-errChan:
				if err == nil && tcase.shouldErr {
					t.Errorf("Failed to get closing stream err")
				} else if err != nil && !tcase.shouldErr {
					t.Errorf("This test is not supposed to err ! => %s", err.Error())
				}
			}
			c.transport.Close()
			select {
			case <-done:
				m.Stop()
			case <-time.After(defaultChannelTimeout):
				t.Errorf("The mock server failed to finish its job !")
			}
		})
	}
}

// Tests the Disconnect method for Components
func TestDisconnect(t *testing.T) {
	c, m := mockComponentConnection(t, testDisconnectPort, handlerForComponentHandshakeDefaultID)
	err := c.transport.Ping()
	if err != nil {
		t.Errorf("Could not ping but not disconnected yet")
	}
	c.Disconnect()
	err = c.transport.Ping()
	if err == nil {
		t.Errorf("Did not disconnect properly")
	}
	m.Stop()
}

// Tests that a streamManager successfully disconnects when a handshake fails between the component and the server.
func TestStreamManagerDisconnect(t *testing.T) {
	// Init mock server
	testComponentAddress := fmt.Sprintf("%s:%d", testComponentDomain, testSManDisconnectPort)
	mock := ServerMock{}
	// Handler fails the handshake, which is currently the only option to disconnect completely when using a streamManager
	// a failed handshake being a permanent error, except for a "conflict"
	mock.Start(t, testComponentAddress, handlerComponentFailedHandshakeDefaultID)

	//==================================
	// Create Component to connect to it
	c := makeBasicComponent(defaultComponentName, testComponentAddress, t)

	//========================================
	// Connect the new Component to the server
	cm := NewStreamManager(c, nil)
	errChan := make(chan error)
	runSMan := func(errChan chan error) {
		errChan <- cm.Run()
	}

	go runSMan(errChan)
	select {
	case <-errChan:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("The component and server seem to still be connected while they should not.")
	}
	mock.Stop()
}

//=============================================================================
// Basic XMPP Server Mock Handlers.

//===============================
// Init mock server and connection
// Creating a mock server and connecting a Component to it. Initialized with given port and handler function
// The Component and mock are both returned
func mockComponentConnection(t *testing.T, port int, handler func(t *testing.T, sc *ServerConn)) (*Component, *ServerMock) {
	// Init mock server
	testComponentAddress := fmt.Sprintf("%s:%d", testComponentDomain, port)
	mock := &ServerMock{}
	mock.Start(t, testComponentAddress, handler)

	//==================================
	// Create Component to connect to it
	c := makeBasicComponent(defaultComponentName, testComponentAddress, t)

	//========================================
	// Connect the new Component to the server
	err := c.Connect()
	if err != nil {
		t.Errorf("%+v", err)
	}

	// Now that the Component is connected, let's set the xml.Decoder for the server

	return c, mock
}

func makeBasicComponent(name string, mockServerAddr string, t *testing.T) *Component {
	opts := ComponentOptions{
		TransportConfiguration: TransportConfiguration{
			Address: mockServerAddr,
			Domain:  "localhost",
		},
		Domain:   testComponentDomain,
		Secret:   "mypass",
		Name:     name,
		Category: "gateway",
		Type:     "service",
	}
	router := NewRouter()
	c, err := NewComponent(opts, router, componentDefaultErrorHandler)
	if err != nil {
		t.Errorf("%+v", err)
	}
	c.transport, err = NewComponentTransport(c.ComponentOptions.TransportConfiguration)
	if err != nil {
		t.Errorf("%+v", err)
	}
	return c
}

// This really should not be used as is.
// It's just meant to be a placeholder when error handling is not needed at this level
func componentDefaultErrorHandler(err error) {

}

// Sends IQ response to Component request.
// No parsing of the request here. We just check that it's valid, and send the default response.
func handlerForComponentIQSend(t *testing.T, sc *ServerConn) {
	// Completes the connection by exchanging handshakes
	handlerForComponentHandshakeDefaultID(t, sc)
	respondToIQ(t, sc)
}

// Used for ID and handshake related tests
func checkOpenStreamHandshakeID(t *testing.T, sc *ServerConn, streamID string) {
	err := sc.connection.SetDeadline(time.Now().Add(defaultTimeout))
	if err != nil {
		t.Fatalf("failed to set deadline: %v", err)
	}
	defer sc.connection.SetDeadline(time.Time{})

	for { // TODO clean up. That for loop is not elegant and I prefer bounded recursion.
		token, err := sc.decoder.Token()
		if err != nil {
			t.Errorf("cannot read next token: %s", err)
		}

		switch elem := token.(type) {
		// Wait for first startElement
		case xml.StartElement:
			if elem.Name.Space != stanza.NSStream || elem.Name.Local != "stream" {
				err = errors.New("xmpp: expected <stream> but got <" + elem.Name.Local + "> in " + elem.Name.Space)
				return
			}
			if _, err := fmt.Fprintf(sc.connection, serverStreamOpen, "localhost", streamID, stanza.NSComponent, stanza.NSStream); err != nil {
				t.Errorf("cannot write server stream open: %s", err)
			}
			return
		}
	}
}

func checkOpenStreamHandshakeDefaultID(t *testing.T, sc *ServerConn) {
	checkOpenStreamHandshakeID(t, sc, defaultStreamID)
}

// Performs a Component connection with a handshake. It uses a default ID defined in this file as a constant.
// This handler is supposed to fail by sending a "message" stanza instead of a <handshake/> stanza to finalize the handshake.
func handlerComponentFailedHandshakeDefaultID(t *testing.T, sc *ServerConn) {
	checkOpenStreamHandshakeDefaultID(t, sc)
	readHandshakeComponent(t, sc.decoder)

	// Send a message, instead of a "<handshake/>" tag, to fail the handshake process dans disconnect the client.
	me := stanza.Message{
		Attrs: stanza.Attrs{Type: stanza.MessageTypeChat, From: defaultServerName, To: defaultComponentName, Lang: "en"},
		Body:  "Fail my handshake.",
	}
	s, _ := xml.Marshal(me)
	_, err := sc.connection.Write(s)
	if err != nil {
		t.Fatalf("could not write message: %v", err)
	}

	return
}

// Reads from the connection with the Component. Expects a handshake request, and returns the <handshake/> tag.
func readHandshakeComponent(t *testing.T, decoder *xml.Decoder) {
	se, err := stanza.NextStart(decoder)
	if err != nil {
		t.Errorf("cannot read auth: %s", err)
		return
	}
	nv := &stanza.Handshake{}
	// Decode element into pointer storage
	if err = decoder.DecodeElement(nv, &se); err != nil {
		t.Errorf("cannot decode handshake: %s", err)
		return
	}
	if len(strings.TrimSpace(nv.Value)) == 0 {
		t.Errorf("did not receive handshake ID")
	}
}

// Performs a Component connection with a handshake. It uses a default ID defined in this file as a constant.
// Used in the mock server as a Handler
func handlerForComponentHandshakeDefaultID(t *testing.T, sc *ServerConn) {
	checkOpenStreamHandshakeDefaultID(t, sc)
	readHandshakeComponent(t, sc.decoder)
	sc.connection.Write([]byte("<handshake/>")) // That's all the server needs to return (see xep-0114)
	return
}
