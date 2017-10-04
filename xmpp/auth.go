package xmpp

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

func authSASL(socket io.ReadWriter, decoder *xml.Decoder, f streamFeatures, user string, password string) (err error) {
	// TODO: Implement other type of SASL Authentication
	havePlain := false
	for _, m := range f.Mechanisms.Mechanism {
		if m == "PLAIN" {
			havePlain = true
			break
		}
	}
	if !havePlain {
		return fmt.Errorf("PLAIN authentication is not supported by server: %v", f.Mechanisms.Mechanism)
	}

	return authPlain(socket, decoder, user, password)
}

// Plain authentication: send base64-encoded \x00 user \x00 password
func authPlain(socket io.ReadWriter, decoder *xml.Decoder, user string, password string) error {
	raw := "\x00" + user + "\x00" + password
	enc := make([]byte, base64.StdEncoding.EncodedLen(len(raw)))
	base64.StdEncoding.Encode(enc, []byte(raw))
	fmt.Fprintf(socket, "<auth xmlns='%s' mechanism='PLAIN'>%s</auth>", nsSASL, enc)

	// Next message should be either success or failure.
	name, val, err := next(decoder)
	if err != nil {
		return err
	}

	switch v := val.(type) {
	case *saslSuccess:
	case *saslFailure:
		// v.Any is type of sub-element in failure, which gives a description of what failed.
		return errors.New("auth failure: " + v.Any.Local)
	default:
		return errors.New("expected success or failure, got " + name.Local + " in " + name.Space)
	}
	return err
}

// XMPP Packet Parsing
type saslMechanisms struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl mechanisms"`
	Mechanism []string `xml:"mechanism"`
}

type saslSuccess struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl success"`
}

type saslFailure struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl failure"`
	Any     xml.Name // error reason is a subelement
}

type auth struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl auth"`
	Mechanism string   `xml:"mecanism,attr"`
	Value     string   `xml:",innerxml"`
}

type bindBind struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-bind bind"`
	Resource string   `xml:"resource,omitempty"`
	Jid      string   `xml:"jid,omitempty"`
}

func (*bindBind) IsIQPayload() {
}

// Session is obsolete in RFC 6121.
// Added for compliance with RFC 3121.
// Remove when ejabberd purely conforms to RFC 6121.
type sessionSession struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-session session"`
	optional xml.Name // If it does exist, it mean we are not required to open session
}
