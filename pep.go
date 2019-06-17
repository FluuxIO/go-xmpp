package xmpp // import "gosrc.io/xmpp"

// TODO: Move to a pubsub file

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

type Mood struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/mood mood"`
	// TODO: Custom parsing to extract mood type from tag name
	// Mood type
	Text string `xml:"text,omitempty"`
}
