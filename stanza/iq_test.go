package stanza_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalIqs(t *testing.T) {
	//var cs1 = new(iot.ControlSet)
	var tests = []struct {
		iqString string
		parsedIQ IQ
	}{
		{"<iq id=\"1\" type=\"set\" to=\"test@localhost\"/>",
			IQ{XMLName: xml.Name{Local: "iq"}, Attrs: Attrs{Type: IQTypeSet, To: "test@localhost", Id: "1"}}},
		//{"<iq xmlns=\"jabber:client\" id=\"2\" type=\"set\" to=\"test@localhost\" from=\"server\"><set xmlns=\"urn:xmpp:iot:control\"/></iq>", IQ{XMLName: xml.Name{Space: "jabber:client", Local: "iq"}, PacketAttrs: PacketAttrs{To: "test@localhost", From: "server", Type: "set", Id: "2"}, Payload: cs1}},
	}

	for _, test := range tests {
		parsedIQ := IQ{}
		err := xml.Unmarshal([]byte(test.iqString), &parsedIQ)
		if err != nil {
			t.Errorf("Unmarshal(%s) returned error", test.iqString)
		}

		if !xmlEqual(parsedIQ, test.parsedIQ) {
			t.Errorf("non matching items\n%s", cmp.Diff(parsedIQ, test.parsedIQ))
		}

	}
}

func TestGenerateIq(t *testing.T) {
	iq := NewIQ(Attrs{Type: IQTypeResult, From: "admin@localhost", To: "test@localhost", Id: "1"})
	payload := DiscoInfo{
		Identity: Identity{
			Name:     "Test Gateway",
			Category: "gateway",
			Type:     "mqtt",
		},
		Features: []Feature{
			{Var: NSDiscoInfo},
			{Var: NSDiscoItems},
		},
	}
	iq.Payload = &payload

	data, err := xml.Marshal(iq)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	if strings.Contains(string(data), "<error ") {
		t.Error("empty error should not be serialized")
	}

	parsedIQ := IQ{}
	if err = xml.Unmarshal(data, &parsedIQ); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !xmlEqual(parsedIQ.Payload, iq.Payload) {
		t.Errorf("non matching items\n%s", cmp.Diff(parsedIQ.Payload, iq.Payload))
	}
}

func TestErrorTag(t *testing.T) {
	xError := Err{
		XMLName: xml.Name{Local: "error"},
		Code:    503,
		Type:    "cancel",
		Reason:  "service-unavailable",
		Text:    "User session not found",
	}

	data, err := xml.Marshal(xError)
	if err != nil {
		t.Errorf("cannot marshal xml structure: %s", err)
	}

	parsedError := Err{}
	if err = xml.Unmarshal(data, &parsedError); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !xmlEqual(parsedError, xError) {
		t.Errorf("non matching items\n%s", cmp.Diff(parsedError, xError))
	}
}

func TestDiscoItems(t *testing.T) {
	iq := NewIQ(Attrs{Type: IQTypeGet, From: "romeo@montague.net/orchard", To: "catalog.shakespeare.lit", Id: "items3"})
	payload := DiscoItems{
		Node: "music",
	}
	iq.Payload = &payload

	data, err := xml.Marshal(iq)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	parsedIQ := IQ{}
	if err = xml.Unmarshal(data, &parsedIQ); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !xmlEqual(parsedIQ.Payload, iq.Payload) {
		t.Errorf("non matching items\n%s", cmp.Diff(parsedIQ.Payload, iq.Payload))
	}
}

func TestUnmarshalPayload(t *testing.T) {
	query := "<iq to='service.localhost' type='get' id='1'><query xmlns='jabber:iq:version'/></iq>"

	parsedIQ := IQ{}
	err := xml.Unmarshal([]byte(query), &parsedIQ)
	if err != nil {
		t.Errorf("Unmarshal(%s) returned error", query)
	}

	if parsedIQ.Payload == nil {
		t.Error("Missing payload")
	}

	namespace := parsedIQ.Payload.Namespace()
	if namespace != "jabber:iq:version" {
		t.Errorf("incorrect namespace: %s", namespace)
	}
}

func TestPayloadWithError(t *testing.T) {
	iq := `<iq xml:lang='en' to='test1@localhost/resource' from='test@localhost' type='error' id='aac1a'>
 <query xmlns='jabber:iq:version'/>
 <error code='407' type='auth'>
  <subscription-required xmlns='urn:ietf:params:xml:ns:xmpp-stanzas'/>
  <text xml:lang='en' xmlns='urn:ietf:params:xml:ns:xmpp-stanzas'>Not subscribed</text>
 </error>
</iq>`

	parsedIQ := IQ{}
	err := xml.Unmarshal([]byte(iq), &parsedIQ)
	if err != nil {
		t.Errorf("Unmarshal error: %s", iq)
		return
	}

	if parsedIQ.Error.Reason != "subscription-required" {
		t.Errorf("incorrect error value: '%s'", parsedIQ.Error.Reason)
	}
}

func TestUnknownPayload(t *testing.T) {
	iq := `<iq type="get" to="service.localhost" id="1" >
 <query xmlns="unknown:ns"/>
</iq>`
	parsedIQ := IQ{}
	err := xml.Unmarshal([]byte(iq), &parsedIQ)
	if err != nil {
		t.Errorf("Unmarshal error: %#v (%s)", err, iq)
		return
	}

	if parsedIQ.Any.XMLName.Space != "unknown:ns" {
		t.Errorf("could not extract namespace: '%s'", parsedIQ.Any.XMLName.Space)
	}
}