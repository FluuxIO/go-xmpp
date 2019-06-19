package xmpp

import "encoding/xml"

// ============================================================================
// Presence Packet

type Presence struct {
	XMLName xml.Name `xml:"presence"`
	Attrs
	Show     string `xml:"show,omitempty"` // away, chat, dnd, xa
	Status   string `xml:"status,omitempty"`
	Priority int    `xml:"priority,omitempty"`
	Error    Err    `xml:"error,omitempty"`
}

func (Presence) Name() string {
	return "presence"
}

func NewPresence(a Attrs) Presence {
	return Presence{
		XMLName: xml.Name{Local: "presence"},
		Attrs:   a,
	}
}

type presenceDecoder struct{}

var presence presenceDecoder

func (presenceDecoder) decode(p *xml.Decoder, se xml.StartElement) (Presence, error) {
	var packet Presence
	err := p.DecodeElement(&packet, &se)
	// TODO Add default presence type (when omitted)
	return packet, err
}
