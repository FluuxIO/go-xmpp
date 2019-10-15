package xmpp

import (
	"crypto/tls"
	"errors"
	"net"
	"time"
)

// XMPPTransport implements the XMPP native TCP transport
type XMPPTransport struct {
	Config    TransportConfiguration
	TLSConfig *tls.Config
	// TCP level connection / can be replaced by a TLS session after starttls
	conn     net.Conn
	isSecure bool
}

func (t *XMPPTransport) Connect() error {
	var err error

	t.conn, err = net.DialTimeout("tcp", t.Config.Address, time.Duration(t.Config.ConnectTimeout)*time.Second)
	if err != nil {
		return NewConnError(err, true)
	}
	return nil
}

func (t XMPPTransport) DoesStartTLS() bool {
	return true
}

func (t XMPPTransport) IsSecure() bool {
	return t.isSecure
}

func (t *XMPPTransport) StartTLS(domain string) error {
	if t.Config.TLSConfig == nil {
		t.TLSConfig = &tls.Config{}
	} else {
		t.TLSConfig = t.Config.TLSConfig.Clone()
	}

	if t.TLSConfig.ServerName == "" {
		t.TLSConfig.ServerName = domain
	}
	tlsConn := tls.Client(t.conn, t.TLSConfig)
	// We convert existing connection to TLS
	if err := tlsConn.Handshake(); err != nil {
		return err
	}

	if !t.TLSConfig.InsecureSkipVerify {
		if err := tlsConn.VerifyHostname(domain); err != nil {
			return err
		}
	}

	t.isSecure = true
	return nil
}

func (t XMPPTransport) Ping() error {
	n, err := t.conn.Write([]byte("\n"))
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("Could not write ping")
	}
	return nil
}

func (t XMPPTransport) Read(p []byte) (n int, err error) {
	return t.conn.Read(p)
}

func (t XMPPTransport) Write(p []byte) (n int, err error) {
	return t.conn.Write(p)
}

func (t XMPPTransport) Close() error {
	return t.conn.Close()
}
