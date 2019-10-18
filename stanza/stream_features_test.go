package stanza_test

import (
	"encoding/xml"
	"testing"

	"gosrc.io/xmpp/stanza"
)

func TestNoStartTLS(t *testing.T) {
	streamFeatures := `<stream:features xmlns:stream='http://etherx.jabber.org/streams'>
</stream:features>`

	var parsedSF stanza.StreamFeatures
	if err := xml.Unmarshal([]byte(streamFeatures), &parsedSF); err != nil {
		t.Errorf("Unmarshal(%s) returned error: %v", streamFeatures, err)
	}

	startTLS, ok := parsedSF.DoesStartTLS()
	if ok {
		t.Error("StartTLS feature should not be enabled")
	}
	if startTLS.Required {
		t.Error("StartTLS cannot be required as default")
	}
}

func TestStartTLS(t *testing.T) {
	streamFeatures := `<stream:features xmlns:stream='http://etherx.jabber.org/streams'>
  <starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'>
    <required/>
  </starttls>
</stream:features>`

	var parsedSF stanza.StreamFeatures
	if err := xml.Unmarshal([]byte(streamFeatures), &parsedSF); err != nil {
		t.Errorf("Unmarshal(%s) returned error: %v", streamFeatures, err)
	}

	startTLS, ok := parsedSF.DoesStartTLS()
	if !ok {
		t.Error("StartTLS feature should be enabled")
	}
	if !startTLS.Required {
		t.Error("StartTLS feature should be required")
	}
}

// TODO: Ability to support / detect previous version of stream management feature
func TestStreamManagement(t *testing.T) {
	streamFeatures := `<stream:features xmlns:stream='http://etherx.jabber.org/streams'>
    <sm xmlns='urn:xmpp:sm:3'/>
</stream:features>`

	var parsedSF stanza.StreamFeatures
	if err := xml.Unmarshal([]byte(streamFeatures), &parsedSF); err != nil {
		t.Errorf("Unmarshal(%s) returned error: %v", streamFeatures, err)
	}

	ok := parsedSF.DoesStreamManagement()
	if !ok {
		t.Error("Stream Management feature should have been detected")
	}
}
