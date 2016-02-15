package xmpp

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestUnmarshalIqs(t *testing.T) {
	var tests = []struct {
		iqString string
		parsedIQ ClientIQ
	}{
		{"<iq id=\"1\" type=\"set\" to=\"test@localhost\"/>", ClientIQ{XMLName: xml.Name{Space: "", Local: "iq"}, Packet: Packet{To: "test@localhost", Type: "set", Id: "1"}}},
	}

	for _, test := range tests {
		var parsedIQ = new(ClientIQ)
		err := xml.Unmarshal([]byte(test.iqString), parsedIQ)
		if err != nil {
			t.Errorf("Unmarshal(%s) returned error", test.iqString)
		}
		if !reflect.DeepEqual(parsedIQ, &test.parsedIQ) {
			t.Errorf("Unmarshal(%s) expecting result %+v = %+v", test.iqString, parsedIQ, &test.parsedIQ)
		}
	}
}
