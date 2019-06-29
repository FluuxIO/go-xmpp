package stanza

import (
	"encoding/xml"
)

// ============================================================================
// StreamFeatures Packet
// Reference: The active stream features are published on
//            https://xmpp.org/registrar/stream-features.html
// Note: That page misses draft and experimental XEP (i.e CSI, etc)

type StreamFeatures struct {
	XMLName xml.Name `xml:"http://etherx.jabber.org/streams features"`
	// Server capabilities hash
	Caps Caps
	// Stream features
	StartTLS         tlsStartTLS
	Mechanisms       saslMechanisms
	Bind             Bind
	StreamManagement streamManagement
	// Obsolete
	Session StreamSession
	// ProcessOne Stream Features
	P1Push   p1Push
	P1Rebind p1Rebind
	p1Ack    p1Ack
	Any      []xml.Name `xml:",any"`
}

func (StreamFeatures) Name() string {
	return "stream:features"
}

type streamFeatureDecoder struct{}

var streamFeatures streamFeatureDecoder

func (streamFeatureDecoder) decode(p *xml.Decoder, se xml.StartElement) (StreamFeatures, error) {
	var packet StreamFeatures
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

// Capabilities
// Reference: https://xmpp.org/extensions/xep-0115.html#stream
//    "A server MAY include its entity capabilities in a stream feature element so that connecting clients
//     and peer servers do not need to send service discovery requests each time they connect."
// This is not a stream feature but a way to let client cache server disco info.
type Caps struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/caps c"`
	Hash    string   `xml:"hash,attr"`
	Node    string   `xml:"node,attr"`
	Ver     string   `xml:"ver,attr"`
	Ext     string   `xml:"ext,attr,omitempty"`
}

// ============================================================================
// Supported Stream Features

// StartTLS feature
// Reference: RFC 6120 - https://tools.ietf.org/html/rfc6120#section-5.4
type tlsStartTLS struct {
	XMLName  xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls starttls"`
	Required bool
}

// UnmarshalXML implements custom parsing startTLS required flag
func (stls *tlsStartTLS) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	stls.XMLName = start.Name

	// Check subelements to extract required field as boolean
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch tt := t.(type) {

		case xml.StartElement:
			elt := new(Node)

			err = d.DecodeElement(elt, &tt)
			if err != nil {
				return err
			}

			if elt.XMLName.Local == "required" {
				stls.Required = true
			}

		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}
	}
}

func (sf *StreamFeatures) DoesStartTLS() (feature tlsStartTLS, isSupported bool) {
	if sf.StartTLS.XMLName.Space+" "+sf.StartTLS.XMLName.Local == nsTLS+" starttls" {
		return sf.StartTLS, true
	}
	return feature, false
}

// Mechanisms
// Reference: RFC 6120 - https://tools.ietf.org/html/rfc6120#section-6.4.1
type saslMechanisms struct {
	XMLName   xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-sasl mechanisms"`
	Mechanism []string `xml:"mechanism"`
}

// StreamManagement
// Reference: XEP-0198 - https://xmpp.org/extensions/xep-0198.html#feature
type streamManagement struct {
	XMLName xml.Name `xml:"urn:xmpp:sm:3 sm"`
}

func (sf *StreamFeatures) DoesStreamManagement() (isSupported bool) {
	if sf.StreamManagement.XMLName.Space+" "+sf.StreamManagement.XMLName.Local == "urn:xmpp:sm:3 sm" {
		return true
	}
	return false
}

// P1 extensions
// Reference: https://docs.ejabberd.im/developer/mobile/core-features/

// p1:push support
type p1Push struct {
	XMLName xml.Name `xml:"p1:push push"`
}

// p1:rebind suppor
type p1Rebind struct {
	XMLName xml.Name `xml:"p1:rebind rebind"`
}

// p1:ack support
type p1Ack struct {
	XMLName xml.Name `xml:"p1:ack ack"`
}

// ============================================================================
// StreamError Packet

type StreamError struct {
	XMLName xml.Name `xml:"http://etherx.jabber.org/streams error"`
	Error   xml.Name `xml:",any"`
	Text    string   `xml:"urn:ietf:params:xml:ns:xmpp-streams text"`
}

func (StreamError) Name() string {
	return "stream:error"
}

type streamErrorDecoder struct{}

var streamError streamErrorDecoder

func (streamErrorDecoder) decode(p *xml.Decoder, se xml.StartElement) (StreamError, error) {
	var packet StreamError
	err := p.DecodeElement(&packet, &se)
	return packet, err
}
