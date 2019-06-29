package stanza_test

import (
	"encoding/xml"
	"testing"

	"gosrc.io/xmpp/stanza"
)

// Check that we can detect optional session from advertised stream features
func TestSessionFeatures(t *testing.T) {
	streamFeatures := stanza.StreamFeatures{Session: stanza.StreamSession{Optional: true}}

	data, err := xml.Marshal(streamFeatures)
	if err != nil {
		t.Errorf("cannot marshal xml structure: %s", err)
	}

	parsedStream := stanza.StreamFeatures{}
	if err = xml.Unmarshal(data, &parsedStream); err != nil {
		t.Errorf("Unmarshal(%s) returned error: %s", data, err)
	}

	if !parsedStream.Session.IsOptional() {
		t.Error("Session should be optional")
	}
}

// Check that the Session tag can be used in IQ decoding
func TestSessionIQ(t *testing.T) {
	iq := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeSet, Id: "session"})
	iq.Payload = &stanza.StreamSession{XMLName: xml.Name{Local: "session"}, Optional: true}

	data, err := xml.Marshal(iq)
	if err != nil {
		t.Errorf("cannot marshal xml structure: %s", err)
		return
	}

	parsedIQ := stanza.IQ{}
	if err = xml.Unmarshal(data, &parsedIQ); err != nil {
		t.Errorf("Unmarshal(%s) returned error: %s", data, err)
		return
	}

	session, ok := parsedIQ.Payload.(*stanza.StreamSession)
	if !ok {
		t.Error("Missing session payload")
		return
	}

	if !session.IsOptional() {
		t.Error("Session should be optional")
	}
}

// TODO Test Sasl mechanism
