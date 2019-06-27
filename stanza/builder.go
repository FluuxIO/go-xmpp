package stanza

import (
	"encoding/xml"
)

type builder struct{ lang string }

// NewBuilder create a builder structure. It act as an interface for packet generation.
// The goal is to work well with code completion to more easily.
//
// Using the builder to format and create packets is optional. You can always prepare
// your packet dealing with the struct manually and initializing them with the right values.
func NewBuilder() *builder {
	return &builder{}
}

// Set default language
func (b *builder) Lang(lang string) *builder {
	b.lang = lang
	return b
}

func (b *builder) IQ(a Attrs) IQ {
	return IQ{
		XMLName: xml.Name{Local: "iq"},
		Attrs:   a,
	}
}

func (b *builder) Message(a Attrs) Message {
	return Message{
		XMLName: xml.Name{Local: "message"},
		Attrs:   a,
	}
}

func (b *builder) Presence(a Attrs) Presence {
	return Presence{
		XMLName: xml.Name{Local: "presence"},
		Attrs:   a,
	}
}

// ======================================================================================
// IQ payloads

// DiscoInfo builds a default DiscoInfo payload
func (*builder) DiscoInfo() *DiscoInfo {
	d := DiscoInfo{
		XMLName: xml.Name{
			Space: NSDiscoInfo,
			Local: "query",
		},
	}
	return &d
}

// Identity builds a identity struct for use in Disco
func (*builder) Identity(name, category, typ string) *Identity {
	return &Identity{}
}
