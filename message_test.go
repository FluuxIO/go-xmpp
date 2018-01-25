package xmpp

import (
	"encoding/xml"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGenerateMessage(t *testing.T) {
	message := NewMessage("chat", "admin@localhost", "test@localhost", "1", "en")
	message.Body = "Hi"
	message.Subject = "Msg Subject"

	data, err := xml.Marshal(message)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	parsedMessage := Message{}
	if err = xml.Unmarshal(data, &parsedMessage); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !xmlEqual(parsedMessage, message) {
		t.Errorf("non matching items\n%s", cmp.Diff(parsedMessage, message))
	}
}
