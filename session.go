package xmpp

import (
	"encoding/xml"
	"errors"
	"fmt"
	"gosrc.io/xmpp/stanza"
	"strconv"
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

func NewSession(c *Client, state SMState) (*Session, error) {
	var s *Session
	if c.Session == nil {
		s = new(Session)
		s.transport = c.transport
		s.SMState = state
		s.init()
	} else {
		s = c.Session
		// We keep information about the previously set session, like the session ID, but we read server provided
		// info again in case it changed between session break and resume, such as features.
		s.init()
	}

	if s.err != nil {
		return nil, NewConnError(s.err, true)
	}

	if !c.transport.IsSecure() {
		s.startTlsIfSupported(c.config)
	}

	if !c.transport.IsSecure() && !c.config.Insecure {
		err := fmt.Errorf("failed to negotiate TLS session : %s", s.err)
		return nil, NewConnError(err, true)
	}

	if s.TlsEnabled {
		s.reset()
	}

	// auth
	s.auth(c.config)
	if s.err != nil {
		return s, s.err
	}
	s.reset()
	if s.err != nil {
		return s, s.err
	}

	// attempt resumption
	if s.resume(c.config) {
		return s, s.err
	}

	// otherwise, bind resource and 'start' XMPP session
	s.bind(c.config)
	if s.err != nil {
		return s, s.err
	}
	s.rfc3921Session()
	if s.err != nil {
		return s, s.err
	}

	// Enable stream management if supported
	s.EnableStreamManagement(c.config)
	if s.err != nil {
		return s, s.err
	}

	return s, s.err
}

func (s *Session) PacketId() string {
	s.lastPacketId++
	return fmt.Sprintf("%x", s.lastPacketId)
}

// init gathers information on the session such as stream features from the server.
func (s *Session) init() {
	s.Features = s.extractStreamFeatures()
}

func (s *Session) reset() {
	if s.StreamId, s.err = s.transport.StartStream(); s.err != nil {
		return
	}

	s.Features = s.extractStreamFeatures()
}

func (s *Session) extractStreamFeatures() (f stanza.StreamFeatures) {
	// extract stream features
	if s.err = s.transport.GetDecoder().Decode(&f); s.err != nil {
		s.err = errors.New("stream open decode features: " + s.err.Error())
	}
	return
}

func (s *Session) startTlsIfSupported(o *Config) {
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

func (s *Session) auth(o *Config) {
	if s.err != nil {
		return
	}

	s.err = authSASL(s.transport, s.transport.GetDecoder(), s.Features, o.parsedJid.Node, o.Credential)
}

// Attempt to resume session using stream management
func (s *Session) resume(o *Config) bool {
	if !s.Features.DoesStreamManagement() {
		return false
	}
	if s.SMState.Id == "" {
		return false
	}

	rsm := stanza.SMResume{
		PrevId: s.SMState.Id,
		H:      &s.SMState.Inbound,
	}
	data, err := xml.Marshal(rsm)

	_, err = s.transport.Write(data)
	if err != nil {
		return false
	}
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

func (s *Session) bind(o *Config) {
	if s.err != nil {
		return
	}

	// Send IQ message asking to bind to the local user name.
	var resource = o.parsedJid.Resource
	iqB, err := stanza.NewIQ(stanza.Attrs{
		Type: stanza.IQTypeSet,
		Id:   s.PacketId(),
	})
	if err != nil {
		s.err = err
		return
	}

	// Check if we already have a resource name, and include it in the request if so
	if resource != "" {
		iqB.Payload = &stanza.Bind{
			Resource: resource,
		}
	} else {
		iqB.Payload = &stanza.Bind{}

	}

	// Send the bind request IQ
	data, err := xml.Marshal(iqB)
	if err != nil {
		s.err = err
		return
	}
	n, err := s.transport.Write(data)
	if err != nil {
		s.err = err
		return
	} else if n == 0 {
		s.err = errors.New("failed to write bind iq stanza to the server : wrote 0 bytes")
		return
	}

	// Check the server response
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
func (s *Session) rfc3921Session() {
	if s.err != nil {
		return
	}

	var iq stanza.IQ
	// We only negotiate session binding if it is mandatory, we skip it when optional.
	if !s.Features.Session.IsOptional() {
		se, err := stanza.NewIQ(stanza.Attrs{
			Type: stanza.IQTypeSet,
			Id:   s.PacketId(),
		})
		if err != nil {
			s.err = err
			return
		}
		se.Payload = &stanza.StreamSession{}
		data, err := xml.Marshal(se)
		if err != nil {
			s.err = err
			return
		}
		n, err := s.transport.Write(data)
		if err != nil {
			s.err = err
			return
		} else if n == 0 {
			s.err = errors.New("there was a problem marshaling the session IQ : wrote 0 bytes to server")
			return
		}

		if s.err = s.transport.GetDecoder().Decode(&iq); s.err != nil {
			s.err = errors.New("expecting iq result after session open: " + s.err.Error())
			return
		}
	}
}

// Enable stream management, with session resumption, if supported.
func (s *Session) EnableStreamManagement(o *Config) {
	if s.err != nil {
		return
	}
	if !s.Features.DoesStreamManagement() || !o.StreamManagementEnable {
		return
	}
	q := stanza.NewUnAckQueue()
	ebleNonza := stanza.SMEnable{Resume: &o.streamManagementResume}
	pktStr, err := xml.Marshal(ebleNonza)
	if err != nil {
		s.err = err
		return
	}
	_, err = s.transport.Write(pktStr)
	if err != nil {
		s.err = err
		return
	}

	var packet stanza.Packet
	packet, s.err = stanza.NextPacket(s.transport.GetDecoder())
	if s.err == nil {
		switch p := packet.(type) {
		case stanza.SMEnabled:
			// Server allows resumption or not using SMEnabled attribute "resume". We must read the server response
			// and update config accordingly
			b, err := strconv.ParseBool(p.Resume)
			if err != nil || !b {
				o.StreamManagementEnable = false
			}
			s.SMState = SMState{Id: p.Id, preferredReconAddr: p.Location}
			s.SMState.UnAckQueue = q
		case stanza.SMFailed:
			// TODO: Store error in SMState, for later inspection
			s.SMState = SMState{StreamErrorGroup: p.StreamErrorGroup}
			s.SMState.UnAckQueue = q
			s.err = errors.New("failed to establish session : " + s.SMState.StreamErrorGroup.GroupErrorName())
		default:
			s.err = errors.New("unexpected reply to SM enable")
		}
	}
	return
}
