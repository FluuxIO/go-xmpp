package xmpp

import (
	"errors"
	"fmt"

	"gosrc.io/xmpp/stanza"
)

type Session struct {
	// Session info
	BindJid      string // Jabber ID as provided by XMPP server
	StreamId     string
	SMState      SMState
	Features     stanza.StreamFeatures
	TlsEnabled   bool
	lastPacketId int

	// read / write
	transport Transport

	// error management
	err error
}

func NewSession(transport Transport, o Config, state SMState) (*Session, error) {
	s := new(Session)
	s.transport = transport
	s.SMState = state
	s.init(o)

	if s.err != nil {
		return nil, NewConnError(s.err, true)
	}

	if !transport.IsSecure() {
		s.startTlsIfSupported(o)
	}

	if !transport.IsSecure() && !o.Insecure {
		err := fmt.Errorf("failed to negotiate TLS session : %s", s.err)
		return nil, NewConnError(err, true)
	}

	if s.TlsEnabled {
		s.reset(o)
	}

	// auth
	s.auth(o)
	s.reset(o)

	// attempt resumption
	if s.resume(o) {
		return s, s.err
	}

	// otherwise, bind resource and 'start' XMPP session
	s.bind(o)
	s.rfc3921Session(o)

	// Enable stream management if supported
	s.EnableStreamManagement(o)

	return s, s.err
}

func (s *Session) PacketId() string {
	s.lastPacketId++
	return fmt.Sprintf("%x", s.lastPacketId)
}

func (s *Session) init(o Config) {
	s.Features = s.open(o.parsedJid.Domain)
}

func (s *Session) reset(o Config) {
	if s.StreamId, s.err = s.transport.StartStream(); s.err != nil {
		return
	}

	s.Features = s.open(o.parsedJid.Domain)
}

func (s *Session) open(domain string) (f stanza.StreamFeatures) {
	// extract stream features
	if s.err = s.transport.GetDecoder().Decode(&f); s.err != nil {
		s.err = errors.New("stream open decode features: " + s.err.Error())
	}
	return
}

func (s *Session) startTlsIfSupported(o Config) {
	if s.err != nil {
		return
	}

	if !s.transport.DoesStartTLS() {
		if !o.Insecure {
			s.err = errors.New("transport does not support starttls")
		}
		return
	}

	if _, ok := s.Features.DoesStartTLS(); ok {
		fmt.Fprintf(s.transport, "<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>")

		var k stanza.TLSProceed
		if s.err = s.transport.GetDecoder().DecodeElement(&k, nil); s.err != nil {
			s.err = errors.New("expecting starttls proceed: " + s.err.Error())
			return
		}

		s.err = s.transport.StartTLS()

		if s.err == nil {
			s.TlsEnabled = true
		}
		return
	}

	// If we do not allow cleartext serverConnections, make it explicit that server do not support starttls
	if !o.Insecure {
		s.err = errors.New("XMPP server does not advertise support for starttls")
	}
}

func (s *Session) auth(o Config) {
	if s.err != nil {
		return
	}

	s.err = authSASL(s.transport, s.transport.GetDecoder(), s.Features, o.parsedJid.Node, o.Credential)
}

// Attempt to resume session using stream management
func (s *Session) resume(o Config) bool {
	if !s.Features.DoesStreamManagement() {
		return false
	}
	if s.SMState.Id == "" {
		return false
	}

	fmt.Fprintf(s.transport, "<resume xmlns='%s' h='%d' previd='%s'/>",
		stanza.NSStreamManagement, s.SMState.Inbound, s.SMState.Id)

	var packet stanza.Packet
	packet, s.err = stanza.NextPacket(s.transport.GetDecoder())
	if s.err == nil {
		switch p := packet.(type) {
		case stanza.SMResumed:
			if p.PrevId != s.SMState.Id {
				s.err = errors.New("session resumption: mismatched id")
				s.SMState = SMState{}
				return false
			}
			return true
		case stanza.SMFailed:
		default:
			s.err = errors.New("unexpected reply to SM resume")
		}
	}
	s.SMState = SMState{}
	return false
}

func (s *Session) bind(o Config) {
	if s.err != nil {
		return
	}

	// Send IQ message asking to bind to the local user name.
	var resource = o.parsedJid.Resource
	if resource != "" {
		fmt.Fprintf(s.transport, "<iq type='set' id='%s'><bind xmlns='%s'><resource>%s</resource></bind></iq>",
			s.PacketId(), stanza.NSBind, resource)
	} else {
		fmt.Fprintf(s.transport, "<iq type='set' id='%s'><bind xmlns='%s'/></iq>", s.PacketId(), stanza.NSBind)
	}

	var iq stanza.IQ
	if s.err = s.transport.GetDecoder().Decode(&iq); s.err != nil {
		s.err = errors.New("error decoding iq bind result: " + s.err.Error())
		return
	}

	// TODO Check all elements
	switch payload := iq.Payload.(type) {
	case *stanza.Bind:
		s.BindJid = payload.Jid // our local id (with possibly randomly generated resource
	default:
		s.err = errors.New("iq bind result missing")
	}

	return
}

// After the bind, if the session is not optional (as per old RFC 3921), we send the session open iq.
func (s *Session) rfc3921Session(o Config) {
	if s.err != nil {
		return
	}

	var iq stanza.IQ
	// We only negotiate session binding if it is mandatory, we skip it when optional.
	if !s.Features.Session.IsOptional() {
		fmt.Fprintf(s.transport, "<iq type='set' id='%s'><session xmlns='%s'/></iq>", s.PacketId(), stanza.NSSession)
		if s.err = s.transport.GetDecoder().Decode(&iq); s.err != nil {
			s.err = errors.New("expecting iq result after session open: " + s.err.Error())
			return
		}
	}
}

// Enable stream management, with session resumption, if supported.
func (s *Session) EnableStreamManagement(o Config) {
	if s.err != nil {
		return
	}
	if !s.Features.DoesStreamManagement() {
		return
	}

	fmt.Fprintf(s.transport, "<enable xmlns='%s' resume='true'/>", stanza.NSStreamManagement)

	var packet stanza.Packet
	packet, s.err = stanza.NextPacket(s.transport.GetDecoder())
	if s.err == nil {
		switch p := packet.(type) {
		case stanza.SMEnabled:
			s.SMState = SMState{Id: p.Id}
		case stanza.SMFailed:
			// TODO: Store error in SMState, for later inspection
		default:
			s.err = errors.New("unexpected reply to SM enable")
		}
	}

	return
}
