package xmpp_test

import (
	"encoding/xml"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gosrc.io/xmpp"
)

func TestGenerateMessage(t *testing.T) {
	message := xmpp.NewMessage(xmpp.MessageTypeChat, xmpp.Attrs{From: "admin@localhost", To: "test@localhost", Id: "1"})
	message.Body = "Hi"
	message.Subject = "Msg Subject"

	data, err := xml.Marshal(message)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	parsedMessage := xmpp.Message{}
	if err = xml.Unmarshal(data, &parsedMessage); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !xmlEqual(parsedMessage, message) {
		t.Errorf("non matching items\n%s", cmp.Diff(parsedMessage, message))
	}
}

func TestDecodeError(t *testing.T) {
	str := `<message from='juliet@capulet.com'
         id='msg_1'
         to='romeo@montague.lit'
         type='error'>
  <error type='cancel'>
    <not-acceptable xmlns='urn:ietf:params:xml:ns:xmpp-stanzas'/>
  </error>
</message>`

	parsedMessage := xmpp.Message{}
	if err := xml.Unmarshal([]byte(str), &parsedMessage); err != nil {
		t.Errorf("message error stanza unmarshall error: %v", err)
		return
	}
	if parsedMessage.Error.Type != "cancel" {
		t.Errorf("incorrect error type: %s", parsedMessage.Error.Type)
	}
}
