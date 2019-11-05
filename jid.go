package xmpp

import (
	"fmt"
	"strings"
	"unicode"
)

type Jid struct {
	Node     string
	Domain   string
	Resource string
}

func NewJid(sjid string) (*Jid, error) {
	jid := new(Jid)

	if sjid == "" {
		return jid, fmt.Errorf("jid cannot be empty")
	}

	s1 := strings.SplitN(sjid, "@", 2)
	if len(s1) == 1 { // This is a server or component JID
		jid.Domain = s1[0]
	} else { // JID has a local username part
		if s1[0] == "" {
			return jid, fmt.Errorf("invalid jid '%s", sjid)
		}
		jid.Node = s1[0]
		if s1[1] == "" {
			return jid, fmt.Errorf("domain cannot be empty")
		}
		jid.Domain = s1[1]
	}

	// Extract resource from domain field
	s2 := strings.SplitN(jid.Domain, "/", 2)
	if len(s2) == 2 { // If len = 1, domain is already correct, and resource is already empty
		jid.Domain = s2[0]
		jid.Resource = s2[1]
	}

	if !isUsernameValid(jid.Node) {
		return jid, fmt.Errorf("invalid Node in JID '%s'", sjid)
	}
	if !isDomainValid(jid.Domain) {
		return jid, fmt.Errorf("invalid domain in JID '%s'", sjid)
	}

	return jid, nil
}

func (j *Jid) Full() string {
	if j.Resource == "" {
		return j.Bare()
	} else if j.Node == "" {
		return j.Node + "/" + j.Resource
	} else {
		return j.Node + "@" + j.Domain + "/" + j.Resource
	}
}

func (j *Jid) Bare() string {
	if j.Node == "" {
		return j.Domain
	} else {
		return j.Node + "@" + j.Domain
	}
}

// ============================================================================
// Helpers, for parsing / validation

func isUsernameValid(username string) bool {
	invalidRunes := []rune{'@', '/', '\'', '"', ':', '<', '>'}
	return strings.IndexFunc(username, isInvalid(invalidRunes)) < 0
}

func isDomainValid(domain string) bool {
	if len(domain) == 0 {
		return false
	}

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
