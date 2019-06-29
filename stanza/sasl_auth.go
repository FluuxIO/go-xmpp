package stanza

import "encoding/xml"

// ============================================================================

// SASLAuth implements SASL Authentication initiation.
// Reference: https://tools.ietf.org/html/rfc6120#section-6.4.2
type SASLAuth struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl auth"`
	Mechanism string   `xml:"mechanism,attr"`
	Value     string   `xml:",innerxml"`
}

// ============================================================================

// SASLSuccess implements SASL Success nonza, sent by server as a result of the
// SASL auth negotiation.
// Reference: https://tools.ietf.org/html/rfc6120#section-6.4.6
type SASLSuccess struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl success"`
}

func (SASLSuccess) Name() string {
	return "sasl:success"
}

// SASLSuccess decoding
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

// SASLFailure decoding
type saslFailureDecoder struct{}

var saslFailure saslFailureDecoder

func (saslFailureDecoder) decode(p *xml.Decoder, se xml.StartElement) (SASLFailure, error) {
	var packet SASLFailure
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

// ===========================================================================
// Resource binding

// Bind is an IQ payload used during session negotiation to bind user resource
// to the current XMPP stream.
// Reference: https://tools.ietf.org/html/rfc6120#section-7
type Bind struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-bind bind"`
	Resource string   `xml:"resource,omitempty"`
	Jid      string   `xml:"jid,omitempty"`
}

func (b *Bind) Namespace() string {
	return b.XMLName.Space
}

// ============================================================================
// Session (Obsolete)

// Session is both a stream feature and an obsolete IQ Payload, used to bind a
// resource to the current XMPP stream on RFC 3121 only XMPP servers.
// Session is obsolete in RFC 6121. It is added to Fluux XMPP for compliance
// with RFC 3121.
// Reference: https://xmpp.org/rfcs/rfc3921.html#session
//
// This is the draft defining how to handle the transition:
//    https://tools.ietf.org/html/draft-cridland-xmpp-session-01
type StreamSession struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-session session"`
	Optional bool     // If element does exist, it mean we are not required to open session
}

func (s *StreamSession) Namespace() string {
	return s.XMLName.Space
}

func (s *StreamSession) IsOptional() bool {
	if s.XMLName.Local == "session" {
		return s.Optional
	}
	// If session element is missing, then we should not use session
	return true
}

// ============================================================================
// Registry init

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{"urn:ietf:params:xml:ns:xmpp-bind", "bind"}, Bind{})
	TypeRegistry.MapExtension(PKTIQ, xml.Name{"urn:ietf:params:xml:ns:xmpp-session", "session"}, StreamSession{})
}
