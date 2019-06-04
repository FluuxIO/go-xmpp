package xmpp // import "gosrc.io/xmpp"

import "encoding/xml"

/*
Support for:
- XEP-0333 - Chat Markers: https://xmpp.org/extensions/xep-0333.html
*/

type Markable struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:chat-markers:0 markable"`
}

type MarkReceived struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:chat-markers:0 received"`
	ID      string
}

func init() {
	typeRegistry.MapExtension(PKTMessage, xml.Name{"urn:xmpp:chat-markers:0", "markable"}, Markable{})
	typeRegistry.MapExtension(PKTMessage, xml.Name{"urn:xmpp:chat-markers:0", "received"}, MarkReceived{})
}
