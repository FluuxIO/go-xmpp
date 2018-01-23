package xmpp // import "fluux.io/xmpp"

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"strconv"

	"fluux.io/xmpp/iot"
)

/*
TODO support ability to put Raw payload
*/

// ============================================================================
// XMPP Errors

type Err struct {
	XMLName xml.Name `xml:"error"`
	Code    int      `xml:"code,attr,omitempty"`
	Type    string   `xml:"type,attr,omitempty"`
	Reason  string
	Text    string `xml:"urn:ietf:params:xml:ns:xmpp-stanzas text,omitempty"`
}

func (*Err) IsIQPayload() {}

// UnmarshalXML implements custom parsing for IQs
func (x *Err) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	x.XMLName = start.Name

	// Extract attributes
	for _, attr := range start.Attr {
		if attr.Name.Local == "type" {
			x.Type = attr.Value
		}
		if attr.Name.Local == "code" {
			if code, err := strconv.Atoi(attr.Value); err == nil {
				x.Code = code
			}
		}
	}

	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch tt := t.(type) {

		case xml.StartElement:
			elt := new(Node)

			err = d.DecodeElement(elt, &tt)
			if err != nil {
				return err
			}

			textName := xml.Name{Space: "urn:ietf:params:xml:ns:xmpp-stanzas", Local: "text"}
			if elt.XMLName == textName {
				x.Text = string(elt.Content)
			} else if elt.XMLName.Space == "urn:ietf:params:xml:ns:xmpp-stanzas" {
				x.Reason = elt.XMLName.Local
			}

		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}
	}
}

func (x Err) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	code := xml.Attr{
		Name:  xml.Name{Local: "code"},
		Value: strconv.Itoa(x.Code),
	}
	typ := xml.Attr{
		Name:  xml.Name{Local: "type"},
		Value: x.Type,
	}
	start.Name = xml.Name{Local: "error"}
	start.Attr = append(start.Attr, code, typ)
	err = e.EncodeToken(start)

	// SubTags
	// Reason
	if x.Reason != "" {
		reason := xml.Name{Space: "urn:ietf:params:xml:ns:xmpp-stanzas", Local: x.Reason}
		e.EncodeToken(xml.StartElement{Name: reason})
		e.EncodeToken(xml.EndElement{Name: reason})
	}

	// Text
	if x.Text != "" {
		text := xml.Name{Space: "urn:ietf:params:xml:ns:xmpp-stanzas", Local: "text"}
		e.EncodeToken(xml.StartElement{Name: text})
		e.EncodeToken(xml.CharData(x.Text))
		e.EncodeToken(xml.EndElement{Name: text})
	}

	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// ============================================================================
// IQ Packet

type IQ struct { // Info/Query
	XMLName xml.Name `xml:"iq"`
	PacketAttrs
	Payload []IQPayload `xml:",omitempty"`
	RawXML  string      `xml:",innerxml"`
	Error   Err         `xml:"error,omitempty"`
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

func (iq IQ) MakeError(xerror Err) IQ {
	from := iq.From
	to := iq.To

	iq.Type = "error"
	iq.From = to
	iq.To = from
	iq.Error = xerror

	return iq
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
	level := 0
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch tt := t.(type) {

		case xml.StartElement:
			level++
			if level <= 1 {
				var elt interface{}
				payloadType := tt.Name.Space + " " + tt.Name.Local
				if payloadType := typeRegistry[payloadType]; payloadType != nil {
					val := reflect.New(payloadType)
					elt = val.Interface()
				} else {
					elt = new(Node)
				}

				if iqPl, ok := elt.(IQPayload); ok {
					err = d.DecodeElement(elt, &tt)
					if err != nil {
						return err
					}
					iq.Payload = append(iq.Payload, iqPl)
				}
			}

		case xml.EndElement:
			level--
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
	Content string     `xml:",innerxml"`
	Nodes   []Node     `xml:",any"`
}

type Attr struct {
	K string
	V string
}

func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Assign	"n.Attrs = start.Attr", without repeating xmlns in attributes:
	for _, attr := range start.Attr {
		// Do not repeat xmlns, it is already in XMLName
		if attr.Name.Local != "xmlns" {
			n.Attrs = append(n.Attrs, attr)
		}
	}
	type node Node
	return d.DecodeElement((*node)(n), &start)
}

func (n Node) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	start.Attr = n.Attrs
	start.Name = n.XMLName

	err = e.EncodeToken(start)
	e.EncodeElement(n.Nodes, xml.StartElement{Name: n.XMLName})
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func (*Node) IsIQPayload() {}

// ============================================================================
// Disco

const (
	NSDiscoInfo = "http://jabber.org/protocol/disco#info"
)

type DiscoInfo struct {
	XMLName  xml.Name  `xml:"http://jabber.org/protocol/disco#info query"`
	Identity Identity  `xml:"identity"`
	Features []Feature `xml:"feature"`
}

func (*DiscoInfo) IsIQPayload() {}

type Identity struct {
	XMLName  xml.Name `xml:"identity,omitempty"`
	Name     string   `xml:"name,attr,omitempty"`
	Category string   `xml:"category,attr,omitempty"`
	Type     string   `xml:"type,attr,omitempty"`
}

type Feature struct {
	XMLName xml.Name `xml:"feature"`
	Var     string   `xml:"var,attr"`
}

// ============================================================================

var typeRegistry = make(map[string]reflect.Type)

func init() {
	typeRegistry["http://jabber.org/protocol/disco#info query"] = reflect.TypeOf(DiscoInfo{})
	typeRegistry["urn:ietf:params:xml:ns:xmpp-bind bind"] = reflect.TypeOf(BindBind{})
	typeRegistry["urn:xmpp:iot:control set"] = reflect.TypeOf(iot.ControlSet{})
}
