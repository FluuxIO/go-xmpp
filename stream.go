package xmpp // import "fluux.io/xmpp"

import "encoding/xml"

// XMPP Packet Parsing
type streamFeatures struct {
	XMLName    xml.Name `xml:"http://etherx.jabber.org/streams features"`
	StartTLS   tlsStartTLS
	Caps       Caps
	Mechanisms saslMechanisms
	Bind       bindBind
	Session    sessionSession
	Any        []xml.Name `xml:",any"`
}

type Caps struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/caps c"`
	Hash    string   `xml:"hash,attr"`
	Node    string   `xml:"node,attr"`
	Ver     string   `xml:"ver,attr"`
	Ext     string   `xml:"ext,attr,omitempty"`
}
