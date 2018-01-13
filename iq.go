package xmpp // import "fluux.io/xmpp"

import (
	"encoding/xml"
	"fmt"

	"fluux.io/xmpp/iot"
)

// ============================================================================
// IQ Packet

type IQ struct { // Info/Query
	XMLName xml.Name `xml:"iq"`
	PacketAttrs
	Payload IQPayload `xml:",omitempty"`
	RawXML  string    `xml:",innerxml"`
	// TODO We need to support detecting the IQ namespace / Query packet
	// 	Error   clientError
}

func (IQ) Name() string {
	return "iq"
}

type iqDecoder struct{}

var iq iqDecoder

func (iqDecoder) decode(p *xml.Decoder, se xml.StartElement) (IQ, error) {
	var packet IQ
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

type IQPayload interface {
	IsIQPayload()
}

// UnmarshalXML implements custom parsing for IQs
func (iq *IQ) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	iq.XMLName = start.Name
	// Extract IQ attributes
	for _, attr := range start.Attr {
		if attr.Name.Local == "id" {
			iq.Id = attr.Value
		}
		if attr.Name.Local == "type" {
			iq.Type = attr.Value
		}
		if attr.Name.Local == "to" {
			iq.To = attr.Value
		}
		if attr.Name.Local == "from" {
			iq.From = attr.Value
		}
		if attr.Name.Local == "lang" {
			iq.Lang = attr.Value
		}
	}

	// decode inner elements
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		var p IQPayload
		switch tt := t.(type) {

		case xml.StartElement:
			switch tt.Name.Space + " " + tt.Name.Local {
			case "urn:ietf:params:xml:ns:xmpp-bind bind":
				p = new(bindBind)
			case "http://jabber.org/protocol/disco#items query":
				p = new(Query)
			case "urn:xmpp:iot:control set":
				p = new(iot.ControlSet)
				// TODO: Add a default Type that passes RawXML
			}
			if p != nil {
				err = d.DecodeElement(p, &tt)
				if err != nil {
					return err
				}
				iq.Payload = p
				p = nil
			}

		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}
	}
}

// XMPPFormat returns the string representation of the XMPP packet.
// TODO: Should I simply rely on xml.Marshal ?
func (iq *IQ) XMPPFormat() string {
	if iq.Payload != nil {
		var payload []byte
		var err error
		if payload, err = xml.Marshal(iq.Payload); err != nil {
			return fmt.Sprintf("<iq to='%s' type='%s' id='%s' xml:lang='en'>"+
				"</iq>",
				iq.To, iq.Type, iq.Id)
		}
		return fmt.Sprintf("<iq to='%s' type='%s' id='%s' xml:lang='en'>"+
			"%s</iq>",
			iq.To, iq.Type, iq.Id, payload)
	}
	return fmt.Sprintf("<iq to='%s' type='%s' id='%s' xml:lang='en'>"+
		"%s</iq>",
		iq.To, iq.Type, iq.Id,
		iq.RawXML)
}

// ============================================================================
// IQ Query subelement

type Query struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/disco#items query"`
}

func (*Query) IsIQPayload() {}
