package xmpp // import "fluux.io/xmpp"

import "encoding/xml"

// XMPP Packet Parsing
type ClientPresence struct {
	XMLName xml.Name `xml:"jabber:client presence"`
	Packet
	Show     string `xml:"show,attr,omitempty"` // away, chat, dnd, xa
	Status   string `xml:"status,attr,omitempty"`
	Priority string `xml:"priority,attr,omitempty"`
	//Error    *clientError
}
