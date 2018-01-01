package iot // import "fluux.io/xmpp/iot"

import "encoding/xml"

type ControlSet struct {
	XMLName xml.Name       `xml:"urn:xmpp:iot:control set"`
	Fields  []ControlField `xml:",any"`
}

func (*ControlSet) IsIQPayload() {
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

func (*ControlSetResponse) IsIQPayload() {
}
