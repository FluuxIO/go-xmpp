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

// ---------------
// Builder helpers

// Version builds a default software version payload
func (iq *IQ) Version() *Version {
	d := Version{
		XMLName: xml.Name{Space: "jabber:iq:version", Local: "query"},
	}
	iq.Payload = &d
	return &d
}

// Set all software version info
func (v *Version) SetInfo(name, version, os string) *Version {
	v.Name = name
	v.Version = version
	v.OS = os
	return v
}

// ============================================================================
// Registry init

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{"jabber:iq:version", "query"}, Version{})
}
