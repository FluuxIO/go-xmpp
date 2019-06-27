package stanza_test

import (
	"encoding/xml"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gosrc.io/xmpp/stanza"
)

func TestGenerateMessage(t *testing.T) {
	message := stanza.NewMessage(stanza.Attrs{Type: stanza.MessageTypeChat, From: "admin@localhost", To: "test@localhost", Id: "1"})
	message.Body = "Hi"
	message.Subject = "Msg Subject"

	data, err := xml.Marshal(message)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	parsedMessage := stanza.Message{}
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

	parsedMessage := stanza.Message{}
	if err := xml.Unmarshal([]byte(str), &parsedMessage); err != nil {
		t.Errorf("message error stanza unmarshall error: %v", err)
		return
	}
	if parsedMessage.Error.Type != "cancel" {
		t.Errorf("incorrect error type: %s", parsedMessage.Error.Type)
	}
}

func TestGetOOB(t *testing.T) {
	image := "https://localhost/image.png"
	msg := stanza.NewMessage(stanza.Attrs{To: "test@localhost"})
	ext := stanza.OOB{
		XMLName: xml.Name{Space: "jabber:x:oob", Local: "x"},
		URL:     image,
	}
	msg.Extensions = append(msg.Extensions, &ext)

	// OOB can properly be found
	var oob stanza.OOB
	// Try to find and
	if ok := msg.Get(&oob); !ok {
		t.Error("could not find oob extension")
		return
	}
	if oob.URL != image {
		t.Errorf("OOB URL was not properly extracted: ''%s", oob.URL)
	}

	// Markable is not found
	var m stanza.Markable
	if ok := msg.Get(&m); ok {
		t.Error("we should not have found markable extension")
	}
}
