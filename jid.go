package xmpp // import "fluux.io/xmpp"

import (
	"errors"
	"strings"
)

type Jid struct {
	Local    string
	Domain   string
	Resource string
}

func NewJid(sjid string) (jid *Jid, err error) {
	s1 := strings.Split(sjid, "@")
	if len(s1) != 2 {
		err = errors.New("invalid JID: " + sjid)
		return
	}
	jid = new(Jid)
	jid.Local = s1[0]

	s2 := strings.Split(s1[1], "/")
	if len(s2) > 2 {
		err = errors.New("invalid JID: " + sjid)
		return
	}
	jid.Domain = s2[0]
	if len(s2) == 2 {
		jid.Resource = s2[1]
	}

	return
}
