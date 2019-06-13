package xmpp_test

import (
	"bytes"
	"encoding/xml"
	"io"
	"testing"

	"gosrc.io/xmpp"
)

var successFlag = []byte("matched")

func TestNameMatcher(t *testing.T) {
	router := xmpp.NewRouter()
	router.HandleFunc("message", func(w io.Writer, p xmpp.Packet) {
		_, _ = w.Write(successFlag)
	})

	// Check that a message packet is properly matched
	var buf bytes.Buffer
	// TODO: We want packet creation code to use struct to use default values
	msg := xmpp.NewMessage("chat", "", "test@localhost", "1", "")
	msg.Body = "Hello"
	router.Route(&buf, msg)
	if !bytes.Equal(buf.Bytes(), successFlag) {
		t.Error("Message was not matched and routed properly")
	}

	// Check that an IQ packet is not matched
	buf = bytes.Buffer{}
	iq := xmpp.NewIQ("get", "", "localhost", "1", "")
	iq.Payload = append(iq.Payload, &xmpp.DiscoInfo{})
	router.Route(&buf, iq)
	if bytes.Equal(buf.Bytes(), successFlag) {
		t.Error("IQ should not have been matched and routed")
	}
}

func TestIQNSMatcher(t *testing.T) {
	router := xmpp.NewRouter()
	router.NewRoute().
		IQNamespaces(xmpp.NSDiscoInfo, xmpp.NSDiscoItems).
		HandlerFunc(func(w io.Writer, p xmpp.Packet) {
			_, _ = w.Write(successFlag)
		})

	// Check that an IQ with proper namespace does match
	var buf bytes.Buffer
	iqDisco := xmpp.NewIQ("get", "", "localhost", "1", "")
	// TODO: Add a function to generate payload with proper namespace initialisation
	iqDisco.Payload = append(iqDisco.Payload, &xmpp.DiscoInfo{
		XMLName: xml.Name{
			Space: xmpp.NSDiscoInfo,
			Local: "query",
		}})
	router.Route(&buf, iqDisco)
	if !bytes.Equal(buf.Bytes(), successFlag) {
		t.Errorf("IQ should have been matched and routed: %v", iqDisco)
	}

	// Check that another namespace is not matched
	buf = bytes.Buffer{}
	iqVersion := xmpp.NewIQ("get", "", "localhost", "1", "")
	// TODO: Add a function to generate payload with proper namespace initialisation
	iqVersion.Payload = append(iqVersion.Payload, &xmpp.DiscoInfo{
		XMLName: xml.Name{
			Space: "jabber:iq:version",
			Local: "query",
		}})
	router.Route(&buf, iqVersion)
	if bytes.Equal(buf.Bytes(), successFlag) {
		t.Errorf("IQ should not have been matched and routed: %v", iqVersion)
	}
}
