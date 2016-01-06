package xmpp

import "os"

type Options struct {
	Address      string
	Jid          string
	parsedJid    *Jid // For easier manipulation
	Password     string
	PacketLogger *os.File // Used for debugging
	Lang         string   // TODO: should default to 'en'
}
