package stanza

import (
	"encoding/xml"
	"testing"
)

func TestNode_Marshal(t *testing.T) {
	jsonData := []byte("{\"key\":\"value\"}")

	iqResp, err := NewIQ(Attrs{Type: "result", From: "admin@localhost", To: "test@localhost", Id: "1"})
	if err != nil {
		t.Fatalf("failed to create IQ: %v", err)
	}
	iqResp.Any = &Node{
		XMLName: xml.Name{Space: "myNS", Local: "space"},
		Content: string(jsonData),
	}

	bytes, err := xml.Marshal(iqResp)
	if err != nil {
		t.Errorf("Could not marshal XML: %v", err)
	}

	parsedIQ := IQ{}
	if err := xml.Unmarshal(bytes, &parsedIQ); err != nil {
		t.Errorf("Unmarshal returned error: %v", err)
	}

	if parsedIQ.Any.Content != string(jsonData) {
		t.Errorf("Cannot find generic any payload in parsedIQ: '%s'", parsedIQ.Any.Content)
	}
}
