package stanza

import (
	"encoding/xml"
)

// XEP-0070: Verifying HTTP Requests via XMPP
type ConfirmPayload struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/http-auth confirm"`
	ID      string   `xml:"id,attr"`
	Method  string   `xml:"method,attr"`
	URL     string   `xml:"url,attr"`
}

func (c ConfirmPayload) Namespace() string {
	return c.XMLName.Space
}

func (c ConfirmPayload) GetSet() *ResultSet {
	return nil
}

// ---------------
// Builder helpers

// Confirm builds a default confirm payload
func (iq *IQ) Confirm() *ConfirmPayload {
	d := ConfirmPayload{
		XMLName: xml.Name{Space: "http://jabber.org/protocol/http-auth", Local: "confirm"},
	}
	iq.Payload = &d
	return &d
}

// Set all confirm info
func (v *ConfirmPayload) SetConfirm(id, method, url string) *ConfirmPayload {
	v.ID = id
	v.Method = method
	v.URL = url
	return v
}

func init() {
	TypeRegistry.MapExtension(PKTMessage,
		xml.Name{Space: "http://jabber.org/protocol/http-auth", Local: "confirm"},
		ConfirmPayload{})
	TypeRegistry.MapExtension(PKTIQ,
		xml.Name{Space: "http://jabber.org/protocol/http-auth", Local: "confirm"},
		ConfirmPayload{})
}
