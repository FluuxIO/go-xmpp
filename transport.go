package xmpp

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrTransportProtocolNotSupported = errors.New("transport protocol not supported")
var ErrTLSNotSupported = errors.New("transport does not support StartTLS")

// TODO: rename to transport config?
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

	StartStream() (string, error)
	GetDecoder() *xml.Decoder
	IsSecure() bool

	Ping() error
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
	// ReceivedStreamClose signals to the transport that a </stream:stream> has been received and that the tcp connection
	// should be closed.
	ReceivedStreamClose()
}

// NewClientTransport creates a new Transport instance for clients.
// The type of transport is determined by the address in the configuration:
// - if the address is a URL with the `ws` or `wss` scheme WebsocketTransport is used
// - in all other cases a XMPPTransport is used
// For XMPPTransport it is mandatory for the address to have a port specified.
func NewClientTransport(config TransportConfiguration) Transport {
	if strings.HasPrefix(config.Address, "ws:") || strings.HasPrefix(config.Address, "wss:") {
		return &WebsocketTransport{Config: config}
	}

	config.Address = ensurePort(config.Address, 5222)
	return &XMPPTransport{
		Config:        config,
		openStatement: clientStreamOpen,
	}
}

// NewComponentTransport creates a new Transport instance for components.
// Only XMPP transports are allowed. If you try to use any other protocol an error
// will be returned.
func NewComponentTransport(config TransportConfiguration) (Transport, error) {
	if strings.HasPrefix(config.Address, "ws:") || strings.HasPrefix(config.Address, "wss:") {
		return nil, fmt.Errorf("components only support XMPP transport: %w", ErrTransportProtocolNotSupported)
	}

	config.Address = ensurePort(config.Address, 5222)
	return &XMPPTransport{
		Config:        config,
		openStatement: componentStreamOpen,
	}, nil
}
