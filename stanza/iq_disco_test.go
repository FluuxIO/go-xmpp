package stanza_test

import (
	"encoding/xml"
	"testing"

	"gosrc.io/xmpp/stanza"
)

func TestDiscoInfoBuilder(t *testing.T) {
	iq := stanza.NewIQ(stanza.Attrs{Type: "get", To: "service.localhost", Id: "disco-get-1"})
	disco := iq.DiscoInfo()
	disco.AddIdentity("Test Component", "gateway", "service")
	disco.AddFeatures(stanza.NSDiscoInfo, stanza.NSDiscoItems, "jabber:iq:version", "urn:xmpp:delegation:1")

	// Marshall
	data, err := xml.Marshal(iq)
	if err != nil {
		t.Errorf("cannot marshal xml structure: %s", err)
		return
	}

	// Unmarshall
	var parsedIQ stanza.IQ
	if err = xml.Unmarshal(data, &parsedIQ); err != nil {
		t.Errorf("Unmarshal(%s) returned error: %s", data, err)
	}

	// Check result
	pp, ok := parsedIQ.Payload.(*stanza.DiscoInfo)
	if !ok {
		t.Errorf("Parsed stanza does not contain an IQ payload")
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
