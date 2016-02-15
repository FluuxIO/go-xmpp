package iot

import "encoding/xml"

type Control struct {
	ControlSet     ControlSet     `xml:",omitempty"`
	ControlGetForm ControlGetForm `xml:",omitempty"`
}

type ControlSet struct {
	XMLName xml.Name       `xml:"urn:xmpp:iot:control set"`
	Fields  []ControlField `xml:",any"`
}

type ControlGetForm struct {
	XMLName xml.Name `xml:"urn:xmpp:iot:control getForm"`
}

type ControlField struct {
	XMLName xml.Name
	Name    string `xml:"name,attr,omitempty"`
	Value   string `xml:"value,attr,omitempty"`
}
