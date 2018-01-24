package xmpp // import "fluux.io/xmpp"

import (
	"encoding/xml"
	"fmt"
)

// ============================================================================
// Message Packet

type Message struct {
	XMLName xml.Name `xml:"message"`
	PacketAttrs
	Subject string `xml:"subject,omitempty"`
	Body    string `xml:"body,omitempty"`
	Thread  string `xml:"thread,omitempty"`
	Error   Err    `xml:"error,omitempty"`
}

func (Message) Name() string {
	return "message"
}

func NewMessage(msgtype, from, to, id, lang string) Message {
	return Message{
		XMLName: xml.Name{Local: "message"},
		PacketAttrs: PacketAttrs{
			Id:   id,
			From: from,
			To:   to,
			Type: msgtype,
			Lang: lang,
		},
	}
}

type messageDecoder struct{}

var message messageDecoder

func (messageDecoder) decode(p *xml.Decoder, se xml.StartElement) (Message, error) {
	var packet Message
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

func (msg *Message) XMPPFormat() string {
	return fmt.Sprintf("<message to='%s' type='chat' xml:lang='en'>"+
		"<body>%s</body></message>",
		msg.To,
		xmlEscape(msg.Body))
}

// TODO: Func new message to create an empty message structure without the XML tag matching elements
