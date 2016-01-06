package xmpp

import (
	"testing"
)

func TestBareJid(t *testing.T) {
	var jid *Jid
	var err error

	bareJid := "test@domain.com"

	if jid, err = NewJid(bareJid); err != nil {
		t.Error("could not parse bare jid")
	}

	if jid.username != "test" {
		t.Error("incorrect bare jid username")
	}

	if jid.domain != "domain.com" {
		t.Error("incorrect bare jid domain")
	}

	if jid.resource != "" {
		t.Error("bare jid resource should be empty")
	}
}
