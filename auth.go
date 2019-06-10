package xmpp // import "gosrc.io/xmpp"

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

func authSASL(socket io.ReadWriter, decoder *xml.Decoder, f StreamFeatures, user string, password string) (err error) {
	// TODO: Implement other type of SASL Authentication
	havePlain := false
	for _, m := range f.Mechanisms.Mechanism {
		if m == "PLAIN" {
			havePlain = true
			break
		}
	}
	if !havePlain {
		err := fmt.Errorf("PLAIN authentication is not supported by server: %v", f.Mechanisms.Mechanism)
		return NewConnError(err, true)
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
	val, err := next(decoder)
	if err != nil {
		return err
	}

	switch v := val.(type) {
	case SASLSuccess:
	case SASLFailure:
		// v.Any is type of sub-element in failure, which gives a description of what failed.
		err := errors.New("auth failure: " + v.Any.Local)
		return NewConnError(err, true)
	default:
		return errors.New("expected SASL success or failure, got " + v.Name())
	}
	return err
}

// ============================================================================
// SASLSuccess

type SASLSuccess struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl success"`
}

func (SASLSuccess) Name() string {
	return "sasl:success"
}

type saslSuccessDecoder struct{}

var saslSuccess saslSuccessDecoder

func (saslSuccessDecoder) decode(p *xml.Decoder, se xml.StartElement) (SASLSuccess, error) {
	var packet SASLSuccess
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

// ============================================================================
// SASLFailure

type SASLFailure struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl failure"`
	Any     xml.Name // error reason is a subelement
}

func (SASLFailure) Name() string {
	return "sasl:failure"
}

type saslFailureDecoder struct{}

var saslFailure saslFailureDecoder

func (saslFailureDecoder) decode(p *xml.Decoder, se xml.StartElement) (SASLFailure, error) {
	var packet SASLFailure
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

// ============================================================================

type auth struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl auth"`
	Mechanism string   `xml:"mecanism,attr"`
	Value     string   `xml:",innerxml"`
}

type BindBind struct {
	IQPayload
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-bind bind"`
	Resource string   `xml:"resource,omitempty"`
	Jid      string   `xml:"jid,omitempty"`
}

// Session is obsolete in RFC 6121.
// Added for compliance with RFC 3121.
// Remove when ejabberd purely conforms to RFC 6121.
type sessionSession struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-session session"`
	optional xml.Name // If it does exist, it mean we are not required to open session
}
