package main

import (
	"github.com/bdlm/log"

	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

func send(c xmpp.Sender, receiver []string, msgText string) {
	msg := stanza.Message{
		Attrs: stanza.Attrs{Type: stanza.MessageTypeChat},
		Body:  msgText,
	}
	if receiverMUC {
		msg.Type = stanza.MessageTypeGroupchat
	}
	for _, to := range receiver {
		msg.To = to
		if err := c.Send(msg); err != nil {
			log.WithFields(map[string]interface{}{
				"muc":  receiverMUC,
				"to":   to,
				"text": msgText,
			}).Errorf("error on send message: %s", err)
		} else {
			log.WithFields(map[string]interface{}{
				"muc":  receiverMUC,
				"to":   to,
				"text": msgText,
			}).Info("send message")
		}
	}
}
