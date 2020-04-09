package stanza

import "encoding/xml"

/*
Support for:
- XEP-0334: Message Processing Hints: https://xmpp.org/extensions/xep-0334.html
Pointers should be used to keep consistent with unmarshal. Eg :
msg.Extensions = append(msg.Extensions, &stanza.HintNoCopy{}, &stanza.HintStore{})
*/

type HintNoPermanentStore struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:hints no-permanent-store"`
}

type HintNoStore struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:hints no-store"`
}

type HintNoCopy struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:hints no-copy"`
}
type HintStore struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:hints store"`
}

func init() {
	TypeRegistry.MapExtension(PKTMessage, xml.Name{Space: "urn:xmpp:hints", Local: "no-permanent-store"}, HintNoPermanentStore{})
	TypeRegistry.MapExtension(PKTMessage, xml.Name{Space: "urn:xmpp:hints", Local: "no-store"}, HintNoStore{})
	TypeRegistry.MapExtension(PKTMessage, xml.Name{Space: "urn:xmpp:hints", Local: "no-copy"}, HintNoCopy{})
	TypeRegistry.MapExtension(PKTMessage, xml.Name{Space: "urn:xmpp:hints", Local: "store"}, HintStore{})
}
