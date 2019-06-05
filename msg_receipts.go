package xmpp // import "gosrc.io/xmpp"

import "encoding/xml"

/*
Support for:
- XEP-0184 - Message Delivery Receipts: https://xmpp.org/extensions/xep-0184.html
*/

// Used on outgoing message, to tell the recipient that you are requesting a message receipt / ack.
type ReceiptRequest struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:receipts request"`
}

func (r ReceiptRequest) Name() xml.Name {
	return r.XMLName
}

type ReceiptReceived struct {
	MsgExtension
	XMLName xml.Name `xml:"urn:xmpp:receipts received"`
	ID      string
}

func (r ReceiptReceived) Name() xml.Name {
	return r.XMLName
}

func init() {
	typeRegistry.MapExtension(PKTMessage, ReceiptRequest{})
	typeRegistry.MapExtension(PKTMessage, ReceiptReceived{})
}
