package xmpp

import (
	"crypto/tls"
)

type TransportConfiguration struct {
	ConnectTimeout int // Client timeout in seconds. Default to 15
	// tls.Config must not be modified after having been passed to NewClient. Any
	// changes made after connecting are ignored.
	TLSConfig *tls.Config
}

type Transport interface {
	Connect(address string) error
	DoesStartTLS() bool
	StartTLS(domain string) error

	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
}
