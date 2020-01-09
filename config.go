package xmpp

import (
	"gosrc.io/xmpp/stanza"
	"os"
	"time"
)

// Config & TransportConfiguration must not be modified after having been passed to NewClient. Any
// changes made after connecting are ignored.
type Config struct {
	TransportConfiguration

	Jid               string
	parsedJid         *stanza.Jid // For easier manipulation
	Credential        Credential
	StreamLogger      *os.File      // Used for debugging
	Lang              string        // TODO: should default to 'en'
	KeepaliveInterval time.Duration // Interval between keepalive packets
	ConnectTimeout    int           // Client timeout in seconds. Default to 15
	// Insecure can be set to true to allow to open a session without TLS. If TLS
	// is supported on the server, we will still try to use it.
	Insecure bool
}
