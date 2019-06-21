package xmpp_test

import (
	"encoding/xml"
	"testing"

	"gosrc.io/xmpp"

	"github.com/google/go-cmp/cmp"
)

func TestGeneratePresence(t *testing.T) {
	presence := xmpp.NewPresence(xmpp.Attrs{From: "admin@localhost", To: "test@localhost", Id: "1"})
	presence.Show = "chat"

	data, err := xml.Marshal(presence)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	var parsedPresence xmpp.Presence
	if err = xml.Unmarshal(data, &parsedPresence); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !xmlEqual(parsedPresence, presence) {
		t.Errorf("non matching items\n%s", cmp.Diff(parsedPresence, presence))
	}
}

func TestPresenceSubElt(t *testing.T) {
	// Test structure to ensure that show, status and priority are correctly defined as presence
	// package sub-elements
	type pres struct {
		Show     xmpp.PresenceShow `xml:"show"`
		Status   string            `xml:"status"`
		Priority int8              `xml:"priority"`
	}

	presence := xmpp.NewPresence(xmpp.Attrs{From: "admin@localhost", To: "test@localhost", Id: "1"})
	presence.Show = xmpp.PresenceShowXA
	presence.Status = "Coding"
	presence.Priority = 10

	data, err := xml.Marshal(presence)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	var parsedPresence pres
	if err = xml.Unmarshal(data, &parsedPresence); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if parsedPresence.Show != presence.Show {
		t.Errorf("cannot read 'show' as presence subelement (%s)", parsedPresence.Show)
	}
	if parsedPresence.Status != presence.Status {
		t.Errorf("cannot read 'status' as presence subelement (%s)", parsedPresence.Status)
	}
	if parsedPresence.Priority != presence.Priority {
		t.Errorf("cannot read 'priority' as presence subelement (%d)", parsedPresence.Priority)
	}
}
