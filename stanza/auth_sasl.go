package stanza

import "encoding/xml"

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

type Auth struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl auth"`
	Mechanism string   `xml:"mecanism,attr"`
	Value     string   `xml:",innerxml"`
}

type BindBind struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-bind bind"`
	Resource string   `xml:"resource,omitempty"`
	Jid      string   `xml:"jid,omitempty"`
}

func (b *BindBind) Namespace() string {
	return b.XMLName.Space
}

// Session is obsolete in RFC 6121.
// Added for compliance with RFC 3121.
// Remove when ejabberd purely conforms to RFC 6121.
type sessionSession struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-session session"`
	Optional xml.Name // If it does exist, it mean we are not required to open session
}
