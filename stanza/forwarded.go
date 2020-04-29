package stanza

import "encoding/xml"

// Forwarded is used to wrapped forwarded stanzas.
type Forwarded struct {
	XMLName xml.Name `xml:"urn:xmpp:forward:0 forwarded"`
	Stanza  Packet
}

// UnmarshalXML is a custom unmarshal function used by xml.Unmarshal to
// transform generic XML content into hierarchical Node structure.
func (f *Forwarded) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Check sub elements to extract required field as boolean
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch tt := t.(type) {

		case xml.StartElement:
			if packet, err := decodeClient(d, tt); err == nil {
				f.Stanza = packet
			}

		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}
	}
}
