package stanza

import (
	"encoding/xml"
	"testing"
)

func TestControlSet(t *testing.T) {
	packet := `
<iq to='test@localhost/jukebox' from='admin@localhost/mbp' type='set' id='2'>
 <set xmlns='urn:xmpp:iot:control' xml:lang='en'>
	<string name='action' value='play'/>
	<string name='url' value='https://soundcloud.com/radiohead/spectre'/>
 </set>
</iq>`

	parsedIQ := IQ{}
	data := []byte(packet)
	if err := xml.Unmarshal(data, &parsedIQ); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if cs, ok := parsedIQ.Payload.(*ControlSet); !ok {
		t.Errorf("Payload is not an iot control set: %v", cs)
	}
}
