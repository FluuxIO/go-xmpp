package stanza

import (
	"encoding/xml"
	"strconv"
)

// ============================================================================
// XMPP Errors

// Err is an XMPP stanza payload that is used to report error on message,
// presence or iq stanza.
// It is intended to be added in the payload of the erroneous stanza.
type Err struct {
	XMLName xml.Name  `xml:"error"`
	Code    int       `xml:"code,attr,omitempty"`
	Type    ErrorType `xml:"type,attr"` // required
	Reason  string
	Text    string `xml:"urn:ietf:params:xml:ns:xmpp-stanzas text,omitempty"`
}

// UnmarshalXML implements custom parsing for XMPP errors
func (x *Err) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	x.XMLName = start.Name

	// Extract attributes
	for _, attr := range start.Attr {
		if attr.Name.Local == "type" {
			x.Type = ErrorType(attr.Value)
		}
		if attr.Name.Local == "code" {
			if code, err := strconv.Atoi(attr.Value); err == nil {
				x.Code = code
			}
		}
	}

	// Check subelements to extract error text and reason (from local namespace).
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
	if x.Code == 0 {
		return nil
	}

	// Encode start element and attributes
	start.Name = xml.Name{Local: "error"}

	code := xml.Attr{
		Name:  xml.Name{Local: "code"},
		Value: strconv.Itoa(x.Code),
	}
	start.Attr = append(start.Attr, code)

	if len(x.Type) > 0 {
		typ := xml.Attr{
			Name:  xml.Name{Local: "type"},
			Value: string(x.Type),
		}
		start.Attr = append(start.Attr, typ)
	}
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
