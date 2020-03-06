package xmpp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"gosrc.io/xmpp/stanza"
	"strconv"
	"testing"
	"time"
)

const (
	streamManagementID = "test-stream_management-id"
)

func TestClient_Send(t *testing.T) {
	buffer := bytes.NewBufferString("")
	client := Client{}
	data := []byte("https://da.wikipedia.org/wiki/J%C3%A6vnd%C3%B8gn")
	if err := client.sendWithWriter(buffer, data); err != nil {
		t.Errorf("Writing failed: %v", err)
	}

	if buffer.String() != string(data) {
		t.Errorf("Incorrect value sent to buffer: '%s'", buffer.String())
	}
}

// Stream management test.
// Connection is established, then the server sends supported features and so on.
// After the bind, client attempts a stream management enablement, and server replies in kind.
func Test_StreamManagement(t *testing.T) {
	serverDone := make(chan struct{})
	clientDone := make(chan struct{})

	client, mock := initSrvCliForResumeTests(t, func(t *testing.T, sc *ServerConn) {
		checkClientOpenStream(t, sc)

		sendStreamFeatures(t, sc) // Send initial features
		readAuth(t, sc.decoder)
		sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

		checkClientOpenStream(t, sc)       // Reset stream
		sendFeaturesStreamManagment(t, sc) // Send post auth features
		bind(t, sc)
		enableStreamManagement(t, sc, false, true)
		serverDone <- struct{}{}
	}, testClientStreamManagement, true, true)
	go func() {
		var state SMState
		var err error
		// Client is ok, we now open XMPP session
		if client.Session, err = NewSession(client, state); err != nil {
			t.Fatalf("failed to open XMPP session: %s", err)
		}
		clientDone <- struct{}{}
	}()

	waitForEntity(t, clientDone)
	waitForEntity(t, serverDone)
	mock.Stop()
}

// Absence of stream management test.
// Connection is established, then the server sends supported features and so on.
// Client has stream management disabled in its config, and should not ask for it. Server is not set up to reply.
func Test_NoStreamManagement(t *testing.T) {
	serverDone := make(chan struct{})
	clientDone := make(chan struct{})

	// Setup Mock server
	client, mock := initSrvCliForResumeTests(t, func(t *testing.T, sc *ServerConn) {
		checkClientOpenStream(t, sc)

		sendStreamFeatures(t, sc) // Send initial features
		readAuth(t, sc.decoder)
		sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

		checkClientOpenStream(t, sc)         // Reset stream
		sendFeaturesNoStreamManagment(t, sc) // Send post auth features
		bind(t, sc)
		serverDone <- struct{}{}
	}, testClientStreamManagement, true, false)

	go func() {
		var state SMState

		// Client is ok, we now open XMPP session
		var err error
		if client.Session, err = NewSession(client, state); err != nil {
			t.Fatalf("failed to open XMPP session: %s", err)
		}
		clientDone <- struct{}{}
	}()

	waitForEntity(t, clientDone)
	waitForEntity(t, serverDone)

	mock.Stop()
}

func Test_StreamManagementNotSupported(t *testing.T) {
	serverDone := make(chan struct{})
	clientDone := make(chan struct{})

	client, mock := initSrvCliForResumeTests(t, func(t *testing.T, sc *ServerConn) {
		checkClientOpenStream(t, sc)

		sendStreamFeatures(t, sc) // Send initial features
		readAuth(t, sc.decoder)
		sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

		checkClientOpenStream(t, sc)         // Reset stream
		sendFeaturesNoStreamManagment(t, sc) // Send post auth features
		bind(t, sc)
		serverDone <- struct{}{}
	}, testClientStreamManagement, true, true)

	go func() {
		var state SMState
		var err error
		// Client is ok, we now open XMPP session
		if client.Session, err = NewSession(client, state); err != nil {
			t.Fatalf("failed to open XMPP session: %s", err)
		}
		clientDone <- struct{}{}
	}()

	// Wait for client
	waitForEntity(t, clientDone)

	// Check if client got a positive stream management response from the server
	if client.Session.Features.DoesStreamManagement() {
		t.Fatalf("server does not provide stream management")
	}

	// Wait for server
	waitForEntity(t, serverDone)
	mock.Stop()
}

func Test_StreamManagementNoResume(t *testing.T) {
	serverDone := make(chan struct{})
	clientDone := make(chan struct{})

	client, mock := initSrvCliForResumeTests(t, func(t *testing.T, sc *ServerConn) {
		checkClientOpenStream(t, sc)

		sendStreamFeatures(t, sc) // Send initial features
		readAuth(t, sc.decoder)
		sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

		checkClientOpenStream(t, sc)       // Reset stream
		sendFeaturesStreamManagment(t, sc) // Send post auth features
		bind(t, sc)
		enableStreamManagement(t, sc, false, false)
		serverDone <- struct{}{}
	}, testClientStreamManagement, true, true)

	go func() {
		var state SMState
		var err error
		// Client is ok, we now open XMPP session
		if client.Session, err = NewSession(client, state); err != nil {
			t.Fatalf("failed to open XMPP session: %s", err)
		}
		clientDone <- struct{}{}
	}()
	waitForEntity(t, clientDone)
	if IsStreamResumable(client) {
		t.Fatalf("server does not support resumption but client says stream is resumable")
	}
	waitForEntity(t, serverDone)
	mock.Stop()
}

func Test_StreamManagementResume(t *testing.T) {
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		checkClientOpenStream(t, sc)

		sendStreamFeatures(t, sc) // Send initial features
		readAuth(t, sc.decoder)
		sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

		checkClientOpenStream(t, sc)       // Reset stream
		sendFeaturesStreamManagment(t, sc) // Send post auth features
		bind(t, sc)
		enableStreamManagement(t, sc, false, true)
		discardPresence(t, sc)
	})

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:                    "test@localhost",
		Credential:             Password("test"),
		Insecure:               true,
		StreamManagementEnable: true,
		streamManagementResume: true} // Enable stream management

	var client *Client
	router := NewRouter()
	client, err := NewClient(&config, router, clientDefaultErrorHandler)
	if err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	err = client.Connect()
	if err != nil {
		t.Fatalf("could not connect client to mock server: %s", err)
	}

	statusCorrectChan := make(chan struct{})
	kill := make(chan struct{})

	transp, ok := client.transport.(*XMPPTransport)
	if !ok {
		t.Fatalf("problem with client transport ")
	}

	transp.conn.Close()
	mock.Stop()

	// Check if status is correctly updated because of the disconnect
	go checkClientResumeStatus(client, statusCorrectChan, kill)
	select {
	case <-statusCorrectChan:
	//	Test passed
	case <-time.After(5 * time.Second):
		kill <- struct{}{}
		t.Fatalf("Client is not in disconnected state while it should be. Timed out")
	}

	// Check if the client can have its connection resumed using its state but also its configuration
	if !IsStreamResumable(client) {
		t.Fatalf("should support resumption")
	}

	// Reboot server. We need to make a new one because (at least for now) the mock server can only have one handler
	// and they should be different between a first connection and a stream resume since exchanged messages
	// are different (See XEP-0198)
	mock2 := ServerMock{}
	mock2.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		//	Reconnect
		checkClientOpenStream(t, sc)

		sendStreamFeatures(t, sc) // Send initial features
		readAuth(t, sc.decoder)
		sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

		checkClientOpenStream(t, sc)       // Reset stream
		sendFeaturesStreamManagment(t, sc) // Send post auth features
		resumeStream(t, sc)
	})

	// Reconnect
	err = client.Resume()
	if err != nil {
		t.Fatalf("could not connect client to mock server: %s", err)
	}
	mock2.Stop()
}

func Test_StreamManagementFail(t *testing.T) {
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		checkClientOpenStream(t, sc)

		sendStreamFeatures(t, sc) // Send initial features
		readAuth(t, sc.decoder)
		sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

		checkClientOpenStream(t, sc)       // Reset stream
		sendFeaturesStreamManagment(t, sc) // Send post auth features
		bind(t, sc)
		enableStreamManagement(t, sc, true, true)
	})

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:                    "test@localhost",
		Credential:             Password("test"),
		Insecure:               true,
		StreamManagementEnable: true,
		streamManagementResume: true} // Enable stream management

	var client *Client
	router := NewRouter()
	client, err := NewClient(&config, router, clientDefaultErrorHandler)
	if err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	var state SMState
	_, err = client.transport.Connect()
	if err != nil {
		return
	}

	// Client is ok, we now open XMPP session
	if client.Session, err = NewSession(client, state); err == nil {
		t.Fatalf("test is supposed to err")
	}
	if client.Session.SMState.StreamErrorGroup == nil {
		t.Fatalf("error was not stored correctly in session state")
	}

	mock.Stop()
}

func Test_SendStanzaQueueWithSM(t *testing.T) {
	// Setup Mock server
	mock := ServerMock{}
	serverDone := make(chan struct{})
	mock.Start(t, testXMPPAddress, func(t *testing.T, sc *ServerConn) {
		checkClientOpenStream(t, sc)

		sendStreamFeatures(t, sc) // Send initial features
		readAuth(t, sc.decoder)
		sc.connection.Write([]byte("<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>"))

		checkClientOpenStream(t, sc)       // Reset stream
		sendFeaturesStreamManagment(t, sc) // Send post auth features
		bind(t, sc)
		enableStreamManagement(t, sc, false, true)

		// Ignore the initial presence sent to the server by the client so we can move on to the next packet.
		discardPresence(t, sc)

		// Used here to silently discard the IQ sent by the client, in order to later trigger a resend
		skipPacket(t, sc)
		// Respond to the client ACK request with a number of processed stanzas of 0. This should trigger a resend
		// of previously ignored stanza to the server, which this handler element will be expecting.
		respondWithAck(t, sc, 0, serverDone)
	})

	// Test / Check result
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testXMPPAddress,
		},
		Jid:                    "test@localhost",
		Credential:             Password("test"),
		Insecure:               true,
		StreamManagementEnable: true,
		streamManagementResume: true} // Enable stream management

	var client *Client
	router := NewRouter()
	client, err := NewClient(&config, router, clientDefaultErrorHandler)
	if err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	err = client.Connect()

	client.SendRaw(`<iq id='ls72g593' type='get'>
  <query xmlns='jabber:iq:roster'/>
</iq>
`)

	// Last stanza was discarded silently by the server. Let's ask an ack for it. This should trigger resend as the server
	// will respond with an acknowledged number of stanzas of 0.
	r := stanza.SMRequest{}
	client.Send(r)

	select {
	case <-time.After(defaultChannelTimeout):
		t.Fatalf("server failed to complete the test in time")
	case <-serverDone:
		// Test completed successfully
	}

	mock.Stop()
}

//========================================================================
// Helper functions for tests

func skipPacket(t *testing.T, sc *ServerConn) {
	var p stanza.IQ
	se, err := stanza.NextStart(sc.decoder)

	if err != nil {
		t.Fatalf("cannot read packet: %s", err)
		return
	}
	if err := sc.decoder.DecodeElement(&p, &se); err != nil {
		t.Fatalf("cannot decode packet: %s", err)
		return
	}
}

func respondWithAck(t *testing.T, sc *ServerConn, h int, serverDone chan struct{}) {

	//  Mock server reads the ack request
	var p stanza.SMRequest
	se, err := stanza.NextStart(sc.decoder)

	if err != nil {
		t.Fatalf("cannot read packet: %s", err)
		return
	}
	if err := sc.decoder.DecodeElement(&p, &se); err != nil {
		t.Fatalf("cannot decode packet: %s", err)
		return
	}

	// Mock server sends the ack response
	a := stanza.SMAnswer{
		H: uint(h),
	}
	data, err := xml.Marshal(a)
	_, err = sc.connection.Write(data)
	if err != nil {
		t.Fatalf("failed to send response ack")
	}

	// Mock server reads the re-sent stanza that was previously discarded intentionally
	var p2 stanza.IQ
	nse, err := stanza.NextStart(sc.decoder)

	if err != nil {
		t.Fatalf("cannot read packet: %s", err)
		return
	}
	if err := sc.decoder.DecodeElement(&p2, &nse); err != nil {
		t.Fatalf("cannot decode packet: %s", err)
		return
	}
	serverDone <- struct{}{}
}

func sendFeaturesStreamManagment(t *testing.T, sc *ServerConn) {
	// This is a basic server, supporting only 2 features after auth: stream management & session binding
	features := `<stream:features>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/>
  <sm xmlns='urn:xmpp:sm:3'/>
</stream:features>`
	if _, err := fmt.Fprintln(sc.connection, features); err != nil {
		t.Fatalf("cannot send stream feature: %s", err)
	}
}

func sendFeaturesNoStreamManagment(t *testing.T, sc *ServerConn) {
	// This is a basic server, supporting only 2 features after auth: stream management & session binding
	features := `<stream:features>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/>
</stream:features>`
	if _, err := fmt.Fprintln(sc.connection, features); err != nil {
		t.Fatalf("cannot send stream feature: %s", err)
	}
}

// enableStreamManagement is a function for the mock server that can either mock a successful session, or fail depending on
// the value of the "fail" boolean. True means the session should fail.
func enableStreamManagement(t *testing.T, sc *ServerConn, fail bool, resume bool) {
	// Decode element into pointer storage
	var ed stanza.SMEnable
	se, err := stanza.NextStart(sc.decoder)

	if err != nil {
		t.Fatalf("cannot read stream management enable: %s", err)
		return
	}
	if err := sc.decoder.DecodeElement(&ed, &se); err != nil {
		t.Fatalf("cannot decode stream management enable: %s", err)
		return
	}

	if fail {
		f := stanza.SMFailed{
			H:                nil,
			StreamErrorGroup: &stanza.UnexpectedRequest{},
		}
		data, err := xml.Marshal(f)
		if err != nil {
			t.Fatalf("failed to marshall error response: %s", err)
		}
		sc.connection.Write(data)
	} else {
		e := &stanza.SMEnabled{
			Resume: strconv.FormatBool(resume),
			Id:     streamManagementID,
		}
		data, err := xml.Marshal(e)
		if err != nil {
			t.Fatalf("failed to marshall error response: %s", err)
		}
		sc.connection.Write(data)
	}
}

func resumeStream(t *testing.T, sc *ServerConn) {
	h := uint(0)
	response := stanza.SMResumed{
		PrevId: streamManagementID,
		H:      &h,
	}

	data, err := xml.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshall stream management enabled response : %s", err)
	}

	writtenChan := make(chan struct{})

	go func() {
		sc.connection.Write(data)
		writtenChan <- struct{}{}
	}()
	select {
	case <-writtenChan:
		// We're done here
		return
	case <-time.After(defaultTimeout):
		t.Fatalf("failed to write enabled nonza to client")
	}
}

func checkClientResumeStatus(client *Client, statusCorrectChan chan struct{}, killChan chan struct{}) {
	for {
		if client.CurrentState.getState() == StateDisconnected {
			statusCorrectChan <- struct{}{}
		}
		select {
		case <-killChan:
			return
		case <-time.After(time.Millisecond * 10):
			//	Keep checking status value
		}
	}
}

func initSrvCliForResumeTests(t *testing.T, serverHandler func(*testing.T, *ServerConn), port int, StreamManagementEnable, StreamManagementResume bool) (*Client, *ServerMock) {
	mock := &ServerMock{}
	testServerAddress := fmt.Sprintf("%s:%d", testClientDomain, port)

	mock.Start(t, testServerAddress, serverHandler)
	config := Config{
		TransportConfiguration: TransportConfiguration{
			Address: testServerAddress,
		},
		Jid:                    "test@localhost",
		Credential:             Password("test"),
		Insecure:               true,
		StreamManagementEnable: StreamManagementEnable,
		streamManagementResume: StreamManagementResume}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(&config, router, clientDefaultErrorHandler); err != nil {
		t.Fatalf("connect create XMPP client: %s", err)
	}

	if _, err = client.transport.Connect(); err != nil {
		t.Fatalf("XMPP connection failed: %s", err)
	}

	return client, mock
}

func waitForEntity(t *testing.T, entityDone chan struct{}) {
	select {
	case <-entityDone:
	case <-time.After(defaultTimeout):
		t.Fatalf("test timed out")
	}
}
