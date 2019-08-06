package main

import (
	"github.com/bdlm/log"

	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

func send(c xmpp.Sender, recipient []string, msgText string) {
	msg := stanza.Message{
		Attrs: stanza.Attrs{Type: stanza.MessageTypeChat},
		Body:  msgText,
	}

	if isMUCRecipient {
		msg.Type = stanza.MessageTypeGroupchat
	}

	for _, to := range recipient {
		msg.To = to
		if err := c.Send(msg); err != nil {
			log.WithFields(map[string]interface{}{
				"muc":  isMUCRecipient,
				"to":   to,
				"text": msgText,
			}).Errorf("error on send message: %s", err)
		} else {
			log.WithFields(map[string]interface{}{
				"muc":  isMUCRecipient,
				"to":   to,
				"text": msgText,
			}).Info("send message")
		}
	}
}
