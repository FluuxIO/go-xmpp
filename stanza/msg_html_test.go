package stanza_test

import (
	"encoding/xml"
	"testing"

	"gosrc.io/xmpp/stanza"
)

func TestHTMLGen(t *testing.T) {
	htmlBody := "<p>Hello <b>World</b></p>"
	msg := stanza.NewMessage(stanza.Attrs{To: "test@localhost"})
	msg.Body = "Hello World"
	body := stanza.HTMLBody{
		InnerXML: htmlBody,
	}
	html := stanza.HTML{Body: body}
	msg.Extensions = append(msg.Extensions, html)

	result := msg.XMPPFormat()
	str := `<message to="test@localhost"><body>Hello World</body><html xmlns="http://jabber.org/protocol/xhtml-im"><body xmlns="http://www.w3.org/1999/xhtml"><p>Hello <b>World</b></p></body></html></message>`
	if result != str {
		t.Errorf("incorrect serialize message:\n%s", result)
	}

	parsedMessage := stanza.Message{}
	if err := xml.Unmarshal([]byte(str), &parsedMessage); err != nil {
		t.Errorf("message HTML unmarshall error: %v", err)
		return
	}

	if parsedMessage.Body != msg.Body {
		t.Errorf("incorrect parsed body: '%s'", parsedMessage.Body)
	}

	var h stanza.HTML
	if ok := parsedMessage.Get(&h); !ok {
		t.Error("could not extract HTML body")
	}

	if h.Body.InnerXML != htmlBody {
		t.Errorf("could not extract html body: '%s'", h.Body.InnerXML)
	}
}
