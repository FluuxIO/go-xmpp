package main

import (
	"encoding/xml"
	"fmt"
	"log"

	"gosrc.io/xmpp/stanza"
)

func main() {
	iq, err := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, To: "service.localhost", Id: "custom-pl-1"})
	if err != nil {
		log.Fatalf("failed to create IQ: %v", err)
	}
	payload := CustomPayload{XMLName: xml.Name{Space: "my:custom:payload", Local: "query"}, Node: "test"}
	iq.Payload = payload

	data, err := xml.Marshal(iq)
	if err != nil {
		log.Fatalf("Cannot marshal iq with custom payload: %s", err)
	}

	var parsedIQ stanza.IQ
	if err = xml.Unmarshal(data, &parsedIQ); err != nil {
		log.Fatalf("Cannot unmarshal(%s): %s", data, err)
	}

	parsedPayload, ok := parsedIQ.Payload.(*CustomPayload)
	if !ok {
		log.Fatalf("Incorrect payload type: %#v", parsedIQ.Payload)
	}

	fmt.Printf("Parsed Payload: %#v", parsedPayload)

	if parsedPayload.Node != "test" {
		log.Fatalf("Incorrect node value: %s", parsedPayload.Node)
	}
}

type CustomPayload struct {
	XMLName xml.Name `xml:"my:custom:payload query"`
	Node    string   `xml:"node,attr,omitempty"`
}

func (c CustomPayload) Namespace() string {
	return c.XMLName.Space
}

func init() {
	stanza.TypeRegistry.MapExtension(stanza.PKTIQ, xml.Name{Space: "my:custom:payload", Local: "query"}, CustomPayload{})
}
