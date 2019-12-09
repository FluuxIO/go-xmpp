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
)

// ClientHandler is passed by the test client to provide custom behaviour to
// the TCP server mock. This allows customizing the server behaviour to allow
// testing clients under various scenarii.
type ClientHandler func(t *testing.T, conn net.Conn)

// ServerMock is a simple TCP server that can be use to mock basic server
// behaviour to test clients.
type ServerMock struct {
	t           *testing.T
	handler     ClientHandler
	listener    net.Listener
	connections []net.Conn
	done        chan struct{}
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
	// Close all existing connections
	for _, c := range mock.connections {
		c.Close()
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

// loop accepts connections and creates a go routine per connection.
// The go routine is running the client handler, that is used to provide the
// real TCP server behaviour.
func (mock *ServerMock) loop() {
	listener := mock.listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-mock.done:
				return
			default:
				mock.t.Error("TCPServerMock accept error:", err.Error())
			}
			return
		}
		mock.connections = append(mock.connections, conn)
		// TODO Create and pass a context to cancel the handler if they are still around = avoid possible leak on complex handlers
		go mock.handler(mock.t, conn)
	}
}

//======================================================================================================================
// A few functions commonly used for tests. Trying to avoid duplicates in client and component test files.
//======================================================================================================================

func respondToIQ(t *testing.T, c net.Conn) {
	recvBuf := make([]byte, 1024)
	var iqR stanza.IQ
	_, err := c.Read(recvBuf[:]) // recv data
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			t.Errorf("read timeout: %s", err)
		} else {
			t.Errorf("read error: %s", err)
		}
	}
	xml.Unmarshal(recvBuf, &iqR)

	if !iqR.IsValid() {
		mockIQError(c)
		return
	}

	// Crafting response
	iqResp := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeResult, From: iqR.To, To: iqR.From, Id: iqR.Id, Lang: "en"})
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

// When a presence stanza is automatically sent (right now it's the case in the client), we may want to discard it
// and test further stanzas.
func discardPresence(t *testing.T, c net.Conn) {
	decoder := xml.NewDecoder(c)
	c.SetDeadline(time.Now().Add(defaultTimeout))
	defer c.SetDeadline(time.Time{})
	var presenceStz stanza.Presence
	err := decoder.Decode(&presenceStz)
	if err != nil {
		t.Errorf("Expected presence but this happened : %s", err.Error())
	}
}

// Reads next request coming from the Component. Expecting it to be an IQ request
func receiveIq(c net.Conn, decoder *xml.Decoder) (*stanza.IQ, error) {
	c.SetDeadline(time.Now().Add(defaultTimeout))
	defer c.SetDeadline(time.Time{})
	var iqStz stanza.IQ
	err := decoder.Decode(&iqStz)
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

func sendStreamFeatures(t *testing.T, c net.Conn, _ *xml.Decoder) {
	// This is a basic server, supporting only 1 stream feature: SASL Plain Auth
	features := `<stream:features>
  <mechanisms xmlns="urn:ietf:params:xml:ns:xmpp-sasl">
    <mechanism>PLAIN</mechanism>
  </mechanisms>
</stream:features>`
	if _, err := fmt.Fprintln(c, features); err != nil {
		t.Errorf("cannot send stream feature: %s", err)
	}
}

// TODO return err in case of error reading the auth params
func readAuth(t *testing.T, decoder *xml.Decoder) string {
	se, err := stanza.NextStart(decoder)
	if err != nil {
		t.Errorf("cannot read auth: %s", err)
		return ""
	}

	var nv interface{}
	nv = &stanza.SASLAuth{}
	// Decode element into pointer storage
	if err = decoder.DecodeElement(nv, &se); err != nil {
		t.Errorf("cannot decode auth: %s", err)
		return ""
	}

	switch v := nv.(type) {
	case *stanza.SASLAuth:
		return v.Value
	}
	return ""
}

func sendBindFeature(t *testing.T, c net.Conn, _ *xml.Decoder) {
	// This is a basic server, supporting only 1 stream feature after auth: resource binding
	features := `<stream:features>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/>
</stream:features>`
	if _, err := fmt.Fprintln(c, features); err != nil {
		t.Errorf("cannot send stream feature: %s", err)
	}
}

func sendRFC3921Feature(t *testing.T, c net.Conn, _ *xml.Decoder) {
	// This is a basic server, supporting only 2 features after auth: resource & session binding
	features := `<stream:features>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/>
  <session xmlns='urn:ietf:params:xml:ns:xmpp-session'/>
</stream:features>`
	if _, err := fmt.Fprintln(c, features); err != nil {
		t.Errorf("cannot send stream feature: %s", err)
	}
}

func bind(t *testing.T, c net.Conn, decoder *xml.Decoder) {
	se, err := stanza.NextStart(decoder)
	if err != nil {
		t.Errorf("cannot read bind: %s", err)
		return
	}

	iq := &stanza.IQ{}
	// Decode element into pointer storage
	if err = decoder.DecodeElement(&iq, &se); err != nil {
		t.Errorf("cannot decode bind iq: %s", err)
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
		fmt.Fprintf(c, result, iq.Id, "test@localhost/test") // TODO use real JID
	}
}

func session(t *testing.T, c net.Conn, decoder *xml.Decoder) {
	se, err := stanza.NextStart(decoder)
	if err != nil {
		t.Errorf("cannot read session: %s", err)
		return
	}

	iq := &stanza.IQ{}
	// Decode element into pointer storage
	if err = decoder.DecodeElement(&iq, &se); err != nil {
		t.Errorf("cannot decode session iq: %s", err)
		return
	}

	switch iq.Payload.(type) {
	case *stanza.StreamSession:
		result := `<iq id='%s' type='result'/>`
		fmt.Fprintf(c, result, iq.Id)
	}
}
