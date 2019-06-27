package stanza

import (
	"encoding/xml"
)

// ============================================================================
// Disco Info

const (
	NSDiscoInfo = "http://jabber.org/protocol/disco#info"
)

// ----------
// Namespaces

type DiscoInfo struct {
	XMLName  xml.Name   `xml:"http://jabber.org/protocol/disco#info query"`
	Node     string     `xml:"node,attr,omitempty"`
	Identity []Identity `xml:"identity"`
	Features []Feature  `xml:"feature"`
}

func (d *DiscoInfo) Namespace() string {
	return d.XMLName.Space
}

// ---------------
// Builder helpers

// DiscoInfo builds a default DiscoInfo payload
func (iq *IQ) DiscoInfo() *DiscoInfo {
	d := DiscoInfo{
		XMLName: xml.Name{
			Space: NSDiscoInfo,
			Local: "query",
		},
	}
	iq.Payload = &d
	return &d
}

func (d *DiscoInfo) AddIdentity(name, category, typ string) {
	identity := Identity{
		XMLName:  xml.Name{Local: "identity"},
		Name:     name,
		Category: category,
		Type:     typ,
	}
	d.Identity = append(d.Identity, identity)
}

func (d *DiscoInfo) AddFeatures(namespace ...string) {
	for _, ns := range namespace {
		d.Features = append(d.Features, Feature{Var: ns})
	}
}

func (d *DiscoInfo) SetNode(node string) {
	d.Node = node
}

func (d *DiscoInfo) SetIdentities(ident ...Identity) *DiscoInfo {
	d.Identity = ident
	return d
}

func (d *DiscoInfo) SetFeatures(namespace ...string) *DiscoInfo {
	for _, ns := range namespace {
		d.Features = append(d.Features, Feature{Var: ns})
	}
	return d
}

// -----------
// SubElements

type Identity struct {
	XMLName  xml.Name `xml:"identity,omitempty"`
	Name     string   `xml:"name,attr,omitempty"`
	Category string   `xml:"category,attr,omitempty"`
	Type     string   `xml:"type,attr,omitempty"`
}

type Feature struct {
	XMLName xml.Name `xml:"feature"`
	Var     string   `xml:"var,attr"`
}

// ============================================================================
// Disco Info

const (
	NSDiscoItems = "http://jabber.org/protocol/disco#items"
)

type DiscoItems struct {
	XMLName xml.Name    `xml:"http://jabber.org/protocol/disco#items query"`
	Node    string      `xml:"node,attr,omitempty"`
	Items   []DiscoItem `xml:"item"`
}

func (d *DiscoItems) Namespace() string {
	return d.XMLName.Space
}

type DiscoItem struct {
	XMLName xml.Name `xml:"item"`
	Name    string   `xml:"name,attr,omitempty"`
	JID     string   `xml:"jid,attr,omitempty"`
	Node    string   `xml:"node,attr,omitempty"`
}

// ============================================================================
// Registry init

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{NSDiscoInfo, "query"}, DiscoInfo{})
	TypeRegistry.MapExtension(PKTIQ, xml.Name{NSDiscoItems, "query"}, DiscoItems{})
}
