package stanza

import (
	"encoding/xml"
	"errors"
)

const (
	NSStreamManagement = "urn:xmpp:sm:3"
)

// Enabled as defined in Stream Management spec
// Reference: https://xmpp.org/extensions/xep-0198.html#enable
type SMEnabled struct {
	XMLName  xml.Name `xml:"urn:xmpp:sm:3 enabled"`
	Id       string   `xml:"id,attr,omitempty"`
	Location string   `xml:"location,attr,omitempty"`
	Resume   string   `xml:"resume,attr,omitempty"`
	Max      uint     `xml:"max,attr,omitempty"`
}

func (SMEnabled) Name() string {
	return "Stream Management: enabled"
}

// Request as defined in Stream Management spec
// Reference: https://xmpp.org/extensions/xep-0198.html#acking
type SMRequest struct {
	XMLName xml.Name `xml:"urn:xmpp:sm:3 r"`
}

func (SMRequest) Name() string {
	return "Stream Management: request"
}

// Answer as defined in Stream Management spec
// Reference: https://xmpp.org/extensions/xep-0198.html#acking
type SMAnswer struct {
	XMLName xml.Name `xml:"urn:xmpp:sm:3 a"`
	H       uint     `xml:"h,attr,omitempty"`
}

func (SMAnswer) Name() string {
	return "Stream Management: answer"
}

// Resumed as defined in Stream Management spec
// Reference: https://xmpp.org/extensions/xep-0198.html#acking
type SMResumed struct {
	XMLName xml.Name `xml:"urn:xmpp:sm:3 resumed"`
	PrevId  string   `xml:"previd,attr,omitempty"`
	H       uint     `xml:"h,attr,omitempty"`
}

func (SMResumed) Name() string {
	return "Stream Management: resumed"
}

// Failed as defined in Stream Management spec
// Reference: https://xmpp.org/extensions/xep-0198.html#acking
type SMFailed struct {
	XMLName xml.Name `xml:"urn:xmpp:sm:3 failed"`
	// TODO: Handle decoding error cause (need custom parsing).
}

func (SMFailed) Name() string {
	return "Stream Management: failed"
}

type smDecoder struct{}

var sm smDecoder

// decode decodes all known nonza in the stream management namespace.
func (s smDecoder) decode(p *xml.Decoder, se xml.StartElement) (Packet, error) {
	switch se.Name.Local {
	case "enabled":
		return s.decodeEnabled(p, se)
	case "resumed":
		return s.decodeResumed(p, se)
	case "r":
		return s.decodeRequest(p, se)
	case "h":
		return s.decodeAnswer(p, se)
	case "failed":
		return s.decodeFailed(p, se)
	default:
		return nil, errors.New("unexpected XMPP packet " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}
}

func (smDecoder) decodeEnabled(p *xml.Decoder, se xml.StartElement) (SMEnabled, error) {
	var packet SMEnabled
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

func (smDecoder) decodeResumed(p *xml.Decoder, se xml.StartElement) (SMResumed, error) {
	var packet SMResumed
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

func (smDecoder) decodeRequest(p *xml.Decoder, se xml.StartElement) (SMRequest, error) {
	var packet SMRequest
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

func (smDecoder) decodeAnswer(p *xml.Decoder, se xml.StartElement) (SMAnswer, error) {
	var packet SMAnswer
	err := p.DecodeElement(&packet, &se)
	return packet, err
}

func (smDecoder) decodeFailed(p *xml.Decoder, se xml.StartElement) (SMFailed, error) {
	var packet SMFailed
	err := p.DecodeElement(&packet, &se)
	return packet, err
}
