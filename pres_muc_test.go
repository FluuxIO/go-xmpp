package xmpp_test

import (
	"encoding/xml"
	"testing"

	"gosrc.io/xmpp"
)

// https://xmpp.org/extensions/xep-0045.html#example-27
func TestMucPassword(t *testing.T) {
	str := `<presence
    from='hag66@shakespeare.lit/pda'
    id='djn4714'
    to='coven@chat.shakespeare.lit/thirdwitch'>
  <x xmlns='http://jabber.org/protocol/muc'>
    <password>cauldronburn</password>
  </x>
</presence>`

	var parsedPresence xmpp.Presence
	if err := xml.Unmarshal([]byte(str), &parsedPresence); err != nil {
		t.Errorf("Unmarshal(%s) returned error", str)
	}

	var muc xmpp.MucPresence
	if ok := parsedPresence.Get(&muc); !ok {
		t.Error("muc presence extension was not found")
	}

	if muc.Password != "cauldronburn" {
		t.Errorf("incorrect password: '%s'", muc.Password)
	}
}

// https://xmpp.org/extensions/xep-0045.html#example-37
func TestMucHistory(t *testing.T) {
	str := `<presence
    from='hag66@shakespeare.lit/pda'
    id='n13mt3l'
    to='coven@chat.shakespeare.lit/thirdwitch'>
  <x xmlns='http://jabber.org/protocol/muc'>
    <history maxstanzas='20'/>
  </x>
</presence>`

	var parsedPresence xmpp.Presence
	if err := xml.Unmarshal([]byte(str), &parsedPresence); err != nil {
		t.Errorf("Unmarshal(%s) returned error: %s", str, err)
		return
	}

	var muc xmpp.MucPresence
	if ok := parsedPresence.Get(&muc); !ok {
		t.Error("muc presence extension was not found")
		return
	}

	if v, ok := muc.History.MaxStanzas.Get(); !ok || v != 20 {
		t.Errorf("incorrect MaxStanzas: '%#v'", muc.History.MaxStanzas)
	}
}

// https://xmpp.org/extensions/xep-0045.html#example-37
func TestMucNoHistory(t *testing.T) {
	str := "<presence" +
		" id=\"n13mt3l\"" +
		" from=\"hag66@shakespeare.lit/pda\"" +
		" to=\"coven@chat.shakespeare.lit/thirdwitch\">" +
		"<x xmlns=\"http://jabber.org/protocol/muc\">" +
		"<history maxstanzas=\"0\"></history>" +
		"</x>" +
		"</presence>"

	maxstanzas := 0

	pres := xmpp.Presence{Attrs: xmpp.Attrs{
		From: "hag66@shakespeare.lit/pda",
		Id:   "n13mt3l",
		To:   "coven@chat.shakespeare.lit/thirdwitch",
	},
		Extensions: []xmpp.PresExtension{
			xmpp.MucPresence{
				History: xmpp.History{MaxStanzas: xmpp.NewNullableInt(maxstanzas)},
			},
		},
	}
	data, err := xml.Marshal(&pres)
	if err != nil {
		t.Error("error on encode:", err)
		return
	}

	if string(data) != str {
		t.Errorf("incorrect stanza: \n%s\n%s", str, data)
	}
}
