package stanza

import (
	"encoding/xml"
)

/*
Support for:
- XEP-0333 - Chat Markers: https://xmpp.org/extensions/xep-0333.html
*/

const NSMsgChatMarkers = "urn:xmpp:chat-markers:0"

type Markable struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:chat-markers:0 markable"`
}

type MarkReceived struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:chat-markers:0 received"`
	ID      string   `xml:"id,attr"`
}

type MarkDisplayed struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:chat-markers:0 displayed"`
	ID      string   `xml:"id,attr"`
}

type MarkAcknowledged struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:chat-markers:0 acknowledged"`
	ID      string   `xml:"id,attr"`
}

func init() {
	TypeRegistry.MapExtension(PKTMessage, xml.Name{NSMsgChatMarkers, "markable"}, Markable{})
	TypeRegistry.MapExtension(PKTMessage, xml.Name{NSMsgChatMarkers, "received"}, MarkReceived{})
	TypeRegistry.MapExtension(PKTMessage, xml.Name{NSMsgChatMarkers, "displayed"}, MarkDisplayed{})
	TypeRegistry.MapExtension(PKTMessage, xml.Name{NSMsgChatMarkers, "acknowledged"}, MarkAcknowledged{})
}
