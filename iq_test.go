package xmpp // import "fluux.io/xmpp"

import (
	"encoding/xml"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalIqs(t *testing.T) {
	//var cs1 = new(iot.ControlSet)
	var tests = []struct {
		iqString string
		parsedIQ IQ
	}{
		{"<iq id=\"1\" type=\"set\" to=\"test@localhost\"/>", IQ{XMLName: xml.Name{Space: "", Local: "iq"}, PacketAttrs: PacketAttrs{To: "test@localhost", Type: "set", Id: "1"}}},
		//{"<iq xmlns=\"jabber:client\" id=\"2\" type=\"set\" to=\"test@localhost\" from=\"server\"><set xmlns=\"urn:xmpp:iot:control\"/></iq>", IQ{XMLName: xml.Name{Space: "jabber:client", Local: "iq"}, PacketAttrs: PacketAttrs{To: "test@localhost", From: "server", Type: "set", Id: "2"}, Payload: cs1}},
	}

	for _, test := range tests {
		var parsedIQ = new(IQ)
		err := xml.Unmarshal([]byte(test.iqString), parsedIQ)
		if err != nil {
			t.Errorf("Unmarshal(%s) returned error", test.iqString)
		}
		if !reflect.DeepEqual(parsedIQ, &test.parsedIQ) {
			t.Errorf("Unmarshal(%s) expecting result %+v = %+v", test.iqString, parsedIQ, &test.parsedIQ)
		}
	}
}

func TestGenerateIq(t *testing.T) {
	iq := NewIQ("result", "admin@localhost", "test@localhost", "1", "en")
	payload := DiscoInfo{
		Identity: Identity{
			Name:     "Test Gateway",
			Category: "gateway",
			Type:     "mqtt",
		},
		Features: []Feature{
			{Var: "http://jabber.org/protocol/disco#info"},
			{Var: "http://jabber.org/protocol/disco#item"},
		},
	}
	iq.AddPayload(&payload)

	data, err := xml.Marshal(iq)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	var parsedIQ = new(IQ)
	if err = xml.Unmarshal(data, parsedIQ); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !xmlEqual(parsedIQ.Payload, iq.Payload) {
		t.Errorf("non matching items\n%s", cmp.Diff(parsedIQ.Payload, iq.Payload))
	}
}

// Compare iq structure but ignore empty namespace as they are set properly on
// marshal / unmarshal. There is no need to manage them on the manually
// crafted structure.
func xmlEqual(x, y interface{}) bool {
	alwaysEqual := cmp.Comparer(func(_, _ interface{}) bool { return true })
	opts := cmp.Options{
		cmp.FilterValues(func(x, y interface{}) bool {
			xx, xok := x.(xml.Name)
			yy, yok := y.(xml.Name)
			if xok && yok {
				zero := xml.Name{}
				if xx == zero || yy == zero {
					return true
				}
			}
			return false
		}, alwaysEqual),
	}

	return cmp.Equal(x, y, opts)
}
