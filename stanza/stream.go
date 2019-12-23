package stanza

import "encoding/xml"

// Start of stream
// Reference: XMPP Core stream open
//            https://tools.ietf.org/html/rfc6120#section-4.2
type Stream struct {
	XMLName xml.Name `xml:"http://etherx.jabber.org/streams stream"`
	From    string   `xml:"from,attr"`
	To      string   `xml:"to,attr"`
	Id      string   `xml:"id,attr"`
	Version string   `xml:"version,attr"`
}

const StreamClose = "</stream:stream>"
