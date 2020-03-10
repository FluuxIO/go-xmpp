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

	// Activate stream management process during session
	StreamManagementEnable bool
	// Enable stream management resume capability
	streamManagementResume bool
}

// IsStreamResumable tells if a stream session is resumable by reading the "config" part of a client.
// It checks if stream management is enabled, and if stream resumption was set and accepted by the server.
func IsStreamResumable(c *Client) bool {
	return c.config.StreamManagementEnable && c.config.streamManagementResume
}
