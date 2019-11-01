package xmpp

import (
	"bufio"
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
	openStatement string
	Config        TransportConfiguration
	TLSConfig     *tls.Config
	decoder       *xml.Decoder
	conn          net.Conn
	readWriter    io.ReadWriter
	logFile       io.Writer
	isSecure      bool
}

var componentStreamOpen = fmt.Sprintf("<?xml version='1.0'?><stream:stream to='%%s' xmlns='%s' xmlns:stream='%s'>", stanza.NSComponent, stanza.NSStream)

var clientStreamOpen = fmt.Sprintf("<?xml version='1.0'?><stream:stream to='%%s' xmlns='%s' xmlns:stream='%s' version='1.0'>", stanza.NSClient, stanza.NSStream)

func (t *XMPPTransport) Connect() (string, error) {
	var err error

	t.conn, err = net.DialTimeout("tcp", t.Config.Address, time.Duration(t.Config.ConnectTimeout)*time.Second)
	if err != nil {
		return "", NewConnError(err, true)
	}

	t.readWriter = newStreamLogger(t.conn, t.logFile)
	t.decoder = xml.NewDecoder(bufio.NewReaderSize(t.readWriter, maxPacketSize))
	t.decoder.CharsetReader = t.Config.CharsetReader
	return t.StartStream()
}

func (t XMPPTransport) StartStream() (string, error) {
	if _, err := fmt.Fprintf(t, t.openStatement, t.Config.Domain); err != nil {
		t.Close()
		return "", NewConnError(err, true)
	}

	sessionID, err := stanza.InitStream(t.GetDecoder())
	if err != nil {
		t.Close()
		return "", NewConnError(err, false)
	}
	return sessionID, nil
}

func (t XMPPTransport) DoesStartTLS() bool {
	return true
}

func (t XMPPTransport) GetDomain() string {
	return t.Config.Domain
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

	t.conn = tlsConn
	t.readWriter = newStreamLogger(tlsConn, t.logFile)
	t.decoder = xml.NewDecoder(bufio.NewReaderSize(t.readWriter, maxPacketSize))
	t.decoder.CharsetReader = t.Config.CharsetReader

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
	t.logFile = logFile
}
