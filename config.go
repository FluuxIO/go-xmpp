package xmpp // import "gosrc.io/xmpp"

import (
	"io"
	"os"
)

type Config struct {
	Address        string
	Jid            string
	parsedJid      *Jid // For easier manipulation
	Password       string
	PacketLogger   *os.File // Used for debugging
	Lang           string   // TODO: should default to 'en'
	ConnectTimeout int      // Connection timeout in seconds. Default to 15
	// Insecure can be set to true to allow to open a session without TLS. If TLS
	// is supported on the server, we will still try to use it.
	Insecure      bool
	CharsetReader func(charset string, input io.Reader) (io.Reader, error) // passed to xml decoder
}
