package xmpp

import (
	"crypto/tls"
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
	return err
}

func (t XMPPTransport) DoesStartTLS() bool {
	return true
}

func (t XMPPTransport) IsSecure() bool {
	return t.isSecure
}

func (t *XMPPTransport) StartTLS(domain string) error {
	if t.Config.TLSConfig == nil {
		t.Config.TLSConfig = &tls.Config{}
	}

	if t.Config.TLSConfig.ServerName == "" {
		t.Config.TLSConfig.ServerName = domain
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

func (t XMPPTransport) Read(p []byte) (n int, err error) {
	return t.conn.Read(p)
}

func (t XMPPTransport) Write(p []byte) (n int, err error) {
	return t.conn.Write(p)
}

func (t XMPPTransport) Close() error {
	return t.conn.Close()
}
