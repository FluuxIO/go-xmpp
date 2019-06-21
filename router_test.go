package xmpp_test

import (
	"bytes"
	"encoding/xml"
	"testing"

	"gosrc.io/xmpp"
)

// ============================================================================
// Test route & matchers

func TestNameMatcher(t *testing.T) {
	router := xmpp.NewRouter()
	router.HandleFunc("message", func(s xmpp.Sender, p xmpp.Packet) {
		_ = s.SendRaw(successFlag)
	})

	// Check that a message packet is properly matched
	conn := NewSenderMock()
	// TODO: We want packet creation code to use struct to use default values
	msg := xmpp.NewMessage("chat", "", "test@localhost", "1", "")
	msg.Body = "Hello"
	router.Route(conn, msg)
	if conn.String() != successFlag {
		t.Error("Message was not matched and routed properly")
	}

	// Check that an IQ packet is not matched
	conn = NewSenderMock()
	iq := xmpp.NewIQ("get", "", "localhost", "1", "")
	iq.Payload = &xmpp.DiscoInfo{}
	router.Route(conn, iq)
	if conn.String() == successFlag {
		t.Error("IQ should not have been matched and routed")
	}
}

func TestIQNSMatcher(t *testing.T) {
	router := xmpp.NewRouter()
	router.NewRoute().
		IQNamespaces(xmpp.NSDiscoInfo, xmpp.NSDiscoItems).
		HandlerFunc(func(s xmpp.Sender, p xmpp.Packet) {
			_ = s.SendRaw(successFlag)
		})

	// Check that an IQ with proper namespace does match
	conn := NewSenderMock()
	iqDisco := xmpp.NewIQ("get", "", "localhost", "1", "")
	// TODO: Add a function to generate payload with proper namespace initialisation
	iqDisco.Payload = &xmpp.DiscoInfo{
		XMLName: xml.Name{
			Space: xmpp.NSDiscoInfo,
			Local: "query",
		}}
	router.Route(conn, iqDisco)
	if conn.String() != successFlag {
		t.Errorf("IQ should have been matched and routed: %v", iqDisco)
	}

	// Check that another namespace is not matched
	conn = NewSenderMock()
	iqVersion := xmpp.NewIQ("get", "", "localhost", "1", "")
	// TODO: Add a function to generate payload with proper namespace initialisation
	iqVersion.Payload = &xmpp.DiscoInfo{
		XMLName: xml.Name{
			Space: "jabber:iq:version",
			Local: "query",
		}}
	router.Route(conn, iqVersion)
	if conn.String() == successFlag {
		t.Errorf("IQ should not have been matched and routed: %v", iqVersion)
	}
}

func TestTypeMatcher(t *testing.T) {
	router := xmpp.NewRouter()
	router.NewRoute().
		StanzaType("normal").
		HandlerFunc(func(s xmpp.Sender, p xmpp.Packet) {
			_ = s.SendRaw(successFlag)
		})

	// Check that a packet with the proper type matches
	conn := NewSenderMock()
	message := xmpp.NewMessage("normal", "", "test@localhost", "1", "")
	message.Body = "hello"
	router.Route(conn, message)

	if conn.String() != successFlag {
		t.Errorf("'normal' message should have been matched and routed: %v", message)
	}

	// We should match on default type 'normal' for message without a type
	conn = NewSenderMock()
	message = xmpp.NewMessage("", "", "test@localhost", "1", "")
	message.Body = "hello"
	router.Route(conn, message)

	if conn.String() != successFlag {
		t.Errorf("message should have been matched and routed: %v", message)
	}

	// We do not match on other types
	conn = NewSenderMock()
	iqVersion := xmpp.NewIQ("get", "service.localhost", "test@localhost", "1", "")
	iqVersion.Payload = &xmpp.DiscoInfo{
		XMLName: xml.Name{
			Space: "jabber:iq:version",
			Local: "query",
		}}
	router.Route(conn, iqVersion)

	if conn.String() == successFlag {
		t.Errorf("iq get should not have been matched and routed: %v", iqVersion)
	}
}

// A blank route with empty matcher will always match
// It can be use to receive all packets that do not match any of the previous route.
func TestCatchallMatcher(t *testing.T) {
	router := xmpp.NewRouter()
	router.NewRoute().
		HandlerFunc(func(s xmpp.Sender, p xmpp.Packet) {
			_ = s.SendRaw(successFlag)
		})

	// Check that we match on several packets
	conn := NewSenderMock()
	message := xmpp.NewMessage("chat", "", "test@localhost", "1", "")
	message.Body = "hello"
	router.Route(conn, message)

	if conn.String() != successFlag {
		t.Errorf("chat message should have been matched and routed: %v", message)
	}

	conn = NewSenderMock()
	iqVersion := xmpp.NewIQ("get", "service.localhost", "test@localhost", "1", "")
	iqVersion.Payload = &xmpp.DiscoInfo{
		XMLName: xml.Name{
			Space: "jabber:iq:version",
			Local: "query",
		}}
	router.Route(conn, iqVersion)

	if conn.String() != successFlag {
		t.Errorf("iq get should have been matched and routed: %v", iqVersion)
	}
}

// ============================================================================
// SenderMock

var successFlag = "matched"

type SenderMock struct {
	buffer *bytes.Buffer
}

func NewSenderMock() SenderMock {
	return SenderMock{buffer: new(bytes.Buffer)}
}

func (s SenderMock) Send(packet xmpp.Packet) error {
	out, err := xml.Marshal(packet)
	if err != nil {
		return err
	}
	s.buffer.Write(out)
	return nil
}

func (s SenderMock) SendRaw(str string) error {
	s.buffer.WriteString(str)
	return nil
}

func (s SenderMock) String() string {
	return s.buffer.String()
}

func TestSenderMock(t *testing.T) {
	conn := NewSenderMock()
	msg := xmpp.NewMessage("", "", "test@localhost", "1", "")
	msg.Body = "Hello"
	if err := conn.Send(msg); err != nil {
		t.Error("Could not send message")
	}
	if conn.String() != "<message id=\"1\" to=\"test@localhost\"><body>Hello</body></message>" {
		t.Errorf("Incorrect packet sent: %s", conn.String())
	}
}
