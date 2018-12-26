package xmpp // import "gosrc.io/xmpp"

import (
	"encoding/xml"
)

// ============================================================================
// StreamFeatures Packet

type streamFeatures struct {
	XMLName    xml.Name `xml:"http://etherx.jabber.org/streams features"`
	StartTLS   tlsStartTLS
	Caps       Caps
	Mechanisms saslMechanisms
	Bind       BindBind
	Session    sessionSession
	Any        []xml.Name `xml:",any"`
}

// ============================================================================
// StreamError Packet

type StreamError struct {
	XMLName xml.Name `xml:"http://etherx.jabber.org/streams error"`
	Error   xml.Name `xml:",any"`
}

func (StreamError) Name() string {
	return "stream:error"
}

type streamErrorDecoder struct{}

var streamError streamErrorDecoder

func (streamErrorDecoder) decode(p *xml.Decoder, se xml.StartElement) (StreamError, error) {
	var packet StreamError
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

// ============================================================================
// Caps subElement

type Caps struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/caps c"`
	Hash    string   `xml:"hash,attr"`
	Node    string   `xml:"node,attr"`
	Ver     string   `xml:"ver,attr"`
	Ext     string   `xml:"ext,attr,omitempty"`
}
