package stanza_test

import (
	"encoding/xml"
	"gosrc.io/xmpp/stanza"
	"strings"
	"testing"
)

func TestDecodeMsgEvent(t *testing.T) {
	str := `<message from='pubsub.shakespeare.lit' to='francisco@denmark.lit' id='foo'>
	 <event xmlns='http://jabber.org/protocol/pubsub#event'>
	   <items node='princely_musings'>
	     <item id='ae890ac52d0df67ed7cfdf51b644e901'>
	       <entry xmlns='http://www.w3.org/2005/Atom'>
	         <title>Soliloquy</title>
	         <summary>
	To be, or not to be: that is the question:
	Whether 'tis nobler in the mind to suffer
	The slings and arrows of outrageous fortune,
	Or to take arms against a sea of troubles,
	And by opposing end them?
	         </summary>
	         <link rel='alternate' type='text/html'
	               href='http://denmark.lit/2003/12/13/atom03'/>
	         <id>tag:denmark.lit,2003:entry-32397</id>
	         <published>2003-12-13T18:30:02Z</published>
	         <updated>2003-12-13T18:30:02Z</updated>
	       </entry>
	     </item>
	   </items>
	 </event>
	</message>
	`
	parsedMessage := stanza.Message{}
	if err := xml.Unmarshal([]byte(str), &parsedMessage); err != nil {
		t.Errorf("message receipt unmarshall error: %v", err)
		return
	}

	if parsedMessage.Body != "" {
		t.Errorf("Unexpected body: '%s'", parsedMessage.Body)
	}

	if len(parsedMessage.Extensions) < 1 {
		t.Errorf("no extension found on parsed message")
		return
	}

	switch ext := parsedMessage.Extensions[0].(type) {
	case *stanza.PubSubEvent:
		if ext.XMLName.Local != "event" {
			t.Fatalf("unexpected extension: %s:%s", ext.XMLName.Space, ext.XMLName.Local)
		}
		tmp, ok := parsedMessage.Extensions[0].(*stanza.PubSubEvent)
		if !ok {
			t.Fatalf("unexpected extension element: %s:%s", ext.XMLName.Space, ext.XMLName.Local)
		}
		ie, ok := tmp.EventElement.(*stanza.ItemsEvent)
		if !ok {
			t.Fatalf("unexpected extension element: %s:%s", ext.XMLName.Space, ext.XMLName.Local)
		}
		if ie.Items[0].Any.Nodes[0].Content != "Soliloquy" {
			t.Fatalf("could not read title ! Read this : %s", ie.Items[0].Any.Nodes[0].Content)
		}

		if len(ie.Items[0].Any.Nodes) != 6 {
			t.Fatalf("some nodes were not correctly parsed")
		}
	default:
		t.Fatalf("could not find pubsub event extension")
	}

}

func TestEncodeEvent(t *testing.T) {
	expected := "<message><event xmlns=\"http://jabber.org/protocol/pubsub#event\">" +
		"<items node=\"princely_musings\"><item id=\"ae890ac52d0df67ed7cfdf51b644e901\">" +
		"<entry xmlns=\"http://www.w3.org/2005/Atom\"><title>My pub item title</title>" +
		"<summary>My pub item content summary</summary><link rel=\"alternate\" " +
		"type=\"text/html\" href=\"http://denmark.lit/2003/12/13/atom03\">" +
		"</link><id>My pub item content ID</id><published>2003-12-13T18:30:02Z</published>" +
		"<updated>2003-12-13T18:30:02Z</updated></entry></item></items></event></message>"
	message := stanza.Message{
		Extensions: []stanza.MsgExtension{
			stanza.PubSubEvent{
				EventElement: stanza.ItemsEvent{
					Items: []stanza.ItemEvent{
						{
							Id: "ae890ac52d0df67ed7cfdf51b644e901",
							Any: &stanza.Node{
								XMLName: xml.Name{
									Space: "http://www.w3.org/2005/Atom",
									Local: "entry",
								},
								Attrs:   nil,
								Content: "",
								Nodes: []stanza.Node{
									{
										XMLName: xml.Name{Space: "", Local: "title"},
										Attrs:   nil,
										Content: "My pub item title",
										Nodes:   nil,
									},
									{
										XMLName: xml.Name{Space: "", Local: "summary"},
										Attrs:   nil,
										Content: "My pub item content summary",
										Nodes:   nil,
									},
									{
										XMLName: xml.Name{Space: "", Local: "link"},
										Attrs: []xml.Attr{
											{
												Name:  xml.Name{Space: "", Local: "rel"},
												Value: "alternate",
											},
											{
												Name:  xml.Name{Space: "", Local: "type"},
												Value: "text/html",
											},
											{
												Name:  xml.Name{Space: "", Local: "href"},
												Value: "http://denmark.lit/2003/12/13/atom03",
											},
										},
									},
									{
										XMLName: xml.Name{Space: "", Local: "id"},
										Attrs:   nil,
										Content: "My pub item content ID",
										Nodes:   nil,
									},
									{
										XMLName: xml.Name{Space: "", Local: "published"},
										Attrs:   nil,
										Content: "2003-12-13T18:30:02Z",
										Nodes:   nil,
									},
									{
										XMLName: xml.Name{Space: "", Local: "updated"},
										Attrs:   nil,
										Content: "2003-12-13T18:30:02Z",
										Nodes:   nil,
									},
								},
							},
						},
					},
					Node:    "princely_musings",
					Retract: nil,
				},
			},
		},
	}

	data, _ := xml.Marshal(message)
	if strings.TrimSpace(string(data)) != strings.TrimSpace(expected) {
		t.Errorf("event was not encoded properly : \nexpected:%s \ngot: %s", expected, string(data))
	}

}
