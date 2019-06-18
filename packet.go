package xmpp

type Packet interface {
	Name() string
}

// PacketAttrs represents the common structure for base XMPP packets.
type PacketAttrs struct {
	Id   string `xml:"id,attr,omitempty"`
	From string `xml:"from,attr,omitempty"`
	To   string `xml:"to,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
	Lang string `xml:"lang,attr,omitempty"`
}

type packetFormatter interface {
	XMPPFormat() string
}
