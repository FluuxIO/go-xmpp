package xmpp

import (
	"encoding/xml"

	"github.com/processone/gox/xmpp/iot"
)

// info/query
type ClientIQ struct {
	XMLName xml.Name `xml:"jabber:client iq"`
	Packet
	Bind bindBind `xml:",omitempty"`
	iot.Control
	RawXML string `xml:",innerxml"`
	// TODO We need to support detecting the IQ namespace / Query packet
	// 	Error   clientError
}
