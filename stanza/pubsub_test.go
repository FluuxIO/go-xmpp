package stanza_test

import (
	"encoding/xml"
	"errors"
	"gosrc.io/xmpp/stanza"
	"strings"
	"testing"
)

var submitFormExample = stanza.NewForm([]*stanza.Field{
	{Var: "FORM_TYPE", Type: stanza.FieldTypeHidden, ValuesList: []string{"http://jabber.org/protocol/pubsub#node_config"}},
	{Var: "pubsub#title", ValuesList: []string{"Princely Musings (Atom)"}},
	{Var: "pubsub#deliver_notifications", ValuesList: []string{"1"}},
	{Var: "pubsub#access_model", ValuesList: []string{"roster"}},
	{Var: "pubsub#roster_groups_allowed", ValuesList: []string{"friends", "servants", "courtiers"}},
	{Var: "pubsub#type", ValuesList: []string{"http://www.w3.org/2005/Atom"}},
	{
		Var:        "pubsub#notification_type",
		Type:       "list-single",
		Label:      "Specify the delivery style for event notifications",
		ValuesList: []string{"headline"},
		Options: []stanza.Option{
			{ValuesList: []string{"normal"}},
			{ValuesList: []string{"headline"}},
		},
	},
}, stanza.FormTypeSubmit)

// ***********************************
// * 6.1 Subscribe to a Node
// ***********************************

func TestNewSubRequest(t *testing.T) {
	expectedReq := "<iq type=\"set\"id=\"sub1\"to=\"pubsub.shakespeare.lit\"> " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub\"> <subscribe node=\"princely_musings\"jid=\"francisco@denmark.lit\"></subscribe>" +
		" </pubsub> </iq>"

	subInfo := stanza.SubInfo{
		Node: "princely_musings", Jid: "francisco@denmark.lit",
	}
	subR, err := stanza.NewSubRq("pubsub.shakespeare.lit", subInfo)
	if err != nil {
		t.Fatalf("failed to create a sub request: %v", err)
	}
	subR.Id = "sub1"

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}

}

func TestNewSubResp(t *testing.T) {
	response := `
<iq type="result" from="pubsub.shakespeare.lit" to="francisco@denmark.lit/barracks" id="sub1">
    <pubsub xmlns="http://jabber.org/protocol/pubsub">
        <subscription node="princely_musings" jid="francisco@denmark.lit"
            subid="ba49252aaa4f5d320c24d3766f0bdcade78c78d3" subscription="subscribed"/>
    </pubsub>
</iq>
`

	pubsub, err := getPubSubGenericPayload(response)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if pubsub.Subscription == nil {
		t.Fatalf("subscription node is nil")
	}
	if pubsub.Subscription.Node == "" ||
		pubsub.Subscription.Jid == "" ||
		pubsub.Subscription.SubId == nil ||
		pubsub.Subscription.SubStatus == "" {
		t.Fatalf("one or more of the subscription attributes was not successfully decoded")
	}

}

// ***********************************
// * 6.2 Unsubscribe from a Node
// ***********************************

func TestNewUnSubRequest(t *testing.T) {
	expectedReq := "<iq type=\"set\"id=\"unsub1\"to=\"pubsub.shakespeare.lit\"> " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub\"> " +
		"<unsubscribe node=\"princely_musings\"jid=\"francisco@denmark.lit\"></unsubscribe> </pubsub> </iq>"

	subInfo := stanza.SubInfo{
		Node: "princely_musings", Jid: "francisco@denmark.lit",
	}
	subR, err := stanza.NewUnsubRq("pubsub.shakespeare.lit", subInfo)
	if err != nil {
		t.Fatalf("failed to create an unsub request: %v", err)
	}
	subR.Id = "unsub1"

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}
	pubsub, ok := subR.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}
	if pubsub.Unsubscribe == nil {
		t.Fatalf("Unsubscribe tag should be present in sub config options request")
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestNewUnsubResp(t *testing.T) {
	response := `
<iq type="result" from="pubsub.shakespeare.lit" to="francisco@denmark.lit/barracks" id="unsub1">
    <pubsub xmlns="http://jabber.org/protocol/pubsub">
        <subscription node="princely_musings" jid="francisco@denmark.lit" subscription="none"
            subid="ba49252aaa4f5d320c24d3766f0bdcade78c78d3"/>
    </pubsub>
</iq>
`

	pubsub, err := getPubSubGenericPayload(response)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if pubsub.Subscription == nil {
		t.Fatalf("subscription node is nil")
	}
	if pubsub.Subscription.Node == "" ||
		pubsub.Subscription.Jid == "" ||
		pubsub.Subscription.SubId == nil ||
		pubsub.Subscription.SubStatus == "" {
		t.Fatalf("one or more of the subscription attributes was not successfully decoded")
	}

}

// ***************************************
// * 6.3 Configure Subscription Options
// ***************************************
func TestNewSubOptsRq(t *testing.T) {
	expectedReq := "<iq type=\"get\"id=\"options1\"to=\"pubsub.shakespeare.lit\"> " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub\"> " +
		"<options node=\"princely_musings\" jid=\"francisco@denmark.lit\"></options> </pubsub> </iq>"

	subInfo := stanza.SubInfo{
		Node: "princely_musings", Jid: "francisco@denmark.lit",
	}
	subR, err := stanza.NewSubOptsRq("pubsub.shakespeare.lit", subInfo)
	if err != nil {
		t.Fatalf("failed to create a sub options request: %v", err)
	}
	subR.Id = "options1"

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}
	if pubsub.SubOptions == nil {
		t.Fatalf("Options tag should be present in sub config options request")
	}
	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestNewNewConfOptsRsp(t *testing.T) {
	response := `
<iq type="result" from="pubsub.shakespeare.lit" to="francisco@denmark.lit/barracks" id="options1">
    <pubsub xmlns="http://jabber.org/protocol/pubsub">
        <options node="princely_musings" jid="francisco@denmark.lit">
            <x xmlns="jabber:x:data" type="form">
                <field var="FORM_TYPE" type="hidden">
                    <value>http://jabber.org/protocol/pubsub#subscribe_options</value>
                </field>
                <field var="pubsub#deliver" type="boolean" label="Enable delivery?">
                    <value>1</value>
                </field>
                <field var="pubsub#digest" type="boolean"
                    label="Receive digest notifications (approx. one per day)?">
                    <value>0</value>
                </field>
                <field var="pubsub#include_body" type="boolean"
                    label="Receive message body in addition to payload?">
                    <value>false</value>
                </field>
                <field var="pubsub#show-values" type="list-multi"
                    label="Select the presence types which are
                    allowed to receive event notifications">
                    <option label="Want to Chat">
                        <value>chat</value>
                    </option>
                    <option label="Available">
                        <value>online</value>
                    </option>
                    <option label="Away">
                        <value>away</value>
                    </option>
                    <option label="Extended Away">
                        <value>xa</value>
                    </option>
                    <option label="Do Not Disturb">
                        <value>dnd</value>
                    </option>
                    <value>chat</value>
                    <value>online</value>
                </field>
            </x>
        </options>
    </pubsub>
</iq>
`

	pubsub, err := getPubSubGenericPayload(response)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if pubsub.SubOptions == nil {
		t.Fatalf("sub options node is nil")
	}
	if pubsub.SubOptions.Form == nil {
		t.Fatalf("the response form is nil")
	}

	if len(pubsub.SubOptions.Form.Fields) != 5 {
		t.Fatalf("one or more fields in the response form could not be parsed correctly")
	}
}

// ***************************************
// * 6.3.5 Form Submission
// ***************************************
func TestNewFormSubmission(t *testing.T) {
	expectedReq := "<iq type=\"set\" id=\"options2\" to=\"pubsub.shakespeare.lit\"> " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub\"> <options node=\"princely_musings\" jid=\"francisco@denmark.lit\"> " +
		"<x xmlns=\"jabber:x:data\" type=\"submit\"> <field var=\"FORM_TYPE\" type=\"hidden\">" +
		" <value>http://jabber.org/protocol/pubsub#node_config</value> </field> <field var=\"pubsub#title\"> " +
		"<value>Princely Musings (Atom)</value> </field> <field var=\"pubsub#deliver_notifications\"> " +
		"<value>1</value> </field> <field var=\"pubsub#access_model\"> <value>roster</value> </field> " +
		"<field var=\"pubsub#roster_groups_allowed\"> <value>friends</value> <value>servants</value>" +
		" <value>courtiers</value> </field> <field var=\"pubsub#type\"> <value>http://www.w3.org/2005/Atom</value> " +
		"</field> <field var=\"pubsub#notification_type\" type=\"list-single\"label=\"Specify the delivery style for event notifications\"> " +
		"<value>headline</value> <option> <value>normal</value> </option> <option> <value>headline</value> </option> " +
		"</field> </x> </options> </pubsub> </iq>"

	subInfo := stanza.SubInfo{
		Node: "princely_musings", Jid: "francisco@denmark.lit",
	}

	subR, err := stanza.NewFormSubmission("pubsub.shakespeare.lit", subInfo, submitFormExample)
	if err != nil {
		t.Fatalf("failed to create a form submission request: %v", err)
	}
	subR.Id = "options2"

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}
	if pubsub.SubOptions == nil {
		t.Fatalf("Options tag should be present in sub config options request")
	}
	if pubsub.SubOptions.Form == nil {
		t.Fatalf("No form in form submit request !")
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

// ***************************************
// * 6.3.7 Subscribe and Configure
// ***************************************

func TestNewSubAndConfig(t *testing.T) {
	expectedReq := "<iq type=\"set\"id=\"sub1\"to=\"pubsub.shakespeare.lit\">" +
		" <pubsub xmlns=\"http://jabber.org/protocol/pubsub\"> <subscribe node=\"princely_musings\" jid=\"francisco@denmark.lit\"> " +
		"</subscribe>" +
		"<options> <x xmlns=\"jabber:x:data\" type=\"submit\"> <field var=\"FORM_TYPE\" type=\"hidden\">" +
		" <value>http://jabber.org/protocol/pubsub#node_config</value> </field> <field var=\"pubsub#title\"> " +
		"<value>Princely Musings (Atom)</value> </field> <field var=\"pubsub#deliver_notifications\"> " +
		"<value>1</value> </field> <field var=\"pubsub#access_model\"> <value>roster</value> </field> " +
		"<field var=\"pubsub#roster_groups_allowed\"> <value>friends</value> <value>servants</value>" +
		" <value>courtiers</value> </field> <field var=\"pubsub#type\"> <value>http://www.w3.org/2005/Atom</value> " +
		"</field> <field var=\"pubsub#notification_type\" type=\"list-single\"label=\"Specify the delivery style for event notifications\"> " +
		"<value>headline</value> <option> <value>normal</value> </option> <option> <value>headline</value> </option> " +
		"</field> </x> </options> </pubsub> </iq>"

	subInfo := stanza.SubInfo{
		Node: "princely_musings", Jid: "francisco@denmark.lit",
	}

	subR, err := stanza.NewSubAndConfig("pubsub.shakespeare.lit", subInfo, submitFormExample)
	if err != nil {
		t.Fatalf("failed to create a sub and config request: %v", err)
	}
	subR.Id = "sub1"

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}
	if pubsub.SubOptions == nil {
		t.Fatalf("Options tag should be present in sub config options request")
	}
	if pubsub.SubOptions.Form == nil {
		t.Fatalf("No form in form submit request !")
	}

	// The <options/> element MUST NOT possess a 'node' attribute or 'jid' attribute
	// See XEP-0060
	if pubsub.SubOptions.SubInfo.Node != "" || pubsub.SubOptions.SubInfo.Jid != "" {
		t.Fatalf("SubInfo node and jid should be empty for the options tag !")
	}
	if pubsub.Subscribe.Node == "" || pubsub.Subscribe.Jid == "" {
		t.Fatalf("SubInfo node and jid should NOT be empty for the subscribe tag !")
	}
	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestNewSubAndConfigResp(t *testing.T) {
	response := `
<iq type="result" from="pubsub.shakespeare.lit" to="francisco@denmark.lit/barracks" id="sub1">
    <pubsub xmlns="http://jabber.org/protocol/pubsub">
        <subscription node="princely_musings" jid="francisco@denmark.lit"
            subid="ba49252aaa4f5d320c24d3766f0bdcade78c78d3" subscription="subscribed"/>
        <options>
            <x xmlns="jabber:x:data" type="result">
                <field var="FORM_TYPE" type="hidden">
                    <value>http://jabber.org/protocol/pubsub#subscribe_options</value>
                </field>
                <field var="pubsub#deliver">
                    <value>1</value>
                </field>
                <field var="pubsub#digest">
                    <value>0</value>
                </field>
                <field var="pubsub#include_body">
                    <value>false</value>
                </field>
                <field var="pubsub#show-values">
                    <value>chat</value>
                    <value>online</value>
                    <value>away</value>
                </field>
            </x>
        </options>
    </pubsub>
</iq>

`

	pubsub, err := getPubSubGenericPayload(response)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if pubsub.Subscription == nil {
		t.Fatalf("sub node is nil")
	}

	if pubsub.SubOptions == nil {
		t.Fatalf("sub options node is nil")
	}
	if pubsub.SubOptions.Form == nil {
		t.Fatalf("the response form is nil")
	}

	if len(pubsub.SubOptions.Form.Fields) != 5 {
		t.Fatalf("one or more fields in the response form could not be parsed correctly")
	}
}

// ***************************************
// * 6.5.2 Requesting All List
// ***************************************
func TestNewItemsRequest(t *testing.T) {
	subR, err := stanza.NewItemsRequest("pubsub.shakespeare.lit", "princely_musings", 0)
	if err != nil {
		t.Fatalf("Could not create an items request : %s", err)
	}

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}
	if pubsub.Items == nil {
		t.Fatalf("List tag should be present to request items from a service")
	}
	if len(pubsub.Items.List) != 0 {
		t.Fatalf("There should be no items in the <items> tag to request all	 items from a service")
	}
}
func TestNewItemsResp(t *testing.T) {
	response := `
<iq type="result" from="pubsub.shakespeare.lit" to="francisco@denmark.lit/barracks" id="items2">
    <pubsub xmlns="http://jabber.org/protocol/pubsub">
        <items node="princely_musings">
            <item id="4e30f35051b7b8b42abe083742187228">
                <entry xmlns="http://www.w3.org/2005/Atom">
                    <title>Alone</title>
                    <summary> Now I am alone. O, what a rogue and peasant slave am I! </summary>
                    <link rel="alternate" type="text/html"
                        href="http://denmark.lit/2003/12/13/atom03"/>
                    <id>tag:denmark.lit,2003:entry-32396</id>
                    <published>2003-12-13T11:09:53Z</published>
                    <updated>2003-12-13T11:09:53Z</updated>
                </entry>
            </item>
            <item id="ae890ac52d0df67ed7cfdf51b644e901">
                <entry xmlns="http://www.w3.org/2005/Atom">
                    <title>Soliloquy</title>
                    <summary> To be, or not to be: that is the question: Whether 'tis nobler in the
                        mind to suffer The slings and arrows of outrageous fortune, Or to take arms
                        against a sea of troubles, And by opposing end them? </summary>
                    <link rel="alternate" type="text/html"
                        href="http://denmark.lit/2003/12/13/atom03"/>
                    <id>tag:denmark.lit,2003:entry-32397</id>
                    <published>2003-12-13T18:30:02Z</published>
                    <updated>2003-12-13T18:30:02Z</updated>
                </entry>
            </item>
        </items>
    </pubsub>
</iq>
`

	pubsub, err := getPubSubGenericPayload(response)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if pubsub.Items == nil {
		t.Fatalf("sub options node is nil")
	}
	if pubsub.Items.List == nil {
		t.Fatalf("the response form is nil")
	}

	if len(pubsub.Items.List) != 2 {
		t.Fatalf("one or more items in the response could not be parsed correctly")
	}
}

// ***************************************
// * 6.5.8 Requesting a Particular Item
// ***************************************
func TestNewSpecificItemRequest(t *testing.T) {
	expectedReq := "<iq type=\"get\" id=\"items3\"to=\"pubsub.shakespeare.lit\"> " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub\"> <items node=\"princely_musings\"> " +
		"<item id=\"ae890ac52d0df67ed7cfdf51b644e901\"></item> </items> </pubsub> </iq>"

	subR, err := stanza.NewSpecificItemRequest("pubsub.shakespeare.lit", "princely_musings", "ae890ac52d0df67ed7cfdf51b644e901")
	if err != nil {
		t.Fatalf("failed to create a specific item request: %v", err)
	}
	subR.Id = "items3"

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}
	if pubsub.Items == nil {
		t.Fatalf("List tag should be present to request items from a service")
	}
	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

// ***************************************
// * 7.1 Publish an Item to a Node
// ***************************************
func TestNewPublishItemRq(t *testing.T) {
	item := stanza.Item{
		XMLName:   xml.Name{},
		Id:        "",
		Publisher: "",
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
	}

	subR, err := stanza.NewPublishItemRq("pubsub.shakespeare.lit", "princely_musings", "bnd81g37d61f49fgn581", item)
	if err != nil {
		t.Fatalf("Could not create an item pub request : %s", err)
	}

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}

	if strings.TrimSpace(pubsub.Publish.Node) == "" {
		t.Fatalf("the <publish/> element MUST possess a 'node' attribute, specifying the NodeID of the node.")
	}
	if pubsub.Publish.Items[0].Id == "" {
		t.Fatalf("an id was provided for the item and it should be used")
	}
}

// ***************************************
// * 7.1.5 Publishing Options
// ***************************************

func TestNewPublishItemOptsRq(t *testing.T) {
	expectedReq := "<iq type=\"set\"id=\"pub1\"to=\"pubsub.shakespeare.lit\"> <pubsub xmlns=\"http://jabber.org/protocol/pubsub\"> " +
		"<publish node=\"princely_musings\"> <item id=\"ae890ac52d0df67ed7cfdf51b644e901\"> " +
		"<entry xmlns=\"http://www.w3.org/2005/Atom\"> <title>Soliloquy</title> " +
		"<summary> To be, or not to be: that is the question: Whether \"tis nobler in the mind to suffer The " +
		"slings and arrows of outrageous fortune, Or to take arms against a sea of troubles, And by opposing end them? " +
		"</summary> <link rel=\"alternate\" type=\"text/html\"href=\"http://denmark.lit/2003/12/13/atom03\"></link> " +
		"<id>tag:denmark.lit,2003:entry-32397</id> <published>2003-12-13T18:30:02Z</published> " +
		"<updated>2003-12-13T18:30:02Z</updated> </entry> </item> </publish> <publish-options> " +
		"<x xmlns=\"jabber:x:data\" type=\"submit\"> <field var=\"FORM_TYPE\" type=\"hidden\"> " +
		"<value>http://jabber.org/protocol/pubsub#publish-options</value> </field> <field var=\"pubsub#access_model\"> " +
		"<value>presence</value> </field> </x> </publish-options> </pubsub> </iq>"

	var iq stanza.IQ
	err := xml.Unmarshal([]byte(expectedReq), &iq)
	if err != nil {
		t.Fatalf("could not unmarshal example request : %s", err)
	}

	pubsub, ok := iq.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}
	if pubsub.Publish == nil {
		t.Fatalf("Publish tag is empty")
	}
	if len(pubsub.Publish.Items) != 1 {
		t.Fatalf("could not parse item properly")
	}
}

// ***************************************
// * 7.2 Delete an Item from a Node
// ***************************************

func TestNewDelItemFromNode(t *testing.T) {
	expectedReq := "<iq type=\"set\"id=\"retract1\"to=\"pubsub.shakespeare.lit\"> " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub\"> <retract node=\"princely_musings\"> " +
		"<item id=\"ae890ac52d0df67ed7cfdf51b644e901\"></item> </retract> </pubsub> </iq>"

	subR, err := stanza.NewDelItemFromNode("pubsub.shakespeare.lit", "princely_musings", "ae890ac52d0df67ed7cfdf51b644e901", nil)
	if err != nil {
		t.Fatalf("failed to create a delete item from node request: %v", err)
	}
	subR.Id = "retract1"

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated del item request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}
	if pubsub.Retract == nil {
		t.Fatalf("Retract tag should be present to del an item from a service")
	}

	if strings.TrimSpace(pubsub.Retract.Items[0].Id) == "" {
		t.Fatalf("Item id, for the item to delete, should be non empty")
	}
	if pubsub.Retract.Items[0].Any != nil {
		t.Fatalf("Item node must be empty")
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

// ***************************************
// * 8.1 Create a Node
// ***************************************

func TestNewCreateNode(t *testing.T) {
	expectedReq := "<iq type=\"set\"id=\"create1\"to=\"pubsub.shakespeare.lit\"> " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub\"> <create node=\"princely_musings\"></create> </pubsub> </iq>"

	subR, err := stanza.NewCreateNode("pubsub.shakespeare.lit", "princely_musings")
	if err != nil {
		t.Fatalf("failed to create a create node request: %v", err)
	}
	subR.Id = "create1"

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated del item request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}
	if pubsub.Create == nil {
		t.Fatalf("Create tag should be present to create a node on a service")
	}

	if strings.TrimSpace(pubsub.Create.Node) == "" {
		t.Fatalf("Expected node name to be present")
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestNewCreateNodeResp(t *testing.T) {
	response := `
<iq type="result" from="pubsub.shakespeare.lit" to="hamlet@denmark.lit/elsinore" id="create2">
    <pubsub xmlns="http://jabber.org/protocol/pubsub">
        <create node="25e3d37dabbab9541f7523321421edc5bfeb2dae"/>
    </pubsub>
</iq>
`
	pubsub, err := getPubSubGenericPayload(response)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if pubsub.Create == nil {
		t.Fatalf("create segment is nil")
	}
	if pubsub.Create.Node == "" {
		t.Fatalf("could not parse generated nodeId")
	}

}

// ***************************************
// * 8.1.3 Create and Configure a Node
// ***************************************

func TestNewCreateAndConfigNode(t *testing.T) {
	expectedReq := "<iq type=\"set\" id=\"create1\" to=\"pubsub.shakespeare.lit\" > " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub\"> <create node=\"princely_musings\"></create> " +
		"<configure> <x xmlns=\"jabber:x:data\" type=\"submit\"> <field var=\"FORM_TYPE\" type=\"hidden\" > " +
		"<value>http://jabber.org/protocol/pubsub#node_config</value> </field> <field var=\"pubsub#notify_retract\"> " +
		"<value>0</value> </field> <field var=\"pubsub#notify_sub\"> <value>0</value> </field> " +
		"<field var=\"pubsub#max_payload_size\"> <value>1028</value> </field> </x> </configure> </pubsub> </iq>"

	subR, err := stanza.NewCreateAndConfigNode("pubsub.shakespeare.lit",
		"princely_musings",
		&stanza.Form{
			Type: stanza.FormTypeSubmit,
			Fields: []*stanza.Field{
				{Var: "FORM_TYPE", Type: stanza.FieldTypeHidden, ValuesList: []string{"http://jabber.org/protocol/pubsub#node_config"}},
				{Var: "pubsub#notify_retract", ValuesList: []string{"0"}},
				{Var: "pubsub#notify_sub", ValuesList: []string{"0"}},
				{Var: "pubsub#max_payload_size", ValuesList: []string{"1028"}},
			},
		})

	if err != nil {
		t.Fatalf("failed to create a create and config node request: %v", err)
	}
	subR.Id = "create1"

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated del item request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}
	if pubsub.Create == nil {
		t.Fatalf("Create tag should be present to create a node on a service")
	}

	if strings.TrimSpace(pubsub.Create.Node) == "" {
		t.Fatalf("Expected node name to be present")
	}

	if pubsub.Configure == nil {
		t.Fatalf("Configure tag should be present to configure a node during its creation on a service")
	}

	if pubsub.Configure.Form == nil {
		t.Fatalf("Expected a form to be present, to configure the node")
	}
	if len(pubsub.Configure.Form.Fields) != 4 {
		t.Fatalf("Expected 4 elements to be present in the config form but got : %v", len(pubsub.Configure.Form.Fields))
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}

}

// ********************************
// * 5.7 Retrieve Subscriptions
// ********************************

func TestNewRetrieveAllSubsRequest(t *testing.T) {
	expected := "<iq type=\"get\" id=\"subscriptions1\" to=\"pubsub.shakespeare.lit\"> " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub\"> <subscriptions></subscriptions> </pubsub> </iq>"

	subR, err := stanza.NewRetrieveAllSubsRequest("pubsub.shakespeare.lit")
	if err != nil {
		t.Fatalf("failed to create a get all subs request: %v", err)
	}
	subR.Id = "subscriptions1"

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated del item request : %s", e)
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expected, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestRetrieveAllSubsResp(t *testing.T) {
	response := `
<iq type="result" from="pubsub.shakespeare.lit" to="francisco@denmark.lit" id="subscriptions1">
    <pubsub xmlns="http://jabber.org/protocol/pubsub">
        <subscriptions>
            <subscription node="node1" jid="francisco@denmark.lit" subscription="subscribed"/>
            <subscription node="node2" jid="francisco@denmark.lit" subscription="subscribed"/>
            <subscription node="node5" jid="francisco@denmark.lit" subscription="unconfigured"/>
            <subscription node="node6" jid="francisco@denmark.lit" subscription="subscribed"
                subid="123-abc"/>
            <subscription node="node6" jid="francisco@denmark.lit" subscription="subscribed"
                subid="004-yyy"/>
        </subscriptions>
    </pubsub>
</iq>
`
	var respIQ stanza.IQ
	err := xml.Unmarshal([]byte(response), &respIQ)

	if err != nil {
		t.Fatalf("could not unmarshal response: %s", err)
	}

	pubsub, ok := respIQ.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("umarshalled payload is not a pubsub")
	}

	if pubsub.Subscriptions == nil {
		t.Fatalf("subscriptions node is nil")
	}
	if len(pubsub.Subscriptions.List) != 5 {
		t.Fatalf("incorrect number of decoded subscriptions")
	}
}

// ********************************
// * 5.7 Retrieve Affiliations
// ********************************

func TestNewRetrieveAllAffilsRequest(t *testing.T) {
	expected := "<iq type=\"get\"id=\"affil1\"to=\"pubsub.shakespeare.lit\"> " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub\"> <affiliations></affiliations> </pubsub> </iq>"

	subR, err := stanza.NewRetrieveAllAffilsRequest("pubsub.shakespeare.lit")
	if err != nil {
		t.Fatalf("failed to create a get all affiliations request: %v", err)
	}
	subR.Id = "affil1"

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated retreive all affiliations request : %s", e)
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expected, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestRetrieveAllAffilsResp(t *testing.T) {
	response := `
<iq type="result" from="pubsub.shakespeare.lit" to="francisco@denmark.lit" id="affil1">
    <pubsub xmlns="http://jabber.org/protocol/pubsub">
        <affiliations>
            <affiliation node="node1" affiliation="owner"/>
            <affiliation node="node2" affiliation="publisher"/>
            <affiliation node="node5" affiliation="outcast"/>
            <affiliation node="node6" affiliation="owner"/>
        </affiliations>
    </pubsub>
</iq>
`
	var respIQ stanza.IQ
	err := xml.Unmarshal([]byte(response), &respIQ)

	if err != nil {
		t.Fatalf("could not unmarshal response: %s", err)
	}

	pubsub, ok := respIQ.Payload.(*stanza.PubSubGeneric)
	if !ok {
		t.Fatalf("umarshalled payload is not a pubsub")
	}

	if pubsub.Affiliations == nil {
		t.Fatalf("subscriptions node is nil")
	}
	if len(pubsub.Affiliations.List) != 4 {
		t.Fatalf("incorrect number of decoded subscriptions")
	}
}

func getPubSubGenericPayload(response string) (*stanza.PubSubGeneric, error) {
	var respIQ stanza.IQ
	err := xml.Unmarshal([]byte(response), &respIQ)

	if err != nil {
		return &stanza.PubSubGeneric{}, err
	}

	pubsub, ok := respIQ.Payload.(*stanza.PubSubGeneric)
	if !ok {
		return nil, errors.New("this iq payload is not a pubsub")
	}

	return pubsub, nil
}
