package stanza_test

import (
	"testing"

	"gosrc.io/xmpp/stanza"
)

// Build a Software Version reply
// https://xmpp.org/extensions/xep-0092.html#example-2
func TestVersion_Builder(t *testing.T) {
	name := "Exodus"
	version := "0.7.0.4"
	os := "Windows-XP 5.01.2600"
	iq, err := stanza.NewIQ(stanza.Attrs{Type: "result", From: "romeo@montague.net/orchard",
		To: "juliet@capulet.com/balcony", Id: "version_1"})
	if err != nil {
		t.Fatalf("failed to create IQ: %v", err)
	}
	iq.Version().SetInfo(name, version, os)

	parsedIQ, err := checkMarshalling(t, iq)
	if err != nil {
		return
	}

	// Check result
	pp, ok := parsedIQ.Payload.(*stanza.Version)
	if !ok {
		t.Errorf("Parsed stanza does not contain correct IQ payload")
	}

	// Check version info
	if pp.Name != name {
		t.Errorf("Name Mismatch (expected: %s): %s", name, pp.Name)
	}
	if pp.Version != version {
		t.Errorf("Version Mismatch (expected: %s): %s", version, pp.Version)
	}
	if pp.OS != os {
		t.Errorf("OS Mismatch (expected: %s): %s", os, pp.OS)
	}
}
