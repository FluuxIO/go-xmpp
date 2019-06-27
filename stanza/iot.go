package stanza

import (
	"encoding/xml"
)

type ControlSet struct {
	XMLName xml.Name       `xml:"urn:xmpp:iot:control set"`
	Fields  []ControlField `xml:",any"`
}

func (c *ControlSet) Namespace() string {
	return c.XMLName.Space
}

type ControlGetForm struct {
	XMLName xml.Name `xml:"urn:xmpp:iot:control getForm"`
}

type ControlField struct {
	XMLName xml.Name
	Name    string `xml:"name,attr,omitempty"`
	Value   string `xml:"value,attr,omitempty"`
}

type ControlSetResponse struct {
	XMLName xml.Name `xml:"urn:xmpp:iot:control setResponse"`
}

func (c *ControlSetResponse) Namespace() string {
	return c.XMLName.Space
}

// ============================================================================
// Registry init

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{"urn:xmpp:iot:control", "set"}, ControlSet{})
}
