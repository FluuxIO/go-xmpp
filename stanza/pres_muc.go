package stanza

import (
	"encoding/xml"
	"strconv"
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

const timeLayout = "2006-01-02T15:04:05Z"

// History implements XEP-0045: Multi-User Chat - 19.1
type History struct {
	XMLName    xml.Name
	MaxChars   NullableInt `xml:"maxchars,attr,omitempty"`
	MaxStanzas NullableInt `xml:"maxstanzas,attr,omitempty"`
	Seconds    NullableInt `xml:"seconds,attr,omitempty"`
	Since      time.Time   `xml:"since,attr,omitempty"`
}

type NullableInt struct {
	Value int
	isSet bool
}

func NewNullableInt(val int) NullableInt {
	return NullableInt{val, true}
}

func (n NullableInt) Get() (v int, ok bool) {
	return n.Value, n.isSet
}

// UnmarshalXML implements custom parsing for history element
func (h *History) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	h.XMLName = start.Name

	// Extract attributes
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "maxchars":
			v, err := strconv.Atoi(attr.Value)
			if err != nil {
				return err
			}
			h.MaxChars = NewNullableInt(v)
		case "maxstanzas":
			v, err := strconv.Atoi(attr.Value)
			if err != nil {
				return err
			}
			h.MaxStanzas = NewNullableInt(v)
		case "seconds":
			v, err := strconv.Atoi(attr.Value)
			if err != nil {
				return err
			}
			h.Seconds = NewNullableInt(v)
		case "since":
			t, err := time.Parse(timeLayout, attr.Value)
			if err != nil {
				return err
			}
			h.Since = t
		}
	}

	// Consume remaining data until element end
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch tt := t.(type) {
		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}
	}
}

func (h History) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	mc, isMcSet := h.MaxChars.Get()
	ms, isMsSet := h.MaxStanzas.Get()
	s, isSSet := h.Seconds.Get()

	// We do not have any value, ignore history element
	if h.Since.IsZero() && !isMcSet && !isMsSet && !isSSet {
		return nil
	}

	// Encode start element and attributes
	start.Name = xml.Name{Local: "history"}

	if isMcSet {
		attr := xml.Attr{
			Name:  xml.Name{Local: "maxchars"},
			Value: strconv.Itoa(mc),
		}
		start.Attr = append(start.Attr, attr)
	}

	if isMsSet {
		attr := xml.Attr{
			Name:  xml.Name{Local: "maxstanzas"},
			Value: strconv.Itoa(ms),
		}
		start.Attr = append(start.Attr, attr)
	}

	if isSSet {
		attr := xml.Attr{
			Name:  xml.Name{Local: "seconds"},
			Value: strconv.Itoa(s),
		}
		start.Attr = append(start.Attr, attr)
	}

	if !h.Since.IsZero() {
		attr := xml.Attr{
			Name:  xml.Name{Local: "since"},
			Value: h.Since.Format(timeLayout),
		}
		start.Attr = append(start.Attr, attr)
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	return e.EncodeToken(xml.EndElement{Name: start.Name})

}

func init() {
	TypeRegistry.MapExtension(PKTPresence, xml.Name{"http://jabber.org/protocol/muc", "x"}, MucPresence{})
}
