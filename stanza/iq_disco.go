package stanza

import "encoding/xml"

// ============================================================================
// Disco

const (
	NSDiscoInfo  = "http://jabber.org/protocol/disco#info"
	NSDiscoItems = "http://jabber.org/protocol/disco#items"
)

// Disco Info
type DiscoInfo struct {
	XMLName  xml.Name  `xml:"http://jabber.org/protocol/disco#info query"`
	Node     string    `xml:"node,attr,omitempty"`
	Identity Identity  `xml:"identity"`
	Features []Feature `xml:"feature"`
}

func (d *DiscoInfo) Namespace() string {
	return d.XMLName.Space
}

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

// Disco Items
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
