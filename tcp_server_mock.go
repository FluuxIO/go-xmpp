package xmpp

import (
	"encoding/xml"
	"fmt"
	"gosrc.io/xmpp/stanza"
	"net"
	"testing"
	"time"
)

//=============================================================================
// TCP Server Mock
const (
	defaultTimeout       = 2 * time.Second
	testComponentDomain  = "localhost"
	defaultServerName    = "testServer"
	defaultStreamID      = "91bd0bba-012f-4d92-bb17-5fc41e6fe545"
	defaultComponentName = "Test Component"
	serverStreamOpen     = "<?xml version='1.0'?><stream:stream to='%s' id='%s' xmlns='%s' xmlns:stream='%s' version='1.0'>"

	// Default port is not standard XMPP port to avoid interfering
	// with local running XMPP server

	// Component tests
	testHandshakePort = iota + 15222
	testDecoderPort
	testSendIqPort
	testSendIqFailPort
	testSendRawPort
	testDisconnectPort
	testSManDisconnectPort

	// Client tests
	testClientBasePort
	testClientRawPort
	testClientIqPort
	testClientIqFailPort
	testClientPostConnectHook

	// Client internal tests
	testClientStreamManagement
)

// ClientHandler is passed by the test client to provide custom behaviour to
// the TCP server mock. This allows customizing the server behaviour to allow
// testing clients under various scenarii.
type ClientHandler func(t *testing.T, serverConn *ServerConn)

// ServerMock is a simple TCP server that can be use to mock basic server
// behaviour to test clients.
type ServerMock struct {
	t                 *testing.T
	handler           ClientHandler
	listener          net.Listener
	serverConnections []*ServerConn
	done              chan struct{}
}

type ServerConn struct {
	connection net.Conn
	decoder    *xml.Decoder
}

// Start launches the mock TCP server, listening to an actual address / port.
func (mock *ServerMock) Start(t *testing.T, addr string, handler ClientHandler) {
	mock.t = t
	mock.handler = handler
	if err := mock.init(addr); err != nil {
		return
	}
	go mock.loop()
}

func (mock *ServerMock) Stop() {
	close(mock.done)
	if mock.listener != nil {
		mock.listener.Close()
	}
	// Close all existing serverConnections
	for _, c := range mock.serverConnections {
		c.connection.Close()
	}
}

//=============================================================================
// Mock Server internals

// init starts listener on the provided address.
func (mock *ServerMock) init(addr string) error {
	mock.done = make(chan struct{})

	l, err := net.Listen("tcp", addr)
	if err != nil {
		mock.t.Errorf("TCPServerMock cannot listen on address: %q", addr)
		return err
	}
	mock.listener = l
	return nil
}

// loop accepts serverConnections and creates a go routine per connection.
// The go routine is running the client handler, that is used to provide the
// real TCP server behaviour.
func (mock *ServerMock) loop() {
	listener := mock.listener
	for {
		conn, err := listener.Accept()
		serverConn := &ServerConn{conn, xml.NewDecoder(conn)}
		if err != nil {
			select {
			case <-mock.done:
				return
			default:
				mock.t.Error("TCPServerMock accept error:", err.Error())
			}
			return
		}
		mock.serverConnections = append(mock.serverConnections, serverConn)

		// TODO Create and pass a context to cancel the handler if they are still around = avoid possible leak on complex handlers
		go mock.handler(mock.t, serverConn)
	}
}

//======================================================================================================================
// A few functions commonly used for tests. Trying to avoid duplicates in client and component test files.
//======================================================================================================================

func respondToIQ(t *testing.T, sc *ServerConn) {
	// Decoder to parse the request
	iqReq, err := receiveIq(sc)
	if err != nil {
		t.Fatalf("failed to receive IQ : %s", err.Error())
	}

	if vld, _ := iqReq.IsValid(); !vld {
		mockIQError(sc.connection)
		return
	}

	// Crafting response
	iqResp, err := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeResult, From: iqReq.To, To: iqReq.From, Id: iqReq.Id, Lang: "en"})
	if err != nil {
		t.Fatalf("failed to create iqResp: %v", err)
	}
	disco := iqResp.DiscoInfo()
	disco.AddFeatures("vcard-temp",
		`http://jabber.org/protocol/address`)

	disco.AddIdentity("Multicast", "service", "multicast")
	iqResp.Payload = disco

	// Sending response to the Component
	mResp, err := xml.Marshal(iqResp)
	_, err = fmt.Fprintln(sc.connection, string(mResp))
	if err != nil {
		t.Errorf("Could not send response stanza : %w", err)
	}
	return
}

// When a presence stanza is automatically sent (right now it's the case in the client), we may want to discard it
// and test further stanzas.
func discardPresence(t *testing.T, sc *ServerConn) {
	err := sc.connection.SetDeadline(time.Now().Add(defaultTimeout))
	if err != nil {
		t.Fatalf("failed to set deadline: %v", err)
	}
	defer sc.connection.SetDeadline(time.Time{})
	var presenceStz stanza.Presence

	recvBuf := make([]byte, len(InitialPresence))
	_, err = sc.connection.Read(recvBuf[:]) // recv data

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			t.Errorf("read timeout: %w", err)
		} else {
			t.Errorf("read error: %w", err)
		}
	}
	err = xml.Unmarshal(recvBuf, &presenceStz)

	if err != nil {
		t.Errorf("Expected presence but this happened : %w", err)
	}
}

// Reads next request coming from the Component. Expecting it to be an IQ request
func receiveIq(sc *ServerConn) (*stanza.IQ, error) {
	err := sc.connection.SetDeadline(time.Now().Add(defaultTimeout))
	if err != nil {
		return nil, err
	}
	defer sc.connection.SetDeadline(time.Time{})
	var iqStz stanza.IQ
	err = sc.decoder.Decode(&iqStz)
	if err != nil {
		return nil, err
	}
	return &iqStz, nil
}

// Should be used in server handlers when an IQ sent by a client or component is invalid.
// This responds as expected from a "real" server, aside from the error message.
func mockIQError(c net.Conn) {
	s := stanza.StreamError{
		XMLName: xml.Name{Local: "stream:error"},
		Error:   xml.Name{Local: "xml-not-well-formed"},
		Text:    `XML was not well-formed`,
	}
	raw, _ := xml.Marshal(s)
	fmt.Fprintln(c, string(raw))
	fmt.Fprintln(c, `</stream:stream>`)
}

func sendStreamFeatures(t *testing.T, sc *ServerConn) {
	// This is a basic server, supporting only 1 stream feature: SASL Plain Auth
	features := `<stream:features>
  <mechanisms xmlns="urn:ietf:params:xml:ns:xmpp-sasl">
    <mechanism>PLAIN</mechanism>
  </mechanisms>
</stream:features>`
	if _, err := fmt.Fprintln(sc.connection, features); err != nil {
		t.Errorf("cannot send stream feature: %w", err)
	}
}

// TODO return err in case of error reading the auth params
func readAuth(t *testing.T, decoder *xml.Decoder) string {
	se, err := stanza.NextStart(decoder)
	if err != nil {
		t.Errorf("cannot read auth: %w", err)
		return ""
	}

	var nv interface{}
	nv = &stanza.SASLAuth{}
	// Decode element into pointer storage
	if err = decoder.DecodeElement(nv, &se); err != nil {
		t.Errorf("cannot decode auth: %w", err)
		return ""
	}

	switch v := nv.(type) {
	case *stanza.SASLAuth:
		return v.Value
	}
	return ""
}

func sendBindFeature(t *testing.T, sc *ServerConn) {
	// This is a basic server, supporting only 1 stream feature after auth: resource binding
	features := `<stream:features>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/>
</stream:features>`
	if _, err := fmt.Fprintln(sc.connection, features); err != nil {
		t.Errorf("cannot send stream feature: %w", err)
	}
}

func sendRFC3921Feature(t *testing.T, sc *ServerConn) {
	// This is a basic server, supporting only 2 features after auth: resource & session binding
	features := `<stream:features>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/>
  <session xmlns='urn:ietf:params:xml:ns:xmpp-session'/>
</stream:features>`
	if _, err := fmt.Fprintln(sc.connection, features); err != nil {
		t.Errorf("cannot send stream feature: %w", err)
	}
}

func bind(t *testing.T, sc *ServerConn) {
	se, err := stanza.NextStart(sc.decoder)
	if err != nil {
		t.Errorf("cannot read bind: %w", err)
		return
	}

	iq := &stanza.IQ{}
	// Decode element into pointer storage
	if err = sc.decoder.DecodeElement(&iq, &se); err != nil {
		t.Errorf("cannot decode bind iq: %w", err)
		return
	}

	// TODO Check all elements
	switch iq.Payload.(type) {
	case *stanza.Bind:
		result := `<iq id='%s' type='result'>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'>
  	<jid>%s</jid>
  </bind>
</iq>`
		fmt.Fprintf(sc.connection, result, iq.Id, "test@localhost/test") // TODO use real Jid
	}
}

func session(t *testing.T, sc *ServerConn) {
	se, err := stanza.NextStart(sc.decoder)
	if err != nil {
		t.Errorf("cannot read session: %w", err)
		return
	}

	iq := &stanza.IQ{}
	// Decode element into pointer storage
	if err = sc.decoder.DecodeElement(&iq, &se); err != nil {
		t.Errorf("cannot decode session iq: %w", err)
		return
	}

	switch iq.Payload.(type) {
	case *stanza.StreamSession:
		result := `<iq id='%s' type='result'/>`
		fmt.Fprintf(sc.connection, result, iq.Id)
	}
}
