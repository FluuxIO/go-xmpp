package xmpp // import "gosrc.io/xmpp"

import (
	"testing"
)

func TestValidJids(t *testing.T) {
	tests := []struct {
		jidstr   string
		expected Jid
	}{
		{jidstr: "test@domain.com", expected: Jid{"test", "domain.com", ""}},
		{jidstr: "test@domain.com/resource", expected: Jid{"test", "domain.com", "resource"}},
		// resource can contain '/' or '@'
		{jidstr: "test@domain.com/a/b", expected: Jid{"test", "domain.com", "a/b"}},
		{jidstr: "test@domain.com/a@b", expected: Jid{"test", "domain.com", "a@b"}},
		{jidstr: "domain.com", expected: Jid{"", "domain.com", ""}},
	}

	for _, tt := range tests {
		jid, err := NewJid(tt.jidstr)
		if err != nil {
			t.Errorf("could not parse correct jid: %s", tt.jidstr)
			continue
		}

		if jid == nil {
			t.Error("jid should not be nil")
		}

		if jid.username != tt.expected.username {
			t.Errorf("incorrect jid username (%s): %s", tt.expected.username, jid.username)
		}

		if jid.username != tt.expected.username {
			t.Errorf("incorrect jid domain (%s): %s", tt.expected.domain, jid.domain)
		}

		if jid.resource != tt.expected.resource {
			t.Errorf("incorrect jid resource (%s): %s", tt.expected.resource, jid.resource)
		}
	}
}

func TestIncorrectJids(t *testing.T) {
	badJids := []string{
		"",
		"user@",
		"@domain.com",
		"user:name@domain.com",
		"user<name@domain.com",
		"test@domain.com@otherdomain.com",
		"test@domain com/resource",
	}

	for _, sjid := range badJids {
		if _, err := NewJid(sjid); err == nil {
			t.Error("parsing incorrect jid should return error: " + sjid)
		}
	}
}
