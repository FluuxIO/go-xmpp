package stanza_test

import (
	"encoding/xml"
	"gosrc.io/xmpp/stanza"
	"testing"
)

func TestMarshalCommands(t *testing.T) {
	input := "<command xmlns=\"http://jabber.org/protocol/commands\" node=\"list\" sessionid=\"list:20020923T213616Z-700\" status=\"completed\"><x " +
		"xmlns=\"jabber:x:data\" type=\"result\"><title xmlns=\"jabber:x:data\">Available Servi" +
		"ces</title><reported xmlns=\"jabber:x:data\"><field xmlns=\"jabber:x:data\" label=\"S" +
		"ervice\" var=\"service\"></field><field xmlns=\"jabber:x:data\" label=\"Single-User mo" +
		"de\" var=\"runlevel-1\"></field><field xmlns=\"jabber:x:data\" label=\"Non-Networked M" +
		"ulti-User mode\" var=\"runlevel-2\"></field><field xmlns=\"jabber:x:data\" label=\"Ful" +
		"l Multi-User mode\" var=\"runlevel-3\"></field><field xmlns=\"jabber:x:data\" label=\"" +
		"X-Window mode\" var=\"runlevel-5\"></field></reported><item xmlns=\"jabber:x:data\"><" +
		"field xmlns=\"jabber:x:data\" var=\"service\"><value xmlns=\"jabber:x:data\">httpd</va" +
		"lue></field><field xmlns=\"jabber:x:data\" var=\"runlevel-1\"><value xmlns=\"jabber:x" +
		":data\">off</value></field><field xmlns=\"jabber:x:data\" var=\"runlevel-2\"><value x" +
		"mlns=\"jabber:x:data\">off</value></field><field xmlns=\"jabber:x:data\" var=\"runlev" +
		"el-3\"><value xmlns=\"jabber:x:data\">on</value></field><field xmlns=\"jabber:x:data" +
		"\" var=\"runlevel-5\"><value xmlns=\"jabber:x:data\">on</value></field></item><item x" +
		"mlns=\"jabber:x:data\"><field xmlns=\"jabber:x:data\" var=\"service\"><value xmlns=\"ja" +
		"bber:x:data\">postgresql</value></field><field xmlns=\"jabber:x:data\" var=\"runleve" +
		"l-1\"><value xmlns=\"jabber:x:data\">off</value></field><field xmlns=\"jabber:x:data" +
		"\" var=\"runlevel-2\"><value xmlns=\"jabber:x:data\">off</value></field><field xmlns=" +
		"\"jabber:x:data\" var=\"runlevel-3\"><value xmlns=\"jabber:x:data\">on</value></field>" +
		"<field xmlns=\"jabber:x:data\" var=\"runlevel-5\"><value xmlns=\"jabber:x:data\">on</v" +
		"alue></field></item><item xmlns=\"jabber:x:data\"><field xmlns=\"jabber:x:data\" var" +
		"=\"service\"><value xmlns=\"jabber:x:data\">jabberd</value></field><field xmlns=\"jab" +
		"ber:x:data\" var=\"runlevel-1\"><value xmlns=\"jabber:x:data\">off</value></field><fi" +
		"eld xmlns=\"jabber:x:data\" var=\"runlevel-2\"><value xmlns=\"jabber:x:data\">off</val" +
		"ue></field><field xmlns=\"jabber:x:data\" var=\"runlevel-3\"><value xmlns=\"jabber:x:" +
		"data\">on</value></field><field xmlns=\"jabber:x:data\" var=\"runlevel-5\"><value xml" +
		"ns=\"jabber:x:data\">on</value></field></item></x></command>"

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
