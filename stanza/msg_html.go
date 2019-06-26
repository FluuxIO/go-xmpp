package stanza

import (
	"encoding/xml"
)

type HTML struct {
	MsgExtension
	XMLName xml.Name `xml:"http://jabber.org/protocol/xhtml-im html"`
	Body    HTMLBody
	Lang    string `xml:"xml:lang,attr,omitempty"`
}

type HTMLBody struct {
	XMLName xml.Name `xml:"http://www.w3.org/1999/xhtml body"`
	// InnerXML MUST be valid xhtml. We do not check if it is valid when generating the XMPP stanza.
	InnerXML string `xml:",innerxml"`
}

func init() {
	TypeRegistry.MapExtension(PKTMessage, xml.Name{"http://jabber.org/protocol/xhtml-im", "html"}, HTML{})
}
