package stanza_test

import (
	"encoding/xml"
	"testing"

	"gosrc.io/xmpp/stanza"
)

// Check that we can detect optional session from advertised stream features
func TestSession(t *testing.T) {
	streamFeatures := stanza.StreamFeatures{Session: stanza.StreamSession{Optional: true}}

	data, err := xml.Marshal(streamFeatures)
	if err != nil {
		t.Errorf("cannot marshal xml structure: %s", err)
	}

	parsedStream := stanza.StreamFeatures{}
	if err = xml.Unmarshal(data, &parsedStream); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !parsedStream.Session.Optional {
		t.Error("Session should be optional")
	}
}

// TODO Test Sasl mechanism
