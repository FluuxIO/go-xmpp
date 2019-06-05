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

func (m Markable) Name() xml.Name {
	return m.XMLName
}

type MarkReceived struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:chat-markers:0 received"`
	ID      string
}

func (m MarkReceived) Name() xml.Name {
	return m.XMLName
}

func init() {
	typeRegistry.MapExtension(PKTMessage, Markable{})
	typeRegistry.MapExtension(PKTMessage, MarkReceived{})
}
