package stanza

type Packet interface {
	Name() string
}

// Attrs represents the common structure for base XMPP packets.
type Attrs struct {
	Type StanzaType `xml:"type,attr,omitempty"`
	Id   string     `xml:"id,attr,omitempty"`
	From string     `xml:"from,attr,omitempty"`
	To   string     `xml:"to,attr,omitempty"`
	Lang string     `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
}

type packetFormatter interface {
	XMPPFormat() string
}
