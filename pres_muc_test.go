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
		t.Errorf("Unmarshal(%s) returned error", str)
	}

	var muc xmpp.MucPresence
	if ok := parsedPresence.Get(&muc); !ok {
		t.Error("muc presence extension was not found")
	}

	if muc.History.MaxStanzas != 20 {
		t.Errorf("incorrect max stanza: '%d'", muc.History.MaxStanzas)
	}
}
