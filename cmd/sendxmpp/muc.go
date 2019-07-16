package main

import (
	"github.com/bdlm/log"

	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

var mucsToLeave []string

func joinMUC(c xmpp.Sender, to, nick string) error {

	toJID, err := xmpp.NewJid(to)
	if err != nil {
		return err
	}
	toJID.Resource = nick
	jid := toJID.Full()

	mucsToLeave = append(mucsToLeave, jid)

	return c.Send(stanza.Presence{Attrs: stanza.Attrs{To: jid},
		Extensions: []stanza.PresExtension{
			stanza.MucPresence{
				History: stanza.History{MaxStanzas: stanza.NewNullableInt(0)},
			}},
	})

}

func leaveMUCs(c xmpp.Sender) {
	for _, muc := range mucsToLeave {
		if err := c.Send(stanza.Presence{Attrs: stanza.Attrs{
			To:   muc,
			Type: stanza.PresenceTypeUnavailable,
		}}); err != nil {
			log.WithField("muc", muc).Errorf("error on leaving muc: %s", err)
		}
	}
}
