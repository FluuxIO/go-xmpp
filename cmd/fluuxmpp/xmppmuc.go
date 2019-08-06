package main

import (
	"github.com/bdlm/log"

	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

func joinMUC(c xmpp.Sender, toJID *xmpp.Jid) error {
	return c.Send(stanza.Presence{Attrs: stanza.Attrs{To: toJID.Full()},
		Extensions: []stanza.PresExtension{
			stanza.MucPresence{
				History: stanza.History{MaxStanzas: stanza.NewNullableInt(0)},
			}},
	})
}

func leaveMUCs(c xmpp.Sender, mucsToLeave []*xmpp.Jid) {
	for _, muc := range mucsToLeave {
		if err := c.Send(stanza.Presence{Attrs: stanza.Attrs{
			To:   muc.Full(),
			Type: stanza.PresenceTypeUnavailable,
		}}); err != nil {
			log.WithField("muc", muc).Errorf("error on leaving muc: %s", err)
		}
	}
}
