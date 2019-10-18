package xmpp

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"gosrc.io/xmpp/stanza"
)

// XMPPTransport implements the XMPP native TCP transport
type XMPPTransport struct {
	Config     TransportConfiguration
	TLSConfig  *tls.Config
	decoder    *xml.Decoder
	conn       net.Conn
	readWriter io.ReadWriter
	isSecure   bool
}

const xmppStreamOpen = "<?xml version='1.0'?><stream:stream to='%s' xmlns='%s' xmlns:stream='%s' version='1.0'>"

func (t *XMPPTransport) Connect() (string, error) {
	var err error

	t.conn, err = net.DialTimeout("tcp", t.Config.Address, time.Duration(t.Config.ConnectTimeout)*time.Second)
	if err != nil {
		return "", NewConnError(err, true)
	}

	if _, err = fmt.Fprintf(t.conn, xmppStreamOpen, t.Config.Domain, stanza.NSClient, stanza.NSStream); err != nil {
		t.conn.Close()
		return "", NewConnError(err, true)
	}

	t.decoder = xml.NewDecoder(t.readWriter)
	t.decoder.CharsetReader = t.Config.CharsetReader
	sessionId, err := stanza.InitStream(t.decoder)
	if err != nil {
		t.conn.Close()
		return "", NewConnError(err, false)
	}
	t.readWriter = t.conn
	return sessionId, nil
}

func (t XMPPTransport) DoesStartTLS() bool {
	return true
}

func (t XMPPTransport) GetDecoder() *xml.Decoder {
	return t.decoder
}

func (t XMPPTransport) IsSecure() bool {
	return t.isSecure
}

func (t *XMPPTransport) StartTLS() error {
	if t.Config.TLSConfig == nil {
		t.TLSConfig = &tls.Config{}
	} else {
		t.TLSConfig = t.Config.TLSConfig.Clone()
	}

	if t.TLSConfig.ServerName == "" {
		t.TLSConfig.ServerName = t.Config.Domain
	}
	tlsConn := tls.Client(t.conn, t.TLSConfig)
	// We convert existing connection to TLS
	if err := tlsConn.Handshake(); err != nil {
		return err
	}

	if !t.TLSConfig.InsecureSkipVerify {
		if err := tlsConn.VerifyHostname(t.Config.Domain); err != nil {
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
	return t.readWriter.Read(p)
}

func (t XMPPTransport) Write(p []byte) (n int, err error) {
	return t.readWriter.Write(p)
}

func (t XMPPTransport) Close() error {
	_, _ = t.readWriter.Write([]byte("</stream:stream>"))
	return t.conn.Close()
}

func (t *XMPPTransport) LogTraffic(logFile io.Writer) {
	t.readWriter = &streamLogger{t.conn, logFile}
}
