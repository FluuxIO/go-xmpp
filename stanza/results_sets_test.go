package stanza_test

import (
	"encoding/xml"
	"gosrc.io/xmpp/stanza"
	"testing"
)

// Limiting the number of items
func TestNewResultSetReq(t *testing.T) {
	expectedRq := "<iq type=\"set\" id=\"limit1\" to=\"users.jabber.org\"> <query xmlns=\"jabber:iq:search\"> " +
		"<nick>Pete</nick> <set xmlns=\"http://jabber.org/protocol/rsm\"> <max>10</max> </set> </query> </iq>"

	items := []stanza.Node{
		{
			XMLName: xml.Name{Local: "nick"},
			Content: "Pete",
		},
	}

	maxVal := 10
	rs := &stanza.ResultSet{
		Max: &maxVal,
	}

	rq, err := stanza.NewResultSetReq("users.jabber.org", "jabber:iq:search", items, rs)
	if err != nil {
		t.Fatalf("failed to build the result set request : %v", err)
	}
	rq.Id = "limit1"

	data, err := xml.Marshal(rq)
	if err := compareMarshal(expectedRq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestUnmarshalResultSeqReq(t *testing.T) {
	//expectedRq := "<iq type=\"set\" id=\"limit1\" to=\"users.jabber.org\"> <query xmlns=\"jabber:iq:search\"> " +
	//	"<nick>Pete</nick> <set xmlns=\"http://jabber.org/protocol/rsm\"> <max>10</max> </set> </query> </iq>"
	//var uReq stanza.IQ
	//err := xml.Unmarshal([]byte(expectedRq), &uReq)
	//if err != nil {
	//	t.Fatalf(err.Error())
	//}
	//items := []stanza.Node{
	//	{
	//		XMLName: xml.Name{Local: "nick"},
	//		Content: "Pete",
	//	},
	//}
	//
	//maxVal := 10
	//rs := &stanza.ResultSet{
	//	XMLName: xml.Name{Local: "set", Space: "http://jabber.org/protocol/rsm"},
	//	Max:     &maxVal,
	//}
	//
	//rq, err := stanza.NewResultSetReq("users.jabber.org", "jabber:iq:search", items, rs)
	//if err != nil {
	//	t.Fatalf("failed to build the result set request : %v", err)
	//}
	//rq.Id = "limit1"
	//
	//// Namespace is unmarshalled as of parent for nodes in the payload. To DeepEqual, we need to set the namespace in
	//// the "expectedRq"
	//n, ok := rq.Payload.(*stanza.QuerySet)
	//if !ok {
	//	t.Fatalf("payload is not a query set: %v", ok)
	//}
	//n.Nodes[0].XMLName.Space = stanza.NSQuerySet
	//
	//data, err := xml.Marshal(rq)
	//fmt.Println(string(data))
	//if !reflect.DeepEqual(rq, uReq) {
	//	t.Fatalf("nope")
	//}

}
