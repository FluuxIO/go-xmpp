package xmpp // import "fluux.io/xmpp"

import (
	"encoding/xml"
	"fmt"

	"fluux.io/xmpp/iot"
)

/*
TODO I would like to be able to write

	newIQ(Id, From, To, Type, Lang).AddPayload(IQPayload)

	xmpp.IQ{
		XMLName: xml.Name{
			Space: "",
			Local: "",
		},
		PacketAttrs: xmpp.PacketAttrs{
			Id:   "",
			From: "",
			To:   "",
			Type: "",
			Lang: "",
		},
		Payload: nil,
		RawXML:  "",
	}

*/

// ============================================================================
// IQ Packet

type IQ struct { // Info/Query
	XMLName xml.Name `xml:"iq"`
	PacketAttrs
	Payload []IQPayload `xml:",omitempty"`
	RawXML  string      `xml:",innerxml"`
	// 	Error   clientError
}

func NewIQ(iqtype, from, to, id, lang string) IQ {
	return IQ{
		XMLName: xml.Name{Local: "iq"},
		PacketAttrs: PacketAttrs{
			Id:   id,
			From: from,
			To:   to,
			Type: iqtype,
			Lang: lang,
		},
	}
}

func (iq *IQ) AddPayload(payload IQPayload) {
	iq.Payload = append(iq.Payload, payload)
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
			case "urn:xmpp:iot:control set":
				p = new(iot.ControlSet)
			default:
				p = new(Node)
			}
			if p != nil {
				err = d.DecodeElement(p, &tt)
				if err != nil {
					return err
				}
				iq.Payload = []IQPayload{p}
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
// Generic IQ Payload

type IQPayload interface {
	IsIQPayload()
}

type Node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	// Content []byte     `xml:",innerxml"`
	Nodes []Node `xml:",any"`
}

func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node Node
	return d.DecodeElement((*node)(n), &start)
}

func (n *Node) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	start.Attr = n.Attrs
	start.Name = n.XMLName

	err = e.EncodeToken(start)
	e.EncodeElement(n.Nodes, xml.StartElement{Name: n.XMLName})
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func (*Node) IsIQPayload() {}
