package xmpp

type Packet interface {
	Name() string
}

// Attrs represents the common structure for base XMPP packets.
type Attrs struct {
	Id   string `xml:"id,attr,omitempty"`
	From string `xml:"from,attr,omitempty"`
	To   string `xml:"to,attr,omitempty"`
}

type packetFormatter interface {
	XMPPFormat() string
}
