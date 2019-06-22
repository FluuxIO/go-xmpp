package xmpp

import "encoding/xml"

// ============================================================================
// Presence Packet

// Presence implements RFC 6120 - A.5 Client Namespace (a part)
type Presence struct {
	XMLName xml.Name `xml:"presence"`
	Attrs
	Show     PresenceShow `xml:"show,omitempty"`
	Status   string       `xml:"status,omitempty"`
	Priority int8         `xml:"priority,omitempty"` // default: 0
	Error    Err          `xml:"error,omitempty"`
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
