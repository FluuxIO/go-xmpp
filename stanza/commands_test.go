package stanza_test

import (
	"encoding/xml"
	"gosrc.io/xmpp/stanza"
	"testing"
)

func TestMarshalCommands(t *testing.T) {
	input := "<command xmlns=\"http://jabber.org/protocol/commands\" node=\"list\" " +
		"sessionid=\"list:20020923T213616Z-700\" status=\"completed\"><x xmlns=\"jabber:x:data\" " +
		"type=\"result\"><title>Available Services</title><reported xmlns=\"jabber:x:data\"><field var=\"service\" " +
		"label=\"Service\"></field><field var=\"runlevel-1\" label=\"Single-User mode\">" +
		"</field><field var=\"runlevel-2\" label=\"Non-Networked Multi-User mode\"></field><field var=\"runlevel-3\" " +
		"label=\"Full Multi-User mode\"></field><field var=\"runlevel-5\" label=\"X-Window mode\"></field></reported>" +
		"<item xmlns=\"jabber:x:data\"><field var=\"service\"><value>httpd</value></field><field var=\"runlevel-1\">" +
		"<value>off</value></field><field var=\"runlevel-2\"><value>off</value></field><field var=\"runlevel-3\">" +
		"<value>on</value></field><field var=\"runlevel-5\"><value>on</value></field></item>" +
		"<item xmlns=\"jabber:x:data\"><field var=\"service\"><value>postgresql</value></field>" +
		"<field var=\"runlevel-1\"><value>off</value></field><field var=\"runlevel-2\"><value>off</value></field>" +
		"<field var=\"runlevel-3\"><value>on</value></field><field var=\"runlevel-5\"><value>on</value></field></item>" +
		"<item xmlns=\"jabber:x:data\"><field var=\"service\"><value>jabberd</value></field><field var=\"runlevel-1\">" +
		"<value>off</value></field><field var=\"runlevel-2\"><value>off</value></field><field var=\"runlevel-3\">" +
		"<value>on</value></field><field var=\"runlevel-5\"><value>on</value></field></item></x></command>"
	var c stanza.Command
	err := xml.Unmarshal([]byte(input), &c)

	if err != nil {
		t.Fatalf("failed to unmarshal initial input")
	}

	data, err := xml.Marshal(c)
	if err != nil {
		t.Fatalf("failed to marshal unmarshalled input")
	}

	if err := compareMarshal(input, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}
