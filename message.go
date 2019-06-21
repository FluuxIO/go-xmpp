package xmpp

import (
	"encoding/xml"
)

// ============================================================================
// Message Packet

// Message implements RFC 6120 - A.5 Client Namespace (a part)
type Message struct {
	XMLName xml.Name `xml:"message"`
	Attrs

	Subject    string         `xml:"subject,omitempty"`
	Body       string         `xml:"body,omitempty"`
	Thread     string         `xml:"thread,omitempty"`
	Error      Err            `xml:"error,omitempty"`
	Extensions []MsgExtension `xml:",omitempty"`
}

func (Message) Name() string {
	return "message"
}

func NewMessage(a Attrs) Message {
	return Message{
		XMLName: xml.Name{Local: "message"},
		Attrs:   a,
	}
}

type messageDecoder struct{}

var message messageDecoder

func (messageDecoder) decode(p *xml.Decoder, se xml.StartElement) (Message, error) {
	var packet Message
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

// XMPPFormat with all Extensions
func (msg *Message) XMPPFormat() string {
	out, err := xml.MarshalIndent(msg, "", "")
	if err != nil {
		return ""
	}
	return string(out)
}

// UnmarshalXML implements custom parsing for IQs
func (msg *Message) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	msg.XMLName = start.Name

	// Extract packet attributes
	for _, attr := range start.Attr {
		if attr.Name.Local == "id" {
			msg.Id = attr.Value
		}
		if attr.Name.Local == "type" {
			msg.Type = StanzaType(attr.Value)
		}
		if attr.Name.Local == "to" {
			msg.To = attr.Value
		}
		if attr.Name.Local == "from" {
			msg.From = attr.Value
		}
		if attr.Name.Local == "lang" {
			msg.Lang = attr.Value
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
			if msgExt := TypeRegistry.GetMsgExtension(tt.Name); msgExt != nil {
				// Decode message extension
				err = d.DecodeElement(msgExt, &tt)
				if err != nil {
					return err
				}
				msg.Extensions = append(msg.Extensions, msgExt)
			} else {
				// Decode standard message sub-elements
				var err error
				switch tt.Name.Local {
				case "body":
					err = d.DecodeElement(&msg.Body, &tt)
				case "thread":
					err = d.DecodeElement(&msg.Thread, &tt)
				case "subject":
					err = d.DecodeElement(&msg.Subject, &tt)
				case "error":
					err = d.DecodeElement(&msg.Error, &tt)
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
