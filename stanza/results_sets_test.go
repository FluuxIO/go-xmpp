package stanza_test

import (
	"gosrc.io/xmpp/stanza"
	"testing"
)

// Limiting the number of items
func TestNewResultSetReq(t *testing.T) {
	expectedRq := "<iq id=\"q29302\" type=\"set\"> <query xmlns=\"urn:xmpp:mam:2\"> " +
		"<x type=\"submit\" xmlns=\"jabber:x:data\"> <field type=\"hidden\" var=\"FORM_TYPE\"> " +
		"<value>urn:xmpp:mam:2</value> </field> <field var=\"start\"> <value>2010-08-07T00:00:00Z</value> </field> </x> " +
		"<set xmlns=\"http://jabber.org/protocol/rsm\"> <max>10</max> </set> </query> </iq>"

	maxVal := 10
	rs := &stanza.ResultSet{
		Max: &maxVal,
	}

	// TODO when Mam is implemented
	_ = expectedRq
	_ = rs
}

func TestUnmarshalResultSeqReq(t *testing.T) {
	// TODO when Mam is implemented

}
