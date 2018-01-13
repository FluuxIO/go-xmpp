package xmpp // import "fluux.io/xmpp"

import (
	"encoding/xml"
	"reflect"
	"testing"
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
