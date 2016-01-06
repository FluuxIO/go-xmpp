package xmpp

import "encoding/xml"

// XMPP Packet Parsing
type clientPresence struct {
	XMLName xml.Name `xml:"jabber:client presence"`
	Packet
	Show     string `xml:"show,attr,omitempty"` // away, chat, dnd, xa
	Status   string `xml:"status,attr,omitempty"`
	Priority string `xml:"priority,attr,omitempty"`
	//Error    *clientError
}
