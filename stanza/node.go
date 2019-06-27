package stanza

import "encoding/xml"

// ============================================================================
// Generic / unknown content

// Node is a generic structure to represent XML data. It is used to parse
// unreferenced or custom stanza payload.
type Node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content string     `xml:",innerxml"`
	Nodes   []Node     `xml:",any"`
}

func (n *Node) Namespace() string {
	return n.XMLName.Space
}

// Attr represents generic XML attributes, as used on the generic XML Node
// representation.
type Attr struct {
	K string
	V string
}

// UnmarshalXML is a custom unmarshal function used by xml.Unmarshal to
// transform generic XML content into hierarchical Node structure.
func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Assign	"n.Attrs = start.Attr", without repeating xmlns in attributes:
	for _, attr := range start.Attr {
		// Do not repeat xmlns, it is already in XMLName
		if attr.Name.Local != "xmlns" {
			n.Attrs = append(n.Attrs, attr)
		}
	}
	type node Node
	return d.DecodeElement((*node)(n), &start)
}

// MarshalXML is a custom XML serializer used by xml.Marshal to serialize a
// Node structure to XML.
func (n Node) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	start.Attr = n.Attrs
	start.Name = n.XMLName

	err = e.EncodeToken(start)
	e.EncodeElement(n.Nodes, xml.StartElement{Name: n.XMLName})
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}
