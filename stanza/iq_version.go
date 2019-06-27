package stanza

import "encoding/xml"

// ============================================================================
// Software Version (XEP-0092)

// Version
type Version struct {
	XMLName xml.Name `xml:"jabber:iq:version query"`
	Name    string   `xml:"name,omitempty"`
	Version string   `xml:"version,omitempty"`
	OS      string   `xml:"os,omitempty"`
}

func (v *Version) Namespace() string {
	return v.XMLName.Space
}

// ============================================================================
// Registry init

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{"jabber:iq:version", "query"}, Version{})
}
