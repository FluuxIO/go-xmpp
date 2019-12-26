package xmpp

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"testing"
	"time"

	"gosrc.io/xmpp/stanza"
)

const (
	// Default port is not standard XMPP port to avoid interfering
	// with local running XMPP server
	testXMPPAddress  = "localhost:15222"
	testClientDomain = "localhost"
)

func TestEventManager(t *testing.T) {
	mgr := EventManager{}
	mgr.updateState(StateConnected)
	if mgr.CurrentState != StateConnected {
		t.Fatal("CurrentState not updated by updateState()")
	}

	mgr.disconnected(SMState{})
	if mgr.CurrentState != StateDisconnected {
		t.Fatalf("CurrentState not reset by disconnected()")
	}

	mgr.streamError(ErrTLSNotSupported.Error(), "")
	if mgr.CurrentState != StateStreamError {
		t.Fatalf("CurrentState not set by streamError()")
	}
}

func TestClient_Connect(t *testing.T) {
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, handlerClientConnectSuccess)

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
		Insecure:   true}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	if err = client.Connect(); err != nil {
		t.Errorf("XMPP connection failed: %s", err)
	}

	mock.Stop()
}

func TestClient_NoInsecure(t *testing.T) {
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		handlerAbortTLS(t, sc)
		closeConn(t, sc)
	})

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
	}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("cannot create XMPP client: %s", err)
	}

	if err = client.Connect(); err == nil {
		// When insecure is not allowed:
		t.Errorf("should fail as insecure connection is not allowed and server does not support TLS")
	}

	mock.Stop()
}

// Check that the client is properly tracking features, as session negotiation progresses.
func TestClient_FeaturesTracking(t *testing.T) {
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		handlerAbortTLS(t, sc)
		closeConn(t, sc)
	})

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
	}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("cannot create XMPP client: %s", err)
	}

	if err = client.Connect(); err == nil {
		// When insecure is not allowed:
		t.Errorf("should fail as insecure connection is not allowed and server does not support TLS")
	}

	mock.Stop()
}

func TestClient_RFC3921Session(t *testing.T) {
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, handlerClientConnectWithSession)

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
		Insecure:   true,
	}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	if err = client.Connect(); err != nil {
		t.Errorf("XMPP connection failed: %s", err)
	}

	mock.Stop()
}

// Testing sending an IQ to the mock server and reading its response.
func TestClient_SendIQ(t *testing.T) {
	done := make(chan struct{})
	// Handler for Mock server
	h := func(t *testing.T, sc *ServerConn) {
		handlerClientConnectSuccess(t, sc)
		discardPresence(t, sc)
		respondToIQ(t, sc)
		done <- struct{}{}
	}
	client, mock := mockClientConnection(t, h, testClientIqPort)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	iqReq := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, From: "test1@localhost/mremond-mbp", To: defaultServerName, Id: defaultStreamID, Lang: "en"})
	disco := iqReq.DiscoInfo()
	iqReq.Payload = disco

	// Handle a possible error
	errChan := make(chan error)
	errorHandler := func(err error) {
		errChan <- err
	}
	client.ErrorHandler = errorHandler
	res, err := client.SendIQ(ctx, iqReq)
	if err != nil {
		t.Errorf(err.Error())
	}

	select {
	case <-res: // If the server responds with an IQ, we pass the test
	case err := <-errChan: // If the server sends an error, or there is a connection error
		cancel()
		t.Fatal(err.Error())
	case <-time.After(defaultChannelTimeout): // If we timeout
		cancel()
		t.Fatal("Failed to receive response, to sent IQ, from mock server")
	}
	select {
	case <-done:
		mock.Stop()
	case <-time.After(defaultChannelTimeout):
		cancel()
		t.Fatal("The mock server failed to finish its job !")
	}
	cancel()
}

func TestClient_SendIQFail(t *testing.T) {
	done := make(chan struct{})
	// Handler for Mock server
	h := func(t *testing.T, sc *ServerConn) {
		handlerClientConnectSuccess(t, sc)
		discardPresence(t, sc)
		respondToIQ(t, sc)
		done <- struct{}{}
	}
	client, mock := mockClientConnection(t, h, testClientIqFailPort)

	//==================
	// Create an IQ to send
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	iqReq := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, From: "test1@localhost/mremond-mbp", To: defaultServerName, Id: defaultStreamID, Lang: "en"})
	disco := iqReq.DiscoInfo()
	iqReq.Payload = disco
	// Removing the id to make the stanza invalid. The IQ constructor makes a random one if none is specified
	// so we need to overwrite it.
	iqReq.Id = ""

	// Handle a possible error
	errChan := make(chan error)
	errorHandler := func(err error) {
		errChan <- err
	}
	client.ErrorHandler = errorHandler
	res, _ := client.SendIQ(ctx, iqReq)

	// Test
	select {
	case <-res: // If the server responds with an IQ
		t.Errorf("Server should not respond with an IQ since the request is expected to be invalid !")
	case <-errChan: // If the server sends an error, the test passes
	case <-time.After(defaultChannelTimeout): // If we timeout
		t.Errorf("Failed to receive response, to sent IQ, from mock server")
	}
	select {
	case <-done:
		mock.Stop()
	case <-time.After(defaultChannelTimeout):
		cancel()
		t.Errorf("The mock server failed to finish its job !")
	}
	cancel()
}

func TestClient_SendRaw(t *testing.T) {
	done := make(chan struct{})
	// Handler for Mock server
	h := func(t *testing.T, sc *ServerConn) {
		handlerClientConnectSuccess(t, sc)
		discardPresence(t, sc)
		respondToIQ(t, sc)
		closeConn(t, sc)
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
		port:      testClientRawPort + 100,
	}
	// Sending an IQ with a missing ID. Should err
	testRequests["IQ with missing ID"] = testCase{
		req:       `<iq type="get" from="test1@localhost/mremond-mbp" to="testServer" lang="en"><query xmlns="http://jabber.org/protocol/disco#info"></query></iq>`,
		shouldErr: true,
		port:      testClientRawPort,
	}

	// A handler for the client.
	// In the failing test, the server returns a stream error, which triggers this handler, client side.
	errChan := make(chan error)
	errHandler := func(err error) {
		errChan <- err
	}

	// Tests for all the IQs
	for name, tcase := range testRequests {
		t.Run(name, func(st *testing.T) {
			//Connecting to a mock server, initialized with given port and handler function
			c, m := mockClientConnection(t, h, tcase.port)
			c.ErrorHandler = errHandler
			// Sending raw xml from test case
			err := c.SendRaw(tcase.req)
			if err != nil {
				t.Errorf("Error sending Raw string")
			}
			// Just wait a little so the message has time to arrive
			select {
			// We don't use the default "long" timeout here because waiting it out means passing the test.
			case <-time.After(100 * time.Millisecond):
				c.Disconnect()
			case err = <-errChan:
				if err == nil && tcase.shouldErr {
					t.Errorf("Failed to get closing stream err")
				} else if err != nil && !tcase.shouldErr {
					t.Errorf("This test is not supposed to err !")
				}
			}
			select {
			case <-done:
				m.Stop()
			case <-time.After(defaultChannelTimeout):
				t.Errorf("The mock server failed to finish its job !")
			}
		})
	}
}

func TestClient_Disconnect(t *testing.T) {
	c, m := mockClientConnection(t, func(t *testing.T, sc *ServerConn) {
		handlerClientConnectSuccess(t, sc)
		closeConn(t, sc)
	}, testClientBasePort)
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

func TestClient_DisconnectStreamManager(t *testing.T) {
	// Init mock server
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		handlerAbortTLS(t, sc)
		closeConn(t, sc)
	})

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
	}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("cannot create XMPP client: %s", err)
	}

	sman := NewStreamManager(client, nil)
	errChan := make(chan error)
	runSMan := func(errChan chan error) {
		errChan <- sman.Run()
	}

	go runSMan(errChan)
	select {
	case <-errChan:
	case <-time.After(defaultChannelTimeout):
		// When insecure is not allowed:
		t.Errorf("should fail as insecure connection is not allowed and server does not support TLS")
	}
	mock.Stop()
}

//=============================================================================
// Basic XMPP Server Mock Handlers.

// Test connection with a basic straightforward workflow
func handlerClientConnectSuccess(t *testing.T, sc *ServerConn) {
	checkClientOpenStream(t, sc)
	sendStreamFeatures(t, sc) // Send initial features
	readAuth(t, sc.decoder)
	fmt.Fprintln(sc.connection, "<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>")

	checkClientOpenStream(t, sc) // Reset stream
	sendBindFeature(t, sc)       // Send post auth features
	bind(t, sc)
}

// closeConn closes the connection on request from the client
func closeConn(t *testing.T, sc *ServerConn) {
	for {
		cls, err := stanza.NextPacket(sc.decoder)
		if err != nil {
			t.Errorf("cannot read from socket: %s", err)
			return
		}
		switch cls.(type) {
		case stanza.StreamClosePacket:
			fmt.Fprintf(sc.connection, stanza.StreamClose)
			return
		}
	}

}

// We expect client will abort on TLS
func handlerAbortTLS(t *testing.T, sc *ServerConn) {
	checkClientOpenStream(t, sc)
	sendStreamFeatures(t, sc) // Send initial features
}

// Test connection with mandatory session (RFC-3921)
func handlerClientConnectWithSession(t *testing.T, sc *ServerConn) {
	checkClientOpenStream(t, sc)

	sendStreamFeatures(t, sc) // Send initial features
	readAuth(t, sc.decoder)
	fmt.Fprintln(sc.connection, "<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>")

	checkClientOpenStream(t, sc) // Reset stream
	sendRFC3921Feature(t, sc)    // Send post auth features
	bind(t, sc)
	session(t, sc)
}

func checkClientOpenStream(t *testing.T, sc *ServerConn) {
	sc.connection.SetDeadline(time.Now().Add(defaultTimeout))
	defer sc.connection.SetDeadline(time.Time{})

	for { // TODO clean up. That for loop is not elegant and I prefer bounded recursion.
		var token xml.Token
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
			if _, err := fmt.Fprintf(sc.connection, serverStreamOpen, "localhost", "streamid1", stanza.NSClient, stanza.NSStream); err != nil {
				t.Errorf("cannot write server stream open: %s", err)
			}
			return
		}
	}
}

func mockClientConnection(t *testing.T, serverHandler func(*testing.T, *ServerConn), port int) (*Client, *ServerMock) {
	mock := &ServerMock{}
	testServerAddress := fmt.Sprintf("%s:%d", testClientDomain, port)

	mock.Start(t, testServerAddress, serverHandler)

	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testServerAddress,
		},
		Jid:        "test@localhost",
		Credential: Password("test"),
		Insecure:   true}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(config, router, clientDefaultErrorHandler); err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	if err = client.Connect(); err != nil {
		t.Errorf("XMPP connection failed: %s", err)
	}

	return client, mock
}

// This really should not be used as is.
// It's just meant to be a placeholder when error handling is not needed at this level
func clientDefaultErrorHandler(err error) {
}
