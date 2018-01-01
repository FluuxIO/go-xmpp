package xmpp // import "fluux.io/xmpp"

// Packet represents the root default structure for an XMPP packet.
type Packet struct {
	Id   string `xml:"id,attr,omitempty"`
	From string `xml:"from,attr,omitempty"`
	To   string `xml:"to,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
	Lang string `xml:"lang,attr,omitempty"`
}

type packetFormatter interface {
	XMPPFormat() string
}
