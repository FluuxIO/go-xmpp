package stanza

import (
	"encoding/xml"
	"reflect"
)

// ============================================================================
// Presence Packet

// Presence implements RFC 6120 - A.5 Client Namespace (a part)
type Presence struct {
	XMLName xml.Name `xml:"presence"`
	Attrs
	Show       PresenceShow    `xml:"show,omitempty"`
	Status     string          `xml:"status,omitempty"`
	Priority   int8            `xml:"priority,omitempty"` // default: 0
	Error      Err             `xml:"error,omitempty"`
	Extensions []PresExtension `xml:",omitempty"`
}

func (Presence) Name() string {
	return "presence"
}

func NewPresence(a Attrs) Presence {
	return Presence{
		XMLName: xml.Name{Local: "presence"},
		Attrs:   a,
	}
}

// Get search and extracts a specific extension on a presence stanza.
// It receives a pointer to an PresExtension. It will panic if the caller
// does not pass a pointer.
// It will return true if the passed extension is found and set the pointer
// to the extension passed as parameter to the found extension.
// It will return false if the extension is not found on the presence.
//
// Example usage:
//   var muc xmpp.MucPresence
//   if ok := msg.Get(&muc); ok {
//     // muc presence extension has been found
//	 }
func (pres *Presence) Get(ext PresExtension) bool {
	target := reflect.ValueOf(ext)
	if target.Kind() != reflect.Ptr {
		panic("you must pass a pointer to the message Get method")
	}

	for _, e := range pres.Extensions {
		if reflect.TypeOf(e) == target.Type() {
			source := reflect.ValueOf(e)
			if source.Kind() != reflect.Ptr {
				source = source.Elem()
			}
			target.Elem().Set(source.Elem())
			return true
		}
	}
	return false
}

type presenceDecoder struct{}

var presence presenceDecoder

func (presenceDecoder) decode(p *xml.Decoder, se xml.StartElement) (Presence, error) {
	var packet Presence
	err := p.DecodeElement(&packet, &se)
	// TODO Add default presence type (when omitted)
	return packet, err
}

// UnmarshalXML implements custom parsing for presence stanza
func (pres *Presence) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	pres.XMLName = start.Name

	// Extract packet attributes
	for _, attr := range start.Attr {
		if attr.Name.Local == "id" {
			pres.Id = attr.Value
		}
		if attr.Name.Local == "type" {
			pres.Type = StanzaType(attr.Value)
		}
		if attr.Name.Local == "to" {
			pres.To = attr.Value
		}
		if attr.Name.Local == "from" {
			pres.From = attr.Value
		}
		if attr.Name.Local == "lang" {
			pres.Lang = attr.Value
		}
	}

	// decode inner elements
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch tt := t.(type) {

		case xml.StartElement:
			if presExt := TypeRegistry.GetPresExtension(tt.Name); presExt != nil {
				// Decode message extension
				err = d.DecodeElement(presExt, &tt)
				if err != nil {
					return err
				}
				pres.Extensions = append(pres.Extensions, presExt)
			} else {
				// Decode standard message sub-elements
				var err error
				switch tt.Name.Local {
				case "show":
					err = d.DecodeElement(&pres.Show, &tt)
				case "status":
					err = d.DecodeElement(&pres.Status, &tt)
				case "priority":
					err = d.DecodeElement(&pres.Priority, &tt)
				case "error":
					err = d.DecodeElement(&pres.Error, &tt)
				}
				if err != nil {
					return err
				}
			}

		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}
	}
}
