package xmpp // import "gosrc.io/xmpp"

import "encoding/xml"

/*
Support for:
- XEP-0066 - Out of Band Data: https://xmpp.org/extensions/xep-0066.html
*/

type OOB struct {
	MsgExtension
	XMLName xml.Name `xml:"jabber:x:oob x"`
	URL     string   `xml:"url"`
	Desc    string   `xml:"desc,omitempty"`
}

func (o OOB) Name() xml.Name {
	return o.XMLName
}

func init() {
	typeRegistry.MapExtension(PKTMessage, OOB{})
}
