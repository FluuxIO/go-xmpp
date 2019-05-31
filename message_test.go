package xmpp_test

import (
	"encoding/xml"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gosrc.io/xmpp"
)

func TestGenerateMessage(t *testing.T) {
	message := xmpp.NewMessage("chat", "admin@localhost", "test@localhost", "1", "en")
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

func TestDecodeXEP0184(t *testing.T) {
	str := `<message
    from='northumberland@shakespeare.lit/westminster'
    id='richard2-4.1.247'
    to='kingrichard@royalty.england.lit/throne'>
  <body>My lord, dispatch; read o'er these articles.</body>
  <request xmlns='urn:xmpp:receipts'/>
</message>`
	parsedMessage := xmpp.Message{}
	if err := xml.Unmarshal([]byte(str), &parsedMessage); err != nil {
		t.Errorf("message receipt unmarshall error: %v", err)
		return
	}

	if parsedMessage.Body != "My lord, dispatch; read o'er these articles." {
		t.Errorf("Unexpected body: '%s'", parsedMessage.Body)
	}

	if len(parsedMessage.Extensions) < 1 {
		t.Errorf("no extension found on parsed message")
		return
	}

	switch parsedMessage.Extensions[0].(type) {
	case *xmpp.ReceiptRequest:
	case *xmpp.ReceiptReceived:
		t.Errorf("wrong local in receipt namespace")
	default:
		t.Errorf("could not find receipt extension")
	}

}
