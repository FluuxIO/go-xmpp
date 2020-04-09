package stanza_test

import (
	"encoding/xml"
	"gosrc.io/xmpp/stanza"
	"reflect"
	"strings"
	"testing"
)

const msg_const = `
<message
    from="romeo@montague.lit/laptop"
    to="juliet@capulet.lit/laptop">
  <body>V unir avtugf pybnx gb uvqr zr sebz gurve fvtug</body>
  <no-copy xmlns="urn:xmpp:hints"></no-copy>
  <no-permanent-store xmlns="urn:xmpp:hints"></no-permanent-store>
  <no-store xmlns="urn:xmpp:hints"></no-store>
  <store xmlns="urn:xmpp:hints"></store>
</message>`

func TestSerializationHint(t *testing.T) {
	msg := stanza.NewMessage(stanza.Attrs{To: "juliet@capulet.lit/laptop", From: "romeo@montague.lit/laptop"})
	msg.Body = "V unir avtugf pybnx gb uvqr zr sebz gurve fvtug"
	msg.Extensions = append(msg.Extensions, stanza.HintNoCopy{}, stanza.HintNoPermanentStore{}, stanza.HintNoStore{}, stanza.HintStore{})
	data, _ := xml.Marshal(msg)
	if strings.ReplaceAll(strings.Join(strings.Fields(msg_const), ""), "\n", "") != strings.Join(strings.Fields(string(data)), "") {
		t.Fatalf("marshalled message does not match expected message")
	}
}

func TestUnmarshalHints(t *testing.T) {
	// Init message as in the const value
	msgConst := stanza.NewMessage(stanza.Attrs{To: "juliet@capulet.lit/laptop", From: "romeo@montague.lit/laptop"})
	msgConst.Body = "V unir avtugf pybnx gb uvqr zr sebz gurve fvtug"
	msgConst.Extensions = append(msgConst.Extensions, &stanza.HintNoCopy{}, &stanza.HintNoPermanentStore{}, &stanza.HintNoStore{}, &stanza.HintStore{})

	// Compare message with the const value
	msg := stanza.Message{}
	err := xml.Unmarshal([]byte(msg_const), &msg)
	if err != nil {
		t.Fatal(err)
	}

	if msgConst.XMLName.Local != msg.XMLName.Local {
		t.Fatalf("message tags do not match. Expected: %s, Actual: %s", msgConst.XMLName.Local, msg.XMLName.Local)
	}
	if msgConst.Body != msg.Body {
		t.Fatalf("message bodies do not match. Expected: %s, Actual: %s", msgConst.Body, msg.Body)
	}

	if !reflect.DeepEqual(msgConst.Attrs, msg.Attrs) {
		t.Fatalf("attributes do not match")
	}

	if !reflect.DeepEqual(msgConst.Error, msg.Error) {
		t.Fatalf("attributes do not match")
	}
	var found bool
	for _, ext := range msgConst.Extensions {
		for _, strExt := range msg.Extensions {
			if reflect.TypeOf(ext) == reflect.TypeOf(strExt) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("extensions do not match")
		}
		found = false
	}
}
