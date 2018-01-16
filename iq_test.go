package xmpp // import "fluux.io/xmpp"

import (
	"encoding/xml"
	"fmt"
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

func TestGenerateIq(t *testing.T) {
	iq := NewIQ("get", "admin@localhost", "test@localhost", "1", "en")
	payload := Node{
		XMLName: xml.Name{
			Space: "http://jabber.org/protocol/disco#info",
			Local: "query",
		},
		Nodes: []Node{
			{XMLName: xml.Name{
				Space: "http://jabber.org/protocol/disco#info",
				Local: "identity",
			},
				Attrs: []xml.Attr{
					{Name: xml.Name{Local: "category"}, Value: "gateway"},
					{Name: xml.Name{Local: "type"}, Value: "mqtt"},
					{Name: xml.Name{Local: "name"}, Value: "Test Gateway"},
				},
				Nodes: nil,
			}},
	}
	iq.AddPayload(&payload)
	data, err := xml.Marshal(iq)
	if err != nil {
		t.Errorf("cannot marshal xml structure")
	}

	fmt.Printf("XML Struct: %s\n", data)

	var parsedIQ = new(IQ)
	if err = xml.Unmarshal(data, parsedIQ); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	if !reflect.DeepEqual(parsedIQ.Payload[0], iq.Payload[0]) {
		t.Errorf("expecting result %+v = %+v", parsedIQ.Payload[0], iq.Payload[0])
	}

	fmt.Println("ParsedIQ", parsedIQ)
}

func TestGenerateIqNew(t *testing.T) {
	iq := NewIQ("get", "admin@localhost", "test@localhost", "1", "en")
	payload := DiscoInfo{
		XMLName: xml.Name{
			Space: "http://jabber.org/protocol/disco#info",
			Local: "query",
		},
		Identity: Identity{
			XMLName: xml.Name{
				Space: "http://jabber.org/protocol/disco#info",
				Local: "identity",
			},
			Name:     "Test Gateway",
			Category: "gateway",
			Type:     "mqtt",
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

	if !reflect.DeepEqual(parsedIQ.Payload[0], iq.Payload[0]) {
		t.Errorf("expecting result %+v = %+v", parsedIQ.Payload[0], iq.Payload[0])
	}
}
