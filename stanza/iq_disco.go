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

func (d *DiscoInfo) SetNode(node string) *DiscoInfo {
	d.Node = node
	return d
}

func (d *DiscoInfo) SetIdentities(ident ...Identity) *DiscoInfo {
	d.Identity = ident
	return d
}

func (d *DiscoInfo) SetFeatures(namespace ...string) *DiscoInfo {
	d.Features = []Feature{}
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

// ---------------
// Builder helpers

// DiscoItems builds a default DiscoItems payload
func (iq *IQ) DiscoItems() *DiscoItems {
	d := DiscoItems{
		XMLName: xml.Name{Space: "http://jabber.org/protocol/disco#items", Local: "query"},
	}
	iq.Payload = &d
	return &d
}

func (d *DiscoItems) SetNode(node string) *DiscoItems {
	d.Node = node
	return d
}

func (d *DiscoItems) AddItem(jid, node, name string) *DiscoItems {
	item := DiscoItem{
		JID:  jid,
		Node: node,
		Name: name,
	}
	d.Items = append(d.Items, item)
	return d
}

type DiscoItem struct {
	XMLName xml.Name `xml:"item"`
	JID     string   `xml:"jid,attr,omitempty"`
	Node    string   `xml:"node,attr,omitempty"`
	Name    string   `xml:"name,attr,omitempty"`
}

// ============================================================================
// Registry init

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{NSDiscoInfo, "query"}, DiscoInfo{})
	TypeRegistry.MapExtension(PKTIQ, xml.Name{NSDiscoItems, "query"}, DiscoItems{})
}
