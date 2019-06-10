package xmpp_test

import (
	"encoding/xml"
	"testing"

	"gosrc.io/xmpp"
)

func TestNoStartTLS(t *testing.T) {
	streamFeatures := `<stream:features xmlns:stream='http://etherx.jabber.org/streams'>
</stream:features>`

	var parsedSF xmpp.StreamFeatures
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

	var parsedSF xmpp.StreamFeatures
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
