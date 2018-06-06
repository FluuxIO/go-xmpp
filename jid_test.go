package xmpp // import "fluux.io/xmpp"

import (
	"testing"
)

func TestValidJids(t *testing.T) {
	var jid *Jid
	var err error

	goodJids := []string{"test@domain.com", "test@domain.com/resource"}

	for i, sjid := range goodJids {
		if jid, err = NewJid(sjid); err != nil {
			t.Error("could not parse correct jid")
		}

		if jid.Local != "test" {
			t.Error("incorrect jid username")
		}

		if jid.Domain != "domain.com" {
			t.Error("incorrect jid domain")
		}

		if i == 0 && jid.Resource != "" {
			t.Error("bare jid resource should be empty")
		}

		if i == 1 && jid.Resource != "resource" {
			t.Error("incorrect full jid resource")
		}
	}
}

// TODO: Check if resource cannot contain a /
func TestIncorrectJids(t *testing.T) {
	badJids := []string{"test@domain.com@otherdomain.com",
		"test@domain.com/test/test"}

	for _, sjid := range badJids {
		if _, err := NewJid(sjid); err == nil {
			t.Error("parsing incorrect jid should return error: " + sjid)
		}
	}
}
