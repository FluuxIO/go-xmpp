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
	Value string `xml:",innerxml"`
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
	// Result sets
	ResultSet *ResultSet `xml:"set,omitempty"`
}

func (d *Delegation) Namespace() string {
	return d.XMLName.Space
}
func (d *Delegation) GetSet() *ResultSet {
	return d.ResultSet
}

type Delegated struct {
	XMLName   xml.Name `xml:"delegated"`
	Namespace string   `xml:"namespace,attr,omitempty"`
}

func init() {
	TypeRegistry.MapExtension(PKTMessage, xml.Name{Space: "urn:xmpp:delegation:1", Local: "delegation"}, Delegation{})
	TypeRegistry.MapExtension(PKTIQ, xml.Name{Space: "urn:xmpp:delegation:1", Local: "delegation"}, Delegation{})
}
