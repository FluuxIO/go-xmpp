package xmpp // import "gosrc.io/xmpp"

import (
	"encoding/xml"
)

// ============================================================================
// StreamFeatures Packet
// Reference: https://xmpp.org/registrar/stream-features.html

type StreamFeatures struct {
	XMLName xml.Name `xml:"http://etherx.jabber.org/streams features"`
	// Server capabilities hash
	Caps Caps
	// Stream features
	StartTLS   tlsStartTLS
	Mechanisms saslMechanisms
	Bind       BindBind
	Session    sessionSession
	Any        []xml.Name `xml:",any"`
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
