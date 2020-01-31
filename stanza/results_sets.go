package stanza

import (
	"encoding/xml"
)

// Support for XEP-0059
// See https://xmpp.org/extensions/xep-0059
const (
	// Common but not only possible namespace for query blocks in a result set context
	NSQuerySet = "jabber:iq:search"
)

type ResultSet struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/rsm set"`
	After   *string  `xml:"after,omitempty"`
	Before  *string  `xml:"before,omitempty"`
	Count   *int     `xml:"count,omitempty"`
	First   *First   `xml:"first,omitempty"`
	Index   *int     `xml:"index,omitempty"`
	Last    *string  `xml:"last,omitempty"`
	Max     *int     `xml:"max,omitempty"`
}

type First struct {
	XMLName xml.Name `xml:"first"`
	Content string
	Index   *int `xml:"index,attr,omitempty"`
}
