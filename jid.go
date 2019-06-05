package xmpp // import "gosrc.io/xmpp"

import (
	"errors"
	"strings"
	"unicode"
)

type Jid struct {
	username string
	domain   string
	resource string
}

func NewJid(sjid string) (*Jid, error) {
	s1 := strings.SplitN(sjid, "@", 2)
	if len(s1) != 2 {
		return nil, errors.New("invalid JID, missing domain: " + sjid)
	}
	jid := new(Jid)
	jid.username = s1[0]
	if !isUsernameValid(jid.username) {
		return jid, errors.New("invalid domain: " + jid.username)
	}

	s2 := strings.SplitN(s1[1], "/", 2)

	jid.domain = s2[0]
	if !isDomainValid(jid.domain) {
		return jid, errors.New("invalid domain: " + jid.domain)
	}

	if len(s2) == 2 {
		jid.resource = s2[1]
	}

	return jid, nil
}

func isUsernameValid(username string) bool {
	invalidRunes := []rune{'@', '/', '\'', '"', ':', '<', '>'}
	return strings.IndexFunc(username, isInvalid(invalidRunes)) < 0
}

func isDomainValid(domain string) bool {
	invalidRunes := []rune{'@', '/'}
	return strings.IndexFunc(domain, isInvalid(invalidRunes)) < 0
}

func isInvalid(invalidRunes []rune) func(c rune) bool {
	isInvalid := func(c rune) bool {
		if unicode.IsSpace(c) {
			return true
		}
		for _, r := range invalidRunes {
			if c == r {
				return true
			}
		}
		return false
	}
	return isInvalid
}
