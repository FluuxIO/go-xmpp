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
	Retry          int      // Number of retries for connect
	ConnectTimeout int      // Connection timeout in seconds. Default to 15
	Insecure       bool     // set to true to allow comms without TLS
	CharsetReader  func(charset string, input io.Reader) (io.Reader, error) // passed to xml decoder
}
