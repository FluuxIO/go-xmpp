package stanza

import (
	"encoding/xml"
)

/*
Support for:
- XEP-0085 - Chat State Notifications: https://xmpp.org/extensions/xep-0085.html
*/

const NSMsgChatStateNotifications = "http://jabber.org/protocol/chatstates"

type StateActive struct {
	MsgExtension
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates active"`
}

type StateComposing struct {
	MsgExtension
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates composing"`
}

type StateGone struct {
	MsgExtension
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates gone"`
}

type StateInactive struct {
	MsgExtension
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates inactive"`
}

type StatePaused struct {
	MsgExtension
	XMLName xml.Name `xml:"http://jabber.org/protocol/chatstates paused"`
}

func init() {
	TypeRegistry.MapExtension(PKTMessage, xml.Name{NSMsgChatStateNotifications, "active"}, StateActive{})
	TypeRegistry.MapExtension(PKTMessage, xml.Name{NSMsgChatStateNotifications, "composing"}, StateComposing{})
	TypeRegistry.MapExtension(PKTMessage, xml.Name{NSMsgChatStateNotifications, "gone"}, StateGone{})
	TypeRegistry.MapExtension(PKTMessage, xml.Name{NSMsgChatStateNotifications, "inactive"}, StateInactive{})
	TypeRegistry.MapExtension(PKTMessage, xml.Name{NSMsgChatStateNotifications, "paused"}, StatePaused{})
}
