package stanza_test

import (
	"encoding/xml"
	"testing"

	"gosrc.io/xmpp/stanza"
)

// Test DiscoInfo Builder with several features
func TestDiscoInfo_Builder(t *testing.T) {
	iq, err := stanza.NewIQ(stanza.Attrs{Type: "get", To: "service.localhost", Id: "disco-get-1"})
	if err != nil {
		t.Fatalf("failed to create IQ: %v", err)
	}
	disco := iq.DiscoInfo()
	disco.AddIdentity("Test Component", "gateway", "service")
	disco.AddFeatures(stanza.NSDiscoInfo, stanza.NSDiscoItems, "jabber:iq:version", "urn:xmpp:delegation:1")

	parsedIQ, err := checkMarshalling(t, iq)
	if err != nil {
		return
	}

	// Check result
	pp, ok := parsedIQ.Payload.(*stanza.DiscoInfo)
	if !ok {
		t.Errorf("Parsed stanza does not contain correct IQ payload")
	}

	// Check features
	features := []string{stanza.NSDiscoInfo, stanza.NSDiscoItems, "jabber:iq:version", "urn:xmpp:delegation:1"}
	if len(pp.Features) != len(features) {
		t.Errorf("Features length mismatch: %#v", pp.Features)
	} else {
		for i, f := range pp.Features {
			if f.Var != features[i] {
				t.Errorf("Missing feature: %s", features[i])
			}
		}
	}

	// Check identity
	if len(pp.Identity) != 1 {
		t.Errorf("Identity length mismatch: %#v", pp.Identity)
	} else {
		if pp.Identity[0].Name != "Test Component" {
			t.Errorf("Incorrect identity name: %#v", pp.Identity[0].Name)
		}
	}
}

// Implements XEP-0030 example 17
// https://xmpp.org/extensions/xep-0030.html#example-17
func TestDiscoItems_Builder(t *testing.T) {
	iq, err := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeResult, From: "catalog.shakespeare.lit",
		To: "romeo@montague.net/orchard", Id: "items-2"})
	if err != nil {
		t.Fatalf("failed to create IQ: %v", err)
	}
	iq.DiscoItems().
		AddItem("catalog.shakespeare.lit", "books", "Books by and about Shakespeare").
		AddItem("catalog.shakespeare.lit", "clothing", "Wear your literary taste with pride").
		AddItem("catalog.shakespeare.lit", "music", "Music from the time of Shakespeare")

	parsedIQ, err := checkMarshalling(t, iq)
	if err != nil {
		return
	}

	// Check result
	pp, ok := parsedIQ.Payload.(*stanza.DiscoItems)
	if !ok {
		t.Errorf("Parsed stanza does not contain correct IQ payload")
	}

	// Check items
	items := []stanza.DiscoItem{{xml.Name{}, "catalog.shakespeare.lit", "books", "Books by and about Shakespeare"},
		{xml.Name{}, "catalog.shakespeare.lit", "clothing", "Wear your literary taste with pride"},
		{xml.Name{}, "catalog.shakespeare.lit", "music", "Music from the time of Shakespeare"}}
	if len(pp.Items) != len(items) {
		t.Errorf("List length mismatch: %#v", pp.Items)
	} else {
		for i, item := range pp.Items {
			if item.JID != items[i].JID {
				t.Errorf("Jid Mismatch (expected: %s): %s", items[i].JID, item.JID)
			}
			if item.Node != items[i].Node {
				t.Errorf("Node Mismatch (expected: %s): %s", items[i].JID, item.JID)
			}
			if item.Name != items[i].Name {
				t.Errorf("Name Mismatch (expected: %s): %s", items[i].JID, item.JID)
			}
		}
	}
}
