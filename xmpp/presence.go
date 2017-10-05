package xmpp

import (
	"encoding/xml"
	"fmt"
)

// XMPP Packet Parsing
type ClientPresence struct {
	XMLName xml.Name `xml:"jabber:client presence"`
	Packet
	Show     string `xml:"show,attr,omitempty"` // away, chat, dnd, xa
	Status   string `xml:"status,attr,omitempty"`
	Priority string `xml:"priority,attr,omitempty"`
	//Error    *clientError
}

func (message *ClientPresence) XMPPFormat() string {
	return fmt.Sprintf("<presence xml:lang='en' from='%s' to='%s'><show>%s</show><status>%s</status></presence>",
		message.From, message.To, message.Show, message.Status)
}
