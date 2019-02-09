package xmpp // import "gosrc.io/xmpp"

import "encoding/xml"

// ============================================================================
// Presence Packet

type Presence struct {
	XMLName xml.Name `xml:"presence"`
	PacketAttrs
	Show     string `xml:"show,omitempty"` // away, chat, dnd, xa
	Status   string `xml:"status,omitempty"`
	Priority string `xml:"priority,omitempty"`
	Error    Err    `xml:"error,omitempty"`
}

func (Presence) Name() string {
	return "presence"
}

func NewPresence(from, to, id, lang string) Presence {
	return Presence{
		XMLName: xml.Name{Local: "presence"},
		PacketAttrs: PacketAttrs{
			Id:   id,
			From: from,
			To:   to,
			Lang: lang,
		},
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
