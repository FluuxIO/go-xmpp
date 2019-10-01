package xmpp

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"gosrc.io/xmpp/stanza"
)

const (
	// Default port is not standard XMPP port to avoid interfering
	// with local running XMPP server
	testXMPPAddress = "localhost:15222"

	defaultTimeout = 2 * time.Second
)

func TestClient_Connect(t *testing.T) {
	// Setup Mock server
	mock := ServerMock{}
	mock.Start(t, testXMPPAddress, handlerConnectSuccess)

	// Test / Check result
	config := Config{Address: testXMPPAddress, Jid: "test@localhost", Credential: Password("test"), Insecure: true}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(config, router); err != nil {
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
	mock.Start(t, testXMPPAddress, handlerAbortTLS)

	// Test / Check result
	config := Config{Address: testXMPPAddress, Jid: "test@localhost", Credential: Password("test")}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(config, router); err != nil {
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
	mock.Start(t, testXMPPAddress, handlerAbortTLS)

	// Test / Check result
	config := Config{Address: testXMPPAddress, Jid: "test@localhost", Credential: Password("test")}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(config, router); err != nil {
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
	mock.Start(t, testXMPPAddress, handlerConnectWithSession)

	// Test / Check result
	config := Config{Address: testXMPPAddress, Jid: "test@localhost", Credential: Password("test"), Insecure: true}

	var client *Client
	var err error
	router := NewRouter()
	if client, err = NewClient(config, router); err != nil {
		t.Errorf("connect create XMPP client: %s", err)
	}

	if err = client.Connect(); err != nil {
		t.Errorf("XMPP connection failed: %s", err)
	}

	mock.Stop()
}

//=============================================================================
// Basic XMPP Server Mock Handlers.

const serverStreamOpen = "<?xml version='1.0'?><stream:stream to='%s' id='%s' xmlns='%s' xmlns:stream='%s' version='1.0'>"

// Test connection with a basic straightforward workflow
func handlerConnectSuccess(t *testing.T, c net.Conn) {
	decoder := xml.NewDecoder(c)
	checkOpenStream(t, c, decoder)

	sendStreamFeatures(t, c, decoder) // Send initial features
	readAuth(t, decoder)
	fmt.Fprintln(c, "<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>")

	checkOpenStream(t, c, decoder) // Reset stream
	sendBindFeature(t, c, decoder) // Send post auth features
	bind(t, c, decoder)
}

// We expect client will abort on TLS
func handlerAbortTLS(t *testing.T, c net.Conn) {
	decoder := xml.NewDecoder(c)
	checkOpenStream(t, c, decoder)
	sendStreamFeatures(t, c, decoder) // Send initial features
}

// Test connection with mandatory session (RFC-3921)
func handlerConnectWithSession(t *testing.T, c net.Conn) {
	decoder := xml.NewDecoder(c)
	checkOpenStream(t, c, decoder)

	sendStreamFeatures(t, c, decoder) // Send initial features
	readAuth(t, decoder)
	fmt.Fprintln(c, "<success xmlns=\"urn:ietf:params:xml:ns:xmpp-sasl\"/>")

	checkOpenStream(t, c, decoder)    // Reset stream
	sendRFC3921Feature(t, c, decoder) // Send post auth features
	bind(t, c, decoder)
	session(t, c, decoder)
}

func checkOpenStream(t *testing.T, c net.Conn, decoder *xml.Decoder) {
	c.SetDeadline(time.Now().Add(defaultTimeout))
	defer c.SetDeadline(time.Time{})

	for { // TODO clean up. That for loop is not elegant and I prefer bounded recursion.
		var token xml.Token
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
			if _, err := fmt.Fprintf(c, serverStreamOpen, "localhost", "streamid1", stanza.NSClient, stanza.NSStream); err != nil {
				t.Errorf("cannot write server stream open: %s", err)
			}
			return
		}
	}
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
