package xmpp

import (
	"crypto/tls"
)

type TransportConfiguration struct {
	// Address is the XMPP Host and port to connect to. Host is of
	// the form 'serverhost:port' i.e "localhost:8888"
	Address        string
	ConnectTimeout int // Client timeout in seconds. Default to 15
	// tls.Config must not be modified after having been passed to NewClient. Any
	// changes made after connecting are ignored.
	TLSConfig *tls.Config
}

type Transport interface {
	Connect() error
	DoesStartTLS() bool
	StartTLS(domain string) error

	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}

func NewTransport(config TransportConfiguration) Transport {
	return &XMPPTransport{Config: config}

}
