package stanza

import (
	"encoding/xml"
	"errors"
	"sync"
)

const (
	NSStreamManagement = "urn:xmpp:sm:3"
)

type SMEnable struct {
	XMLName xml.Name `xml:"urn:xmpp:sm:3 enable"`
	Max     *uint    `xml:"max,attr,omitempty"`
	Resume  *bool    `xml:"resume,attr,omitempty"`
}

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

type UnAckQueue struct {
	Uslice []*UnAckedStz
	sync.RWMutex
}
type UnAckedStz struct {
	Id  int
	Stz string
}

func NewUnAckQueue() *UnAckQueue {
	return &UnAckQueue{
		Uslice:  make([]*UnAckedStz, 0, 10), // Capacity is 0 to comply with "Push" implementation (so that no reachable element is nil)
		RWMutex: sync.RWMutex{},
	}
}

func (u *UnAckedStz) QueueableName() string {
	return "Un-acknowledged stanza"
}

func (uaq *UnAckQueue) PeekN(n int) []Queueable {
	if uaq == nil {
		return nil
	}
	if n <= 0 {
		return nil
	}
	if len(uaq.Uslice) < n {
		n = len(uaq.Uslice)
	}

	if len(uaq.Uslice) == 0 {
		return nil
	}
	var r []Queueable
	for i := 0; i < n; i++ {
		r = append(r, uaq.Uslice[i])
	}
	return r
}

// No guarantee regarding thread safety !
func (uaq *UnAckQueue) Pop() Queueable {
	if uaq == nil {
		return nil
	}
	r := uaq.Peek()
	if r != nil {
		uaq.Uslice = uaq.Uslice[1:]
	}
	return r
}

// No guarantee regarding thread safety !
func (uaq *UnAckQueue) PopN(n int) []Queueable {
	if uaq == nil {
		return nil
	}
	r := uaq.PeekN(n)
	uaq.Uslice = uaq.Uslice[len(r):]
	return r
}

func (uaq *UnAckQueue) Peek() Queueable {
	if uaq == nil {
		return nil
	}
	if len(uaq.Uslice) == 0 {
		return nil
	}
	r := uaq.Uslice[0]
	return r
}

func (uaq *UnAckQueue) Push(s Queueable) error {
	if uaq == nil {
		return nil
	}
	pushIdx := 1
	if len(uaq.Uslice) != 0 {
		pushIdx = uaq.Uslice[len(uaq.Uslice)-1].Id + 1
	}

	sStz, ok := s.(*UnAckedStz)
	if !ok {
		return errors.New("element in not compatible with this queue. expected an UnAckedStz")
	}

	e := UnAckedStz{
		Id:  pushIdx,
		Stz: sStz.Stz,
	}

	uaq.Uslice = append(uaq.Uslice, &e)

	return nil
}

func (uaq *UnAckQueue) Empty() bool {
	if uaq == nil {
		return true
	}
	r := len(uaq.Uslice)
	return r == 0
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
	H       uint     `xml:"h,attr"`
}

func (SMAnswer) Name() string {
	return "Stream Management: answer"
}

// Resumed as defined in Stream Management spec
// Reference: https://xmpp.org/extensions/xep-0198.html#acking
type SMResumed struct {
	XMLName xml.Name `xml:"urn:xmpp:sm:3 resumed"`
	PrevId  string   `xml:"previd,attr,omitempty"`
	H       *uint    `xml:"h,attr,omitempty"`
}

func (SMResumed) Name() string {
	return "Stream Management: resumed"
}

// Resume as defined in Stream Management spec
// Reference: https://xmpp.org/extensions/xep-0198.html#acking
type SMResume struct {
	XMLName xml.Name `xml:"urn:xmpp:sm:3 resume"`
	PrevId  string   `xml:"previd,attr,omitempty"`
	H       *uint    `xml:"h,attr,omitempty"`
}

func (SMResume) Name() string {
	return "Stream Management: resume"
}

// Failed as defined in Stream Management spec
// Reference: https://xmpp.org/extensions/xep-0198.html#acking
type SMFailed struct {
	XMLName xml.Name `xml:"urn:xmpp:sm:3 failed"`
	H       *uint    `xml:"h,attr,omitempty"`

	StreamErrorGroup StanzaErrorGroup
}

func (SMFailed) Name() string {
	return "Stream Management: failed"
}

func (smf *SMFailed) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	smf.XMLName = start.Name

	// According to https://xmpp.org/rfcs/rfc3920.html#def we should have no attributes aside from the namespace
	// which we don't use internally

	// decode inner elements
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch tt := t.(type) {

		case xml.StartElement:
			// Decode sub-elements
			var err error
			switch tt.Name.Local {
			case "bad-format":
				bf := BadFormat{}
				err = d.DecodeElement(&bf, &tt)
				smf.StreamErrorGroup = &bf
			case "bad-namespace-prefix":
				bnp := BadNamespacePrefix{}
				err = d.DecodeElement(&bnp, &tt)
				smf.StreamErrorGroup = &bnp
			case "conflict":
				c := Conflict{}
				err = d.DecodeElement(&c, &tt)
				smf.StreamErrorGroup = &c
			case "connection-timeout":
				ct := ConnectionTimeout{}
				err = d.DecodeElement(&ct, &tt)
				smf.StreamErrorGroup = &ct
			case "host-gone":
				hg := HostGone{}
				err = d.DecodeElement(&hg, &tt)
				smf.StreamErrorGroup = &hg
			case "host-unknown":
				hu := HostUnknown{}
				err = d.DecodeElement(&hu, &tt)
				smf.StreamErrorGroup = &hu
			case "improper-addressing":
				ia := ImproperAddressing{}
				err = d.DecodeElement(&ia, &tt)
				smf.StreamErrorGroup = &ia
			case "internal-server-error":
				ise := InternalServerError{}
				err = d.DecodeElement(&ise, &tt)
				smf.StreamErrorGroup = &ise
			case "invalid-from":
				ifrm := InvalidForm{}
				err = d.DecodeElement(&ifrm, &tt)
				smf.StreamErrorGroup = &ifrm
			case "invalid-id":
				id := InvalidId{}
				err = d.DecodeElement(&id, &tt)
				smf.StreamErrorGroup = &id
			case "invalid-namespace":
				ins := InvalidNamespace{}
				err = d.DecodeElement(&ins, &tt)
				smf.StreamErrorGroup = &ins
			case "invalid-xml":
				ix := InvalidXML{}
				err = d.DecodeElement(&ix, &tt)
				smf.StreamErrorGroup = &ix
			case "not-authorized":
				na := NotAuthorized{}
				err = d.DecodeElement(&na, &tt)
				smf.StreamErrorGroup = &na
			case "not-well-formed":
				nwf := NotWellFormed{}
				err = d.DecodeElement(&nwf, &tt)
				smf.StreamErrorGroup = &nwf
			case "policy-violation":
				pv := PolicyViolation{}
				err = d.DecodeElement(&pv, &tt)
				smf.StreamErrorGroup = &pv
			case "remote-connection-failed":
				rcf := RemoteConnectionFailed{}
				err = d.DecodeElement(&rcf, &tt)
				smf.StreamErrorGroup = &rcf
			case "resource-constraint":
				rc := ResourceConstraint{}
				err = d.DecodeElement(&rc, &tt)
				smf.StreamErrorGroup = &rc
			case "restricted-xml":
				rx := RestrictedXML{}
				err = d.DecodeElement(&rx, &tt)
				smf.StreamErrorGroup = &rx
			case "see-other-host":
				soh := SeeOtherHost{}
				err = d.DecodeElement(&soh, &tt)
				smf.StreamErrorGroup = &soh
			case "system-shutdown":
				ss := SystemShutdown{}
				err = d.DecodeElement(&ss, &tt)
				smf.StreamErrorGroup = &ss
			case "undefined-condition":
				uc := UndefinedCondition{}
				err = d.DecodeElement(&uc, &tt)
				smf.StreamErrorGroup = &uc
			case "unexpected-request":
				ur := UnexpectedRequest{}
				err = d.DecodeElement(&ur, &tt)
				smf.StreamErrorGroup = &ur
			case "unsupported-encoding":
				ue := UnsupportedEncoding{}
				err = d.DecodeElement(&ue, &tt)
				smf.StreamErrorGroup = &ue
			case "unsupported-stanza-type":
				ust := UnsupportedStanzaType{}
				err = d.DecodeElement(&ust, &tt)
				smf.StreamErrorGroup = &ust
			case "unsupported-version":
				uv := UnsupportedVersion{}
				err = d.DecodeElement(&uv, &tt)
				smf.StreamErrorGroup = &uv
			case "xml-not-well-formed":
				xnwf := XMLNotWellFormed{}
				err = d.DecodeElement(&xnwf, &tt)
				smf.StreamErrorGroup = &xnwf
			default:
				return errors.New("error is unknown")
			}
			if err != nil {
				return err
			}
		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}
	}
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
	case "resume":
		return s.decodeResume(p, se)
	case "r":
		return s.decodeRequest(p, se)
	case "a":
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

func (smDecoder) decodeResume(p *xml.Decoder, se xml.StartElement) (SMResume, error) {
	var packet SMResume
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
