package stanza_test

import (
	"encoding/xml"
	"testing"

	"gosrc.io/xmpp/stanza"
)

func TestDecodeRequest(t *testing.T) {
	str := `<message
    from='northumberland@shakespeare.lit/westminster'
    id='richard2-4.1.247'
    to='kingrichard@royalty.england.lit/throne'>
  <body>My lord, dispatch; read o'er these articles.</body>
  <request xmlns='urn:xmpp:receipts'/>
</message>`
	parsedMessage := stanza.Message{}
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

	switch ext := parsedMessage.Extensions[0].(type) {
	case *stanza.ReceiptRequest:
		if ext.XMLName.Local != "request" {
			t.Errorf("unexpected extension: %s:%s", ext.XMLName.Space, ext.XMLName.Local)
		}
	default:
		t.Errorf("could not find receipts extension")
	}

}
