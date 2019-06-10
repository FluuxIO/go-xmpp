package xmpp // import "gosrc.io/xmpp"

import (
	"crypto/tls"
	"encoding/xml"
)

var DefaultTlsConfig tls.Config

// Used during stream initiation / session establishment
type tlsProceed struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls proceed"`
}

type tlsFailure struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-tls failure"`
}
