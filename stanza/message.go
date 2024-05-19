package stanza

import (
	"encoding/xml"
	"reflect"
)

// ============================================================================
// Message Packet

// LocalizedString is a string node with a language attribute.
type LocalizedString struct {
	Content string `xml:",chardata"`
	Lang    string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
}

// Message implements RFC 6120 - A.5 Client Namespace (a part)
type Message struct {
	XMLName xml.Name `xml:"message"`
	Attrs

	Subject    []LocalizedString `xml:"subject,omitempty"`
	Body       []LocalizedString `xml:"body,omitempty"`
	Thread     string            `xml:"thread,omitempty"`
	Error      Err               `xml:"error,omitempty"`
	Extensions []MsgExtension    `xml:",omitempty"`
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

// Get search and extracts a specific extension on a message.
// It receives a pointer to an MsgExtension. It will panic if the caller
// does not pass a pointer.
// It will return true if the passed extension is found and set the pointer
// to the extension passed as parameter to the found extension.
// It will return false if the extension is not found on the message.
//
// Example usage:
//   var oob xmpp.OOB
//   if ok := msg.Get(&oob); ok {
//     // oob extension has been found
//	 }
func (msg *Message) Get(ext MsgExtension) bool {
	target := reflect.ValueOf(ext)
	if target.Kind() != reflect.Ptr {
		panic("you must pass a pointer to the message Get method")
	}

	for _, e := range msg.Extensions {
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

// UnmarshalXML implements custom parsing for messages
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
