package xmpp

import (
	"crypto/tls"
	"net"
	"time"
)

type Transport interface {
	Connect(address string, c Config) error
	DoesStartTLS() bool
	StartTLS(domain string, c Config) error

	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}

// XMPPTransport implements the XMPP native TCP transport
type XMPPTransport struct {
	TLSConfig *tls.Config
	// TCP level connection / can be replaced by a TLS session after starttls
	conn net.Conn
}

func (t *XMPPTransport) Connect(address string, c Config) error {
	var err error

	t.conn, err = net.DialTimeout("tcp", address, time.Duration(c.ConnectTimeout)*time.Second)
	return err
}

func (t XMPPTransport) DoesStartTLS() bool {
	return true
}

func (t *XMPPTransport) StartTLS(domain string, c Config) error {
	if t.TLSConfig == nil {
		if c.TLSConfig != nil {
			t.TLSConfig = c.TLSConfig
		} else {
			t.TLSConfig = &tls.Config{}
		}
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
