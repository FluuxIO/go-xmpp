package stanza

import "encoding/xml"

// Open Packet
// Reference: WebSocket connections must start with this element
//            https://tools.ietf.org/html/rfc7395#section-3.4
type WebsocketOpen struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-framing open"`
	From    string   `xml:"from,attr"`
	Id      string   `xml:"id,attr"`
	Version string   `xml:"version,attr"`
}
