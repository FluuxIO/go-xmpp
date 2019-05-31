package xmpp_test // import "gosrc.io/xmpp"

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gosrc.io/xmpp"
)

func TestUnmarshalIqs(t *testing.T) {
	//var cs1 = new(iot.ControlSet)
	var tests = []struct {
		iqString string
		parsedIQ xmpp.IQ
	}{
		{"<iq id=\"1\" type=\"set\" to=\"test@localhost\"/>",
			xmpp.IQ{XMLName: xml.Name{Space: "", Local: "iq"}, PacketAttrs: xmpp.PacketAttrs{To: "test@localhost", Type: "set", Id: "1"}}},
		//{"<iq xmlns=\"jabber:client\" id=\"2\" type=\"set\" to=\"test@localhost\" from=\"server\"><set xmlns=\"urn:xmpp:iot:control\"/></iq>", IQ{XMLName: xml.Name{Space: "jabber:client", Local: "iq"}, PacketAttrs: PacketAttrs{To: "test@localhost", From: "server", Type: "set", Id: "2"}, Payload: cs1}},
	}

	for _, test := range tests {
		parsedIQ := xmpp.IQ{}
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
	iq := xmpp.NewIQ("result", "admin@localhost", "test@localhost", "1", "en")
	payload := xmpp.DiscoInfo{
		Identity: xmpp.Identity{
			Name:     "Test Gateway",
			Category: "gateway",
			Type:     "mqtt",
		},
		Features: []xmpp.Feature{
			{Var: "http://jabber.org/protocol/disco#info"},
			{Var: "http://jabber.org/protocol/disco#item"},
		},
	}
	iq.AddPayload(&payload)

	data, err := xml.Marshal(iq)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	if strings.Contains(string(data), "<error ") {
		t.Error("empty error should not be serialized")
	}

	parsedIQ := xmpp.IQ{}
	if err = xml.Unmarshal(data, &parsedIQ); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !xmlEqual(parsedIQ.Payload, iq.Payload) {
		t.Errorf("non matching items\n%s", cmp.Diff(parsedIQ.Payload, iq.Payload))
	}
}

func TestErrorTag(t *testing.T) {
	xError := xmpp.Err{
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

	parsedError := xmpp.Err{}
	if err = xml.Unmarshal(data, &parsedError); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !xmlEqual(parsedError, xError) {
		t.Errorf("non matching items\n%s", cmp.Diff(parsedError, xError))
	}
}

func TestDiscoItems(t *testing.T) {
	iq := xmpp.NewIQ("get", "romeo@montague.net/orchard", "catalog.shakespeare.lit", "items3", "en")
	payload := xmpp.DiscoItems{
		Node: "music",
	}
	iq.AddPayload(&payload)

	data, err := xml.Marshal(iq)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	parsedIQ := xmpp.IQ{}
	if err = xml.Unmarshal(data, &parsedIQ); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !xmlEqual(parsedIQ.Payload, iq.Payload) {
		t.Errorf("non matching items\n%s", cmp.Diff(parsedIQ.Payload, iq.Payload))
	}
}
