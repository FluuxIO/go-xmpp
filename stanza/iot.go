package stanza

import (
	"encoding/xml"
)

type ControlSet struct {
	XMLName xml.Name       `xml:"urn:xmpp:iot:control set"`
	Fields  []ControlField `xml:",any"`
	// Result sets
	ResultSet *ResultSet `xml:"set,omitempty"`
}

func (c *ControlSet) Namespace() string {
	return c.XMLName.Space
}

func (c *ControlSet) GetSet() *ResultSet {
	return c.ResultSet
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
func (c *ControlSetResponse) GetSet() *ResultSet {
	return nil
}

// ============================================================================
// Registry init

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{Space: "urn:xmpp:iot:control", Local: "set"}, ControlSet{})
}
