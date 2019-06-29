package stanza

import (
	"encoding/xml"
)

// ============================================================================
// Handshake Stanza

// Handshake is a stanza used by XMPP components to authenticate on XMPP
// component port.
type Handshake struct {
	XMLName xml.Name `xml:"jabber:component:accept handshake"`
	// TODO Add handshake value with test for proper serialization
	// Value string     `xml:",innerxml"`
}

func (Handshake) Name() string {
	return "component:handshake"
}

// Handshake decoding wrapper

type handshakeDecoder struct{}

var handshake handshakeDecoder

func (handshakeDecoder) decode(p *xml.Decoder, se xml.StartElement) (Handshake, error) {
	var packet Handshake
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

// ============================================================================
// Component delegation
// XEP-0355

// Delegation can be used both on message (for delegated) and IQ (for Forwarded),
// depending on the context.
type Delegation struct {
	MsgExtension
	XMLName   xml.Name   `xml:"urn:xmpp:delegation:1 delegation"`
	Forwarded *Forwarded // This is used in iq to wrap delegated iqs
	Delegated *Delegated // This is used in a message to confirm delegated namespace
}

func (d *Delegation) Namespace() string {
	return d.XMLName.Space
}

// Forwarded is used to wrapped forwarded stanzas.
// TODO: Move it in another file, as it is not limited to components.
type Forwarded struct {
	XMLName xml.Name `xml:"urn:xmpp:forward:0 forwarded"`
	Stanza  Packet
}

// UnmarshalXML is a custom unmarshal function used by xml.Unmarshal to
// transform generic XML content into hierarchical Node structure.
func (f *Forwarded) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Check subelements to extract required field as boolean
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch tt := t.(type) {

		case xml.StartElement:
			if packet, err := decodeClient(d, tt); err == nil {
				f.Stanza = packet
			}

		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}
	}
}

type Delegated struct {
	XMLName   xml.Name `xml:"delegated"`
	Namespace string   `xml:"namespace,attr,omitempty"`
}

func init() {
	TypeRegistry.MapExtension(PKTMessage, xml.Name{"urn:xmpp:delegation:1", "delegation"}, Delegation{})
	TypeRegistry.MapExtension(PKTIQ, xml.Name{"urn:xmpp:delegation:1", "delegation"}, Delegation{})
}
