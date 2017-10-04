package xmpp

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"testing"
)

const (
	// Default port is not standard XMPP port to avoid interfering
	// with local running XMPP server
	testXMPPAddress = "localhost:15222"
)

func TestClient_Connect(t *testing.T) {
	// Setup Mock server
	mock := XMPPServerMock{}
	mock.Start(t, handlerConnackSuccess)

	// Test / Check result
	options := Options{Address: testXMPPAddress, Jid: "test@localhost", Password: "test"}

	var client *Client
	var err error
	if client, err = NewClient(options); err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	var session *Session
	if session, err = client.Connect(); err != nil {
		t.Errorf("XMPP connection failed: %s", err)
	}

	fmt.Println("Stream opened, we have streamID = ", session.StreamId)

	mock.Stop()
}

//=============================================================================
// Basic XMPP Server Mock Handlers.

const serverStreamOpen = "<?xml version='1.0'?><stream:stream to='%s' id='%s' xmlns='%s' xmlns:stream='%s' version='1.0'>"

func handlerConnackSuccess(t *testing.T, c net.Conn) {
	decoder := xml.NewDecoder(c)
	checkOpenStream(t, decoder)

	if _, err := fmt.Fprintf(c, serverStreamOpen, "localhost", "streamid1", NSClient, NSStream); err != nil {
		t.Errorf("cannot write server stream open: %s", err)
	}
	fmt.Println("Sent stream Open")
	sendStreamFeatures(t, c, decoder)
	fmt.Println("Sent stream feature")
	readAuth(t, decoder)
	fmt.Fprintln(c, "<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>")

	checkOpenStream(t, decoder)
	if _, err := fmt.Fprintf(c, serverStreamOpen, "localhost", "streamid1", NSClient, NSStream); err != nil {
		t.Errorf("cannot write server stream open: %s", err)
	}
	sendBindFeature(t, c, decoder)

	bind(t, c, decoder)
	session(t, c, decoder)
}

func checkOpenStream(t *testing.T, decoder *xml.Decoder) {
	for {
		var token xml.Token
		token, err := decoder.Token()
		if err != nil {
			t.Errorf("cannot read next token: %s", err)
		}

		switch elem := token.(type) {
		// Wait for first startElement
		case xml.StartElement:
			if elem.Name.Space != NSStream || elem.Name.Local != "stream" {
				err = errors.New("xmpp: expected <stream> but got <" + elem.Name.Local + "> in " + elem.Name.Space)
			}
			fmt.Printf("Received: %v\n", elem.Name.Local)
			return
		case xml.ProcInst:
			fmt.Printf("Received: %v\n", elem.Inst)
		}
	}
}

func sendStreamFeatures(t *testing.T, c net.Conn, decoder *xml.Decoder) {
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
	se, err := nextStart(decoder)
	if err != nil {
		t.Errorf("cannot read auth: %s", err)
		return ""
	}

	var nv interface{}
	nv = &auth{}
	// Decode element into pointer storage
	if err = decoder.DecodeElement(nv, &se); err != nil {
		fmt.Println(err)
		t.Errorf("cannot decode auth: %s", err)
		return ""
	}

	switch v := nv.(type) {
	case *auth:
		return v.Value
	}
	return ""
}

func sendBindFeature(t *testing.T, c net.Conn, decoder *xml.Decoder) {
	// This is a basic server, supporting only 1 stream feature: SASL Plain Auth
	features := `<stream:features>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'/>
</stream:features>`
	if _, err := fmt.Fprintln(c, features); err != nil {
		t.Errorf("cannot send stream feature: %s", err)
	}
}

func bind(t *testing.T, c net.Conn, decoder *xml.Decoder) {
	se, err := nextStart(decoder)
	if err != nil {
		t.Errorf("cannot read bind: %s", err)
		return
	}

	iq := &ClientIQ{}
	// Decode element into pointer storage
	if err = decoder.DecodeElement(&iq, &se); err != nil {
		fmt.Println(err)
		t.Errorf("cannot decode bind iq: %s", err)
		return
	}

	switch payload := iq.Payload.(type) {
	case *bindBind:
		fmt.Println("JID:", payload.Jid)
	}
	result := `<iq id='%s' type='result'>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'>
  	<jid>%s</jid>
  </bind>
</iq>`
	fmt.Fprintf(c, result, iq.Id, "test@localhost/test") // TODO use real JID
}

func session(t *testing.T, c net.Conn, decoder *xml.Decoder) {

}

type testHandler func(t *testing.T, conn net.Conn)

type XMPPServerMock struct {
	t           *testing.T
	handler     testHandler
	listener    net.Listener
	connections []net.Conn
	done        chan struct{}
}

func (mock *XMPPServerMock) Start(t *testing.T, handler testHandler) {
	mock.t = t
	mock.handler = handler
	if err := mock.init(); err != nil {
		return
	}
	go mock.loop()
}

func (mock *XMPPServerMock) Stop() {
	close(mock.done)
	if mock.listener != nil {
		mock.listener.Close()
	}
	// Close all existing connections
	for _, c := range mock.connections {
		c.Close()
	}
}

func (mock *XMPPServerMock) init() error {
	mock.done = make(chan struct{})

	l, err := net.Listen("tcp", testXMPPAddress)
	if err != nil {
		mock.t.Errorf("TCPServerMock cannot listen on address: %q", testXMPPAddress)
		return err
	}
	mock.listener = l
	return nil
}

func (mock *XMPPServerMock) loop() {
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
