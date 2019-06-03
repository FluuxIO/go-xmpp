package xmpp

import "encoding/xml"

/*
Support for:
- XEP-0184 - Message Delivery Receipts: https://xmpp.org/extensions/xep-0184.html
*/

const (
	NSReceipts = "urn:xmpp:receipts"
)

// XEP-0184 message receipt markers
type Receipt struct {
	MsgExtension
	XMLName xml.Name
	Id      string
}

func init() {
	typeRegistry.RegisterMsgExt(NSReceipts, Receipt{})
}
