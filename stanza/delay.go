package stanza

import "encoding/xml"

/*
Support for:
- XEP-0203: Delayed Delivery: https://xmpp.org/extensions/xep-0203.html
*/

const NSDelay = "urn:xmpp:delay"

type Delay struct {
	XMLName xml.Name   `xml:"urn:xmpp:delay delay"`
	From    string     `xml:"from,attr"`
	date    JabberDate // TODO
}

func init() {
	TypeRegistry.MapExtension(PKTMessage, xml.Name{Space: NSDelay, Local: "delay"}, Delay{})
}
