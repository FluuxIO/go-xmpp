package xmpp

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"gosrc.io/xmpp/stanza"
	"net"
	"strings"
	"testing"
	"time"
)

// Tests are ran in parallel, so each test creating a server must use a different port so we do not get any
// conflict. Using iota for this should do the trick.
const (
	testComponentDomain  = "localhost"
	defaultServerName    = "testServer"
	defaultStreamID      = "91bd0bba-012f-4d92-bb17-5fc41e6fe545"
	defaultComponentName = "Test Component"

	// Default port is not standard XMPP port to avoid interfering
	// with local running XMPP server
	testHandshakePort = iota + 15222
	testDecoderPort
	testSendIqPort
	testSendRawPort
	testDisconnectPort
	testSManDisconnectPort
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
// Tests multiple session IDs. All connections should generate a unique stream ID
func TestGenerateHandshake(t *testing.T) {
	// Using this array with a channel to make a queue of values to test
	// These are stream IDs that will be used to test the connection process, mixing them with the "secret" to generate
	// some handshake value
	var uuidsArray = [5]string{
		"cc9b3249-9582-4780-825f-4311b42f9b0e",
		"bba8be3c-d98e-4e26-b9bb-9ed34578a503",
		"dae72822-80e8-496b-b763-ab685f53a188",
		"a45d6c06-de49-4bb0-935b-1a2201b71028",
		"7dc6924f-0eca-4237-9898-18654b8d891e",
	}

	// Channel to pass stream IDs as a queue
	var uchan = make(chan string, len(uuidsArray))
	// Populate test channel
	for _, elt := range uuidsArray {
		uchan <- elt
	}

	// Performs a Component connection with a handshake. It expects to have an ID sent its way through the "uchan"
	// channel of this file. Otherwise it will hang for ever.
	h := func(t *testing.T, c net.Conn) {
		decoder := xml.NewDecoder(c)
		checkOpenStreamHandshakeID(t, c, decoder, <-uchan)
		readHandshakeComponent(t, decoder)
		fmt.Fprintln(c, "<handshake/>") // That's all the server needs to return (see xep-0114)
		return
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
	c, err := NewComponent(opts, router)
	if err != nil {
		t.Errorf("%+v", err)
	}
	c.transport, err = NewComponentTransport(c.ComponentOptions.TransportConfiguration)
	if err != nil {
		t.Errorf("%+v", err)
	}

	// Try connecting, and storing the resulting streamID in a map.
	m := make(map[string]bool)
	for _, _ = range uuidsArray {
		streamId, _ := c.transport.Connect()
		m[c.handshake(streamId)] = true
	}
	if len(uuidsArray) != len(m) {
		t.Errorf("Handshake does not produce a unique id. Expected: %d unique ids, got: %d", len(uuidsArray), len(m))
	}
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
	c, _ := mockConnection(t, testDecoderPort, handlerForComponentHandshakeDefaultID)
	if c.transport.GetDecoder() == nil {
		t.Errorf("Failed to initialize decoder. Decoder is nil.")
	}
}

// Tests sending an IQ to the server, and getting the response
func TestSendIq(t *testing.T) {
	//Connecting to a mock server, initialized with given port and handler function
	c, m := mockConnection(t, testSendIqPort, handlerForComponentIQSend)

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	iqReq := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, From: "test1@localhost/mremond-mbp", To: defaultServerName, Id: defaultStreamID, Lang: "en"})
	disco := iqReq.DiscoInfo()
	iqReq.Payload = disco

	var res chan stanza.IQ
	res, _ = c.SendIQ(ctx, iqReq)

	select {
	case <-res:
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Failed to receive response, to sent IQ, from mock server")
	}

	m.Stop()
}

// Tests sending raw xml to the mock server.
// TODO : check the server response client side ?
// Right now, the server response is not checked and an err is passed in a channel if the test is supposed to err.
// In this test, we use IQs
func TestSendRaw(t *testing.T) {
	// Error channel for the handler
	errChan := make(chan error)
	// Handler for the mock server
	h := func(t *testing.T, c net.Conn) {
		// Completes the connection by exchanging handshakes
		handlerForComponentHandshakeDefaultID(t, c)
		receiveRawIq(t, c, errChan)
		return
	}

	type testCase struct {
		req       string
		shouldErr bool
	}
	testRequests := make(map[string]testCase)
	// Sending a correct IQ of type get. Not supposed to err
	testRequests["Correct IQ"] = testCase{
		req:       `<iq type="get" id="91bd0bba-012f-4d92-bb17-5fc41e6fe545" from="test1@localhost/mremond-mbp" to="testServer" lang="en"><query xmlns="http://jabber.org/protocol/disco#info"></query></iq>`,
		shouldErr: false,
	}
	// Sending an IQ with a missing ID. Should err
	testRequests["IQ with missing ID"] = testCase{
		req:       `<iq type="get" from="test1@localhost/mremond-mbp" to="testServer" lang="en"><query xmlns="http://jabber.org/protocol/disco#info"></query></iq>`,
		shouldErr: true,
	}

	// Tests for all the IQs
	for name, tcase := range testRequests {
		t.Run(name, func(st *testing.T) {
			//Connecting to a mock server, initialized with given port and handler function
			c, m := mockConnection(t, testSendRawPort, h)

			// Sending raw xml from test case
			err := c.SendRaw(tcase.req)
			if err != nil {
				t.Errorf("Error sending Raw string")
			}
			// Just wait a little so the message has time to arrive
			select {
			case <-time.After(100 * time.Millisecond):
			case err = <-errChan:
				if err == nil && tcase.shouldErr {
					t.Errorf("Failed to get closing stream err")
				}
			}
			c.transport.Close()
			m.Stop()
		})
	}
}

// Tests the Disconnect method for Components
func TestDisconnect(t *testing.T) {
	c, m := mockConnection(t, testDisconnectPort, handlerForComponentHandshakeDefaultID)
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
// Performs a Component connection with a handshake. It uses a default ID defined in this file as a constant.
// Used in the mock server as a Handler
func handlerForComponentHandshakeDefaultID(t *testing.T, c net.Conn) {
	decoder := xml.NewDecoder(c)
	checkOpenStreamHandshakeDefaultID(t, c, decoder)
	readHandshakeComponent(t, decoder)
	fmt.Fprintln(c, "<handshake/>") // That's all the server needs to return (see xep-0114)
	return
}

// Performs a Component connection with a handshake. It uses a default ID defined in this file as a constant.
// This handler is supposed to fail by sending a "message" stanza instead of a <handshake/> stanza to finalize the handshake.
func handlerComponentFailedHandshakeDefaultID(t *testing.T, c net.Conn) {
	decoder := xml.NewDecoder(c)
	checkOpenStreamHandshakeDefaultID(t, c, decoder)
	readHandshakeComponent(t, decoder)

	// Send a message, instead of a "<handshake/>" tag, to fail the handshake process dans disconnect the client.
	me := stanza.Message{
		Attrs: stanza.Attrs{Type: stanza.MessageTypeChat, From: defaultServerName, To: defaultComponentName, Lang: "en"},
		Body:  "Fail my handshake.",
	}
	s, _ := xml.Marshal(me)
	fmt.Fprintln(c, string(s))

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

func checkOpenStreamHandshakeDefaultID(t *testing.T, c net.Conn, decoder *xml.Decoder) {
	checkOpenStreamHandshakeID(t, c, decoder, defaultStreamID)
}

// Used for ID and handshake related tests
func checkOpenStreamHandshakeID(t *testing.T, c net.Conn, decoder *xml.Decoder, streamID string) {
	c.SetDeadline(time.Now().Add(defaultTimeout))
	defer c.SetDeadline(time.Time{})

	for { // TODO clean up. That for loop is not elegant and I prefer bounded recursion.
		token, err := decoder.Token()
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
			if _, err := fmt.Fprintf(c, serverStreamOpen, "localhost", streamID, stanza.NSComponent, stanza.NSStream); err != nil {
				t.Errorf("cannot write server stream open: %s", err)
			}
			return
		}
	}
}

//=============================================================================
// Sends IQ response to Component request.
// No parsing of the request here. We just check that it's valid, and send the default response.
func handlerForComponentIQSend(t *testing.T, c net.Conn) {
	// Completes the connection by exchanging handshakes
	handlerForComponentHandshakeDefaultID(t, c)

	// Decoder to parse the request
	decoder := xml.NewDecoder(c)

	iqReq, err := receiveIq(t, c, decoder)
	if err != nil {
		t.Errorf("Error receiving the IQ stanza : %v", err)
	} else if !iqReq.IsValid() {
		t.Errorf("server received an IQ stanza : %v", iqReq)
	}

	// Crafting response
	iqResp := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeResult, From: iqReq.To, To: iqReq.From, Id: iqReq.Id, Lang: "en"})
	disco := iqResp.DiscoInfo()
	disco.AddFeatures("vcard-temp",
		`http://jabber.org/protocol/address`)

	disco.AddIdentity("Multicast", "service", "multicast")
	iqResp.Payload = disco

	// Sending response to the Component
	mResp, err := xml.Marshal(iqResp)
	_, err = fmt.Fprintln(c, string(mResp))
	if err != nil {
		t.Errorf("Could not send response stanza : %s", err)
	}
	return
}

// Reads next request coming from the Component. Expecting it to be an IQ request
func receiveIq(t *testing.T, c net.Conn, decoder *xml.Decoder) (stanza.IQ, error) {
	c.SetDeadline(time.Now().Add(defaultTimeout))
	defer c.SetDeadline(time.Time{})
	var iqStz stanza.IQ
	err := decoder.Decode(&iqStz)
	if err != nil {
		t.Errorf("cannot read the received IQ stanza: %s", err)
	}
	if !iqStz.IsValid() {
		t.Errorf("received IQ stanza is invalid : %s", err)
	}
	return iqStz, nil
}

func receiveRawIq(t *testing.T, c net.Conn, errChan chan error) {
	c.SetDeadline(time.Now().Add(defaultTimeout))
	defer c.SetDeadline(time.Time{})
	decoder := xml.NewDecoder(c)
	var iq stanza.IQ
	err := decoder.Decode(&iq)
	if err != nil || !iq.IsValid() {
		s := stanza.StreamError{
			XMLName: xml.Name{Local: "stream:error"},
			Error:   xml.Name{Local: "xml-not-well-formed"},
			Text:    `XML was not well-formed`,
		}
		raw, _ := xml.Marshal(s)
		fmt.Fprintln(c, string(raw))
		fmt.Fprintln(c, `</stream:stream>`) // TODO : check this client side
		errChan <- fmt.Errorf("invalid xml")
		return
	}
	errChan <- nil
	return
}

//===============================
// Init mock server and connection
// Creating a mock server and connecting a Component to it. Initialized with given port and handler function
// The Component and mock are both returned
func mockConnection(t *testing.T, port int, handler func(t *testing.T, c net.Conn)) (*Component, *ServerMock) {
	// Init mock server
	testComponentAddress := fmt.Sprintf("%s:%d", testComponentDomain, port)
	mock := ServerMock{}
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

	return c, &mock
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
	c, err := NewComponent(opts, router)
	if err != nil {
		t.Errorf("%+v", err)
	}
	c.transport, err = NewComponentTransport(c.ComponentOptions.TransportConfiguration)
	if err != nil {
		t.Errorf("%+v", err)
	}
	return c
}
