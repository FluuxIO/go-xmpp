package stanza

import (
	"encoding/xml"
)

type Tune struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/tune tune"`
	Artist  string   `xml:"artist,omitempty"`
	Length  int      `xml:"length,omitempty"`
	Rating  int      `xml:"rating,omitempty"`
	Source  string   `xml:"source,omitempty"`
	Title   string   `xml:"title,omitempty"`
	Track   string   `xml:"track,omitempty"`
	Uri     string   `xml:"uri,omitempty"`
}

// Mood defines deta model for XEP-0107 - User Mood
// See: https://xmpp.org/extensions/xep-0107.html
type Mood struct {
	MsgExtension          // Mood can be added as a message extension
	XMLName      xml.Name `xml:"http://jabber.org/protocol/mood mood"`
	// TODO: Custom parsing to extract mood type from tag name.
	// Note: the list is predefined.
	// Mood type
	Text string `xml:"text,omitempty"`
}
