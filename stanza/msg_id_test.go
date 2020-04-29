package stanza

import (
	"bytes"
	"encoding/xml"
	"testing"
)

const expectedMarshal = `<stanza-id xmlns="urn:xmpp:sid:0" by="jid" id="unique-id"></stanza-id>`

func TestMarshal(t *testing.T) {
	d := StanzaId{
		By: "jid",
		Id: "unique-id",
	}
	data, e := xml.Marshal(d)
	if e != nil || !bytes.Equal(data, []byte(expectedMarshal)) {
		t.Fatalf("Marshal failed. Expected: %v, Actual: %v", expectedMarshal, string(data))
	}
}
