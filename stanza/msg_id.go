package stanza

import "encoding/xml"

/*
Support for:
- XEP-0313 - Message Archive Management (MAM): https://xmpp.org/extensions/xep-0313.html
This MUST NOT be interpreted as an archive ID unless the server has previously advertised support for 'urn:xmpp:mam:2'
See : https://xmpp.org/extensions/xep-0313.html#archives_id
*/

type StanzaId struct {
	XMLName xml.Name `xml:"urn:xmpp:sid:0 stanza-id"`
	By      string   `xml:"by,attr"`
	Id      string   `xml:"id,attr"`
}
