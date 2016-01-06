package xmpp

import (
	"errors"
	"strings"
)

type Jid struct {
	username string
	domain   string
	resource string
}

func NewJid(sjid string) (jid *Jid, err error) {
	s1 := strings.Split(sjid, "@")
	if len(s1) != 2 {
		err = errors.New("invalid JID: " + sjid)
		return
	}
	jid = new(Jid)
	jid.username = s1[0]

	s2 := strings.Split(s1[1], "/")
	if len(s2) > 2 {
		err = errors.New("invalid JID: " + sjid)
		return
	}
	jid.domain = s2[0]
	if len(s2) == 2 {
		jid.resource = s2[1]
	}

	return
}
