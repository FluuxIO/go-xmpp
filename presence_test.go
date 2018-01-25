package xmpp

import (
	"encoding/xml"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGeneratePresence(t *testing.T) {
	presence := NewPresence("admin@localhost", "test@localhost", "1", "en")
	presence.Show = "chat"

	data, err := xml.Marshal(presence)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	parsedPresence := Presence{}
	if err = xml.Unmarshal(data, &parsedPresence); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !xmlEqual(parsedPresence, presence) {
		t.Errorf("non matching items\n%s", cmp.Diff(parsedPresence, presence))
	}
}
