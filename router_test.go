package xmpp

import (
	"bytes"
	"context"
	"encoding/xml"
	"runtime"
	"testing"

	"gosrc.io/xmpp/stanza"
)

// ============================================================================
// Test route & matchers

func TestIQResultRoutes(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	if router.IQResultRoutes == nil {
		t.Fatal("NewRouter does not initialize isResultRoutes")
	}

	// Check other IQ does not matcah
	conn := NewSenderMock()
	iq := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeResult, Id: "4321"})
	router.NewIQResultRoute(context.Background(), "1234").HandlerFunc(func(ctx context.Context, s Sender, iq stanza.IQ) {
		_ = s.SendRaw(successFlag)
	})
	if conn.String() == successFlag {
		t.Fatal("IQ result with wrong ID was matched")
	}

	// Check if the IQ handler was called
	conn = NewSenderMock()
	iq = stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeResult, Id: "1234"})
	router.route(conn, iq)
	if conn.String() != successFlag {
		t.Fatal("IQ result was not matched")
	}

	// The match must only happen once, so we if receive the same package again it
	// must not be matched.
	conn = NewSenderMock()
	router.route(conn, iq)
	if conn.String() == successFlag {
		t.Fatal("IQ result was matched twice")
	}

	// After cancelling a route it should no longer match
	conn = NewSenderMock()
	ctx, cancel := context.WithCancel(context.Background())
	iq = stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeResult, Id: "1234"})
	router.NewIQResultRoute(ctx, "1234").HandlerFunc(func(ctx context.Context, s Sender, iq stanza.IQ) {
		_ = s.SendRaw(successFlag)
	}).TimeoutHandlerFunc(func(err error) {
		conn.SendRaw(cancelledFlag)
	})
	cancel()
	// Yield the processor so the cancellation goroutine is triggered
	runtime.Gosched()
	router.route(conn, iq)
	if conn.String() != cancelledFlag {
		t.Fatal("IQ result route was matched after cancellation")
	}
}

func TestNameMatcher(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("message", func(s Sender, p stanza.Packet) {
		_ = s.SendRaw(successFlag)
	})

	// Check that a message packet is properly matched
	conn := NewSenderMock()
	msg := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, To: "test@localhost", Id: "1"})
	msg.Body = "Hello"
	router.route(conn, msg)
	if conn.String() != successFlag {
		t.Error("Message was not matched and routed properly")
	}

	// Check that an IQ packet is not matched
	conn = NewSenderMock()
	iq := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, To: "localhost", Id: "1"})
	iq.Payload = &stanza.DiscoInfo{}
	router.route(conn, iq)
	if conn.String() == successFlag {
		t.Error("IQ should not have been matched and routed")
	}
}

func TestIQNSMatcher(t *testing.T) {
	router := NewRouter()
	router.NewRoute().
		IQNamespaces(stanza.NSDiscoInfo, stanza.NSDiscoItems).
		HandlerFunc(func(s Sender, p stanza.Packet) {
			_ = s.SendRaw(successFlag)
		})

	// Check that an IQ with proper namespace does match
	conn := NewSenderMock()
	iqDisco := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, To: "localhost", Id: "1"})
	// TODO: Add a function to generate payload with proper namespace initialisation
	iqDisco.Payload = &stanza.DiscoInfo{
		XMLName: xml.Name{
			Space: stanza.NSDiscoInfo,
			Local: "query",
		}}
	router.route(conn, iqDisco)
	if conn.String() != successFlag {
		t.Errorf("IQ should have been matched and routed: %v", iqDisco)
	}

	// Check that another namespace is not matched
	conn = NewSenderMock()
	iqVersion := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, To: "localhost", Id: "1"})
	// TODO: Add a function to generate payload with proper namespace initialisation
	iqVersion.Payload = &stanza.DiscoInfo{
		XMLName: xml.Name{
			Space: "jabber:iq:version",
			Local: "query",
		}}
	router.route(conn, iqVersion)
	if conn.String() == successFlag {
		t.Errorf("IQ should not have been matched and routed: %v", iqVersion)
	}
}

func TestTypeMatcher(t *testing.T) {
	router := NewRouter()
	router.NewRoute().
		StanzaType("normal").
		HandlerFunc(func(s Sender, p stanza.Packet) {
			_ = s.SendRaw(successFlag)
		})

	// Check that a packet with the proper type matches
	conn := NewSenderMock()
	message := stanza.NewMessage(stanza.Attrs{Type: "normal", To: "test@localhost", Id: "1"})
	message.Body = "hello"
	router.route(conn, message)

	if conn.String() != successFlag {
		t.Errorf("'normal' message should have been matched and routed: %v", message)
	}

	// We should match on default type 'normal' for message without a type
	conn = NewSenderMock()
	message = stanza.NewMessage(stanza.Attrs{To: "test@localhost", Id: "1"})
	message.Body = "hello"
	router.route(conn, message)

	if conn.String() != successFlag {
		t.Errorf("message should have been matched and routed: %v", message)
	}

	// We do not match on other types
	conn = NewSenderMock()
	iqVersion := stanza.NewIQ(stanza.Attrs{Type: "get", From: "service.localhost", To: "test@localhost", Id: "1"})
	iqVersion.Payload = &stanza.DiscoInfo{
		XMLName: xml.Name{
			Space: "jabber:iq:version",
			Local: "query",
		}}
	router.route(conn, iqVersion)

	if conn.String() == successFlag {
		t.Errorf("iq get should not have been matched and routed: %v", iqVersion)
	}
}

func TestCompositeMatcher(t *testing.T) {
	router := NewRouter()
	router.NewRoute().
		IQNamespaces("jabber:iq:version").
		StanzaType("get").
		HandlerFunc(func(s Sender, p stanza.Packet) {
			_ = s.SendRaw(successFlag)
		})

	// Data set
	getVersionIq := stanza.NewIQ(stanza.Attrs{Type: "get", From: "service.localhost", To: "test@localhost", Id: "1"})
	getVersionIq.Payload = &stanza.Version{
		XMLName: xml.Name{
			Space: "jabber:iq:version",
			Local: "query",
		}}

	setVersionIq := stanza.NewIQ(stanza.Attrs{Type: "set", From: "service.localhost", To: "test@localhost", Id: "1"})
	setVersionIq.Payload = &stanza.Version{
		XMLName: xml.Name{
			Space: "jabber:iq:version",
			Local: "query",
		}}

	GetDiscoIq := stanza.NewIQ(stanza.Attrs{Type: "get", From: "service.localhost", To: "test@localhost", Id: "1"})
	GetDiscoIq.Payload = &stanza.DiscoInfo{
		XMLName: xml.Name{
			Space: "http://jabber.org/protocol/disco#info",
			Local: "query",
		}}

	message := stanza.NewMessage(stanza.Attrs{Type: "normal", To: "test@localhost", Id: "1"})
	message.Body = "hello"

	tests := []struct {
		name  string
		input stanza.Packet
		want  bool
	}{
		{name: "match get version iq", input: getVersionIq, want: true},
		{name: "ignore set version iq", input: setVersionIq, want: false},
		{name: "ignore get discoinfo iq", input: GetDiscoIq, want: false},
		{name: "ignore message", input: message, want: false},
	}

	//
	for _, tc := range tests {
		t.Run(tc.name, func(st *testing.T) {
			conn := NewSenderMock()
			router.route(conn, tc.input)

			res := conn.String() == successFlag
			if tc.want != res {
				st.Errorf("incorrect result for %#v\nMatch = %#v, expecting %#v", tc.input, res, tc.want)
			}
		})
	}
}

// A blank route with empty matcher will always match
// It can be use to receive all packets that do not match any of the previous route.
func TestCatchallMatcher(t *testing.T) {
	router := NewRouter()
	router.NewRoute().
		HandlerFunc(func(s Sender, p stanza.Packet) {
			_ = s.SendRaw(successFlag)
		})

	// Check that we match on several packets
	conn := NewSenderMock()
	message := stanza.NewMessage(stanza.Attrs{Type: "chat", To: "test@localhost", Id: "1"})
	message.Body = "hello"
	router.route(conn, message)

	if conn.String() != successFlag {
		t.Errorf("chat message should have been matched and routed: %v", message)
	}

	conn = NewSenderMock()
	iqVersion := stanza.NewIQ(stanza.Attrs{Type: "get", From: "service.localhost", To: "test@localhost", Id: "1"})
	iqVersion.Payload = &stanza.DiscoInfo{
		XMLName: xml.Name{
			Space: "jabber:iq:version",
			Local: "query",
		}}
	router.route(conn, iqVersion)

	if conn.String() != successFlag {
		t.Errorf("iq get should have been matched and routed: %v", iqVersion)
	}
}

// ============================================================================
// SenderMock

const successFlag = "matched"
const cancelledFlag = "cancelled"

type SenderMock struct {
	buffer *bytes.Buffer
}

func NewSenderMock() SenderMock {
	return SenderMock{buffer: new(bytes.Buffer)}
}

func (s SenderMock) Send(packet stanza.Packet) error {
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
	msg := stanza.NewMessage(stanza.Attrs{To: "test@localhost", Id: "1"})
	msg.Body = "Hello"
	if err := conn.Send(msg); err != nil {
		t.Error("Could not send message")
	}
	if conn.String() != "<message id=\"1\" to=\"test@localhost\"><body>Hello</body></message>" {
		t.Errorf("Incorrect packet sent: %s", conn.String())
	}
}
