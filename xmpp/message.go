package xmpp

import (
	"encoding/xml"
	"fmt"
)

// XMPP Packet Parsing
type ClientMessage struct {
	XMLName xml.Name `xml:"jabber:client message"`
	Packet
	Subject string `xml:"subject,omitempty"`
	Body    string `xml:"body,omitempty"`
	Thread  string `xml:"thread,omitempty"`
}

// TODO: Func new message to create an empty message structure without the XML tag matching elements

func (message *ClientMessage) XMPPFormat() string {
	return fmt.Sprintf("<message to='%s' type='chat' xml:lang='en'>"+
		"<body>%s</body></message>",
		message.To,
		xmlEscape(message.Body))
}
