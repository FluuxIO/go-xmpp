package xmpp

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"

	"gosrc.io/xmpp/stanza"
)

const xmppStreamOpen = "<?xml version='1.0'?><stream:stream to='%s' xmlns='%s' xmlns:stream='%s' version='1.0'>"

type Session struct {
	// Session info
	BindJid      string // Jabber ID as provided by XMPP server
	StreamId     string
	SMState      SMState
	Features     stanza.StreamFeatures
	TlsEnabled   bool
	lastPacketId int

	// read / write
	streamLogger io.ReadWriter
	decoder      *xml.Decoder

	// error management
	err error
}

func NewSession(conn net.Conn, o Config, state SMState) (net.Conn, *Session, error) {
	s := new(Session)
	s.SMState = state
	s.init(conn, o)

	// starttls
	var tlsConn net.Conn
	tlsConn = s.startTlsIfSupported(conn, o.parsedJid.Domain, o)

	if s.err != nil {
		return nil, nil, NewConnError(s.err, true)
	}

	if !s.TlsEnabled && !o.Insecure {
		err := fmt.Errorf("failed to negotiate TLS session : %s", s.err)
		return nil, nil, NewConnError(err, true)
	}

	if s.TlsEnabled {
		s.reset(conn, tlsConn, o)
	}

	// auth
	s.auth(o)
	s.reset(tlsConn, tlsConn, o)

	// attempt resumption
	if s.resume(o) {
		return tlsConn, s, s.err
	}

	// otherwise, bind resource and 'start' XMPP session
	s.bind(o)
	s.rfc3921Session(o)

	// Enable stream management if supported
	s.EnableStreamManagement(o)

	return tlsConn, s, s.err
}

func (s *Session) PacketId() string {
	s.lastPacketId++
	return fmt.Sprintf("%x", s.lastPacketId)
}

func (s *Session) init(conn net.Conn, o Config) {
	s.setStreamLogger(nil, conn, o)
	s.Features = s.open(o.parsedJid.Domain)
}

func (s *Session) reset(conn net.Conn, newConn net.Conn, o Config) {
	if s.err != nil {
		return
	}

	s.setStreamLogger(conn, newConn, o)
	s.Features = s.open(o.parsedJid.Domain)
}

func (s *Session) setStreamLogger(conn net.Conn, newConn net.Conn, o Config) {
	if newConn != conn {
		s.streamLogger = newStreamLogger(newConn, o.StreamLogger)
	}
	s.decoder = xml.NewDecoder(s.streamLogger)
	s.decoder.CharsetReader = o.CharsetReader
}

func (s *Session) open(domain string) (f stanza.StreamFeatures) {
	// Send stream open tag
	if _, s.err = fmt.Fprintf(s.streamLogger, xmppStreamOpen, domain, stanza.NSClient, stanza.NSStream); s.err != nil {
		return
	}

	// Set xml decoder and extract streamID from reply
	s.StreamId, s.err = stanza.InitStream(s.decoder) // TODO refactor / rename
	if s.err != nil {
		return
	}

	// extract stream features
	if s.err = s.decoder.Decode(&f); s.err != nil {
		s.err = errors.New("stream open decode features: " + s.err.Error())
	}
	return
}

func (s *Session) startTlsIfSupported(conn net.Conn, domain string, o Config) net.Conn {
	if s.err != nil {
		return conn
	}

	if _, ok := s.Features.DoesStartTLS(); ok {
		fmt.Fprintf(s.streamLogger, "<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>")

		var k stanza.TLSProceed
		if s.err = s.decoder.DecodeElement(&k, nil); s.err != nil {
			s.err = errors.New("expecting starttls proceed: " + s.err.Error())
			return conn
		}

		if o.TLSConfig == nil {
			o.TLSConfig = &tls.Config{}
		}

		if o.TLSConfig.ServerName == "" {
			o.TLSConfig.ServerName = domain
		}
		tlsConn := tls.Client(conn, o.TLSConfig)
		// We convert existing connection to TLS
		if s.err = tlsConn.Handshake(); s.err != nil {
			return tlsConn
		}

		if !o.TLSConfig.InsecureSkipVerify {
			s.err = tlsConn.VerifyHostname(domain)
		}

		if s.err == nil {
			s.TlsEnabled = true
		}
		return tlsConn
	}

	// If we do not allow cleartext connections, make it explicit that server do not support starttls
	if !o.Insecure {
		s.err = errors.New("XMPP server does not advertise support for starttls")
	}

	// starttls is not supported => we do not upgrade the connection:
	return conn
}

func (s *Session) auth(o Config) {
	if s.err != nil {
		return
	}

	s.err = authSASL(s.streamLogger, s.decoder, s.Features, o.parsedJid.Node, o.Password)
}

// Attempt to resume session using stream management
func (s *Session) resume(o Config) bool {
	if !s.Features.DoesStreamManagement() {
		return false
	}
	if s.SMState.Id == "" {
		return false
	}

	fmt.Fprintf(s.streamLogger, "<resume xmlns='%s' h='%d' previd='%s'/>",
		stanza.NSStreamManagement, s.SMState.Inbound, s.SMState.Id)

	var packet stanza.Packet
	packet, s.err = stanza.NextPacket(s.decoder)
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
			fmt.Println("MREMOND SM Failed")
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
		fmt.Fprintf(s.streamLogger, "<iq type='set' id='%s'><bind xmlns='%s'><resource>%s</resource></bind></iq>",
			s.PacketId(), stanza.NSBind, resource)
	} else {
		fmt.Fprintf(s.streamLogger, "<iq type='set' id='%s'><bind xmlns='%s'/></iq>", s.PacketId(), stanza.NSBind)
	}

	var iq stanza.IQ
	if s.err = s.decoder.Decode(&iq); s.err != nil {
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
		fmt.Fprintf(s.streamLogger, "<iq type='set' id='%s'><session xmlns='%s'/></iq>", s.PacketId(), stanza.NSSession)
		if s.err = s.decoder.Decode(&iq); s.err != nil {
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

	fmt.Fprintf(s.streamLogger, "<enable xmlns='%s' resume='true'/>", stanza.NSStreamManagement)

	var packet stanza.Packet
	packet, s.err = stanza.NextPacket(s.decoder)
	if s.err == nil {
		switch p := packet.(type) {
		case stanza.SMEnabled:
			s.SMState = SMState{Id: p.Id}
		case stanza.SMFailed:
			// TODO: Store error in SMState
		default:
			fmt.Println(p)
			s.err = errors.New("unexpected reply to SM enable")
		}
	}

	return
}
