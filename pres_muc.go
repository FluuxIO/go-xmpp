package xmpp

import (
	"encoding/xml"
	"time"
)

// ============================================================================
// MUC Presence extension

// MucPresence implements XEP-0045: Multi-User Chat - 19.1
type MucPresence struct {
	PresExtension
	XMLName  xml.Name `xml:"http://jabber.org/protocol/muc x"`
	Password string   `xml:"password,omitempty"`
	History  History  `xml:"history,omitempty"`
}

// History implements XEP-0045: Multi-User Chat - 19.1
type History struct {
	MaxChars   *int       `xml:"maxchars,attr,omitempty"`
	MaxStanzas *int       `xml:"maxstanzas,attr,omitempty"`
	Seconds    *int       `xml:"seconds,attr,omitempty"`
	Since      *time.Time `xml:"since,attr,omitempty"`
}

func init() {
	TypeRegistry.MapExtension(PKTPresence, xml.Name{"http://jabber.org/protocol/muc", "x"}, MucPresence{})
}
