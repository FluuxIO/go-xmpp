package xmpp

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
)

const xmppStreamOpen = "<?xml version='1.0'?><stream:stream to='%s' xmlns='%s' xmlns:stream='%s' version='1.0'>"

type Session struct {
	// Session info
	BindJid      string // Jabber ID as provided by XMPP server
	StreamId     string
	Features     streamFeatures
	TlsEnabled   bool
	lastPacketId int

	// Session interface
	In  chan interface{}
	Out chan interface{}

	// read / write
	socketProxy io.ReadWriter
	decoder     *xml.Decoder

	// error management
	err error
}

func NewSession(conn net.Conn, o Options) (net.Conn, *Session, error) {
	s := new(Session)
	s.init(conn, o)

	// starttls
	var tlsConn net.Conn
	tlsConn = s.startTlsIfSupported(conn, o.parsedJid.domain)
	s.reset(conn, tlsConn, o)

	// auth
	s.auth(o)
	s.reset(tlsConn, tlsConn, o)

	// bind resource and 'start' XMPP session
	s.bind(o)
	s.rfc3921Session(o)

	return tlsConn, s, s.err
}

func (s *Session) PacketId() string {
	s.lastPacketId++
	return fmt.Sprintf("%x", s.lastPacketId)
}

func (s *Session) init(conn net.Conn, o Options) {
	s.setProxy(nil, conn, o)
	s.Features = s.open(o.parsedJid.domain)
}

func (s *Session) reset(conn net.Conn, newConn net.Conn, o Options) {
	if s.err != nil {
		return
	}

	s.setProxy(conn, newConn, o)
	s.Features = s.open(o.parsedJid.domain)
}

// TODO: setProxyLogger ? better name ? This is not a TCP / HTTP proxy
func (s *Session) setProxy(conn net.Conn, newConn net.Conn, o Options) {
	if newConn != conn {
		s.socketProxy = newSocketProxy(newConn, o.PacketLogger)
	}
	s.decoder = xml.NewDecoder(s.socketProxy)
}

func (s *Session) open(domain string) (f streamFeatures) {
	// Send stream open tag
	if _, s.err = fmt.Fprintf(s.socketProxy, xmppStreamOpen, domain, nsClient, nsStream); s.err != nil {
		return
	}

	// Set xml decoder and extract streamID from reply
	s.StreamId, s.err = initDecoder(s.decoder) // TODO refactor / rename
	if s.err != nil {
		return
	}

	// extract stream features
	if s.err = s.decoder.Decode(&f); s.err != nil {
		s.err = errors.New("stream open decode features: " + s.err.Error())
	}
	return
}

func (s *Session) startTlsIfSupported(conn net.Conn, domain string) net.Conn {
	if s.err != nil {
		return conn
	}

	if s.Features.StartTLS.XMLName.Space+" "+s.Features.StartTLS.XMLName.Local == nsTLS+" starttls" {
		fmt.Fprintf(s.socketProxy, "<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>")

		var k tlsProceed
		if s.err = s.decoder.DecodeElement(&k, nil); s.err != nil {
			s.err = errors.New("expecting starttls proceed: " + s.err.Error())
			return conn
		}
		s.TlsEnabled = true

		// TODO: add option to accept all TLS certificates: insecureSkipTlsVerify (DefaultTlsConfig.InsecureSkipVerify)
		DefaultTlsConfig.ServerName = domain
		var tlsConn *tls.Conn = tls.Client(conn, &DefaultTlsConfig)
		// We convert existing connection to TLS
		if s.err = tlsConn.Handshake(); s.err != nil {
			return tlsConn
		}

		// We check that cert matches hostname
		s.err = tlsConn.VerifyHostname(domain)
		return tlsConn
	}

	// starttls is not supported => we do not upgrade the connection:
	return conn
}

func (s *Session) auth(o Options) {
	if s.err != nil {
		return
	}

	s.err = authSASL(s.socketProxy, s.decoder, s.Features, o.parsedJid.username, o.Password)
}

func (s *Session) bind(o Options) {
	if s.err != nil {
		return
	}

	// Send IQ message asking to bind to the local user name.
	var resource = o.parsedJid.resource
	if resource != "" {
		fmt.Fprintf(s.socketProxy, "<iq type='set' id='%s'><bind xmlns='%s'><resource>%s</resource></bind></iq>",
			s.PacketId(), nsBind, resource)
	} else {
		fmt.Fprintf(s.socketProxy, "<iq type='set' id='%s'><bind xmlns='%s'/></iq>", s.PacketId(), nsBind)
	}

	var iq clientIQ
	if s.err = s.decoder.Decode(&iq); s.err != nil || &iq.Bind == nil {
		s.err = errors.New("iq bind result missing: " + s.err.Error())
		return
	}
	s.BindJid = iq.Bind.Jid // our local id (with possibly randomly generated resource
	return
}

// TODO: remove when ejabberd is fixed: https://github.com/processone/ejabberd/issues/869
// After the bind, if the session is required (as per old RFC 3921), we send the session open iq
func (s *Session) rfc3921Session(o Options) {
	if s.err != nil {
		return
	}

	var iq clientIQ

	// TODO: Do no send unconditionally, check if session is optional and omit it
	fmt.Fprintf(s.socketProxy, "<iq type='set' id='%s'><session xmlns='%s'/></iq>", s.PacketId(), nsSession)
	if s.err = s.decoder.Decode(&iq); s.err != nil {
		s.err = errors.New("expecting iq result after session open: " + s.err.Error())
		return
	}
}
