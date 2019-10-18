package xmpp

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

var TLSNotSupported = errors.New("Transport does not support StartTLS")

type TransportConfiguration struct {
	// Address is the XMPP Host and port to connect to. Host is of
	// the form 'serverhost:port' i.e "localhost:8888"
	Address        string
	Domain         string
	ConnectTimeout int // Client timeout in seconds. Default to 15
	// tls.Config must not be modified after having been passed to NewClient. Any
	// changes made after connecting are ignored.
	TLSConfig     *tls.Config
	CharsetReader func(charset string, input io.Reader) (io.Reader, error) // passed to xml decoder
}

type Transport interface {
	Connect() (string, error)
	DoesStartTLS() bool
	StartTLS() error

	LogTraffic(logFile io.Writer)

	GetDecoder() *xml.Decoder
	IsSecure() bool

	Ping() error
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}

// NewTransport creates a new Transport instance.
// The type of transport is determined by the address in the configuration:
// - if the address is a URL with the `ws` or `wss` scheme WebsocketTransport is used
// - in all other cases a XMPPTransport is used
// For XMPPTransport it is mandatory for the address to have a port specified.
func NewTransport(config TransportConfiguration) Transport {
	if strings.HasPrefix(config.Address, "ws:") || strings.HasPrefix(config.Address, "wss:") {
		return &WebsocketTransport{Config: config}
	}

	config.Address = ensurePort(config.Address, 5222)
	return &XMPPTransport{Config: config}
}
