package xmpp

import "encoding/xml"

type clientIQ struct { // info/query
	XMLName xml.Name `xml:"jabber:client iq"`
	Packet
	Bind bindBind
	// TODO We need to support detecting the IQ namespace / Query packet
	// 	Error   clientError
}
