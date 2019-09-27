package stanza

import (
	"encoding/xml"
	"testing"
)

func TestErr_UnmarshalXML(t *testing.T) {
	packet := `
 <iq from='pubsub.example.com'
       id='kj4vz31m'
       to='romeo@example.net/foo'
       type='error'>
  <error type='wait'>
    <resource-constraint
        xmlns='urn:ietf:params:xml:ns:xmpp-stanzas'/>
    <text xmlns='urn:ietf:params:xml:ns:xmpp-stanzas'>System overloaded, please retry</text>
  </error>
 </iq>`

	parsedIQ := IQ{}
	data := []byte(packet)
	if err := xml.Unmarshal(data, &parsedIQ); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	xmppError := parsedIQ.Error
	if xmppError.Text != "System overloaded, please retry" {
		t.Errorf("Could not extract error text: '%s'", xmppError.Text)
	}
}
