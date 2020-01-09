package stanza

import (
	"encoding/xml"
	"strings"
	"testing"
)

const (
	formSubmit = "<pubsub xmlns=\"http://jabber.org/protocol/pubsub#owner\">" +
		"<configure node=\"princely_musings\">" +
		"<x xmlns=\"jabber:x:data\" type=\"submit\">" +
		"<field var=\"FORM_TYPE\" type=\"hidden\">" +
		"<value>http://jabber.org/protocol/pubsub#node_config</value>" +
		"</field>" +
		"<field var=\"pubsub#title\">" +
		"<value>Princely Musings (Atom)</value>" +
		"</field>" +
		"<field var=\"pubsub#deliver_notifications\">" +
		"<value>1</value>" +
		"</field>" +
		"<field var=\"pubsub#access_model\">" +
		"<value>roster</value>" +
		"</field>" +
		"<field var=\"pubsub#roster_groups_allowed\">" +
		"<value>friends</value>" +
		"<value>servants</value>" +
		"<value>courtiers</value>" +
		"</field>" +
		"<field var=\"pubsub#type\">" +
		"<value>http://www.w3.org/2005/Atom</value>" +
		"</field>" +
		"<field var=\"pubsub#notification_type\" type=\"list-single\"" +
		"label=\"Specify the delivery style for event notifications\">" +
		"<value>headline</value>" +
		"<option>" +
		"<value>normal</value>" +
		"</option>" +
		"<option>" +
		"<value>headline</value>" +
		"</option>" +
		"</field>" +
		"</x>" +
		"</configure>" +
		"</pubsub>"

	clientJid   = "hamlet@denmark.lit/elsinore"
	serviceJid  = "pubsub.shakespeare.lit"
	iqId        = "config1"
	serviceNode = "princely_musings"
)

func TestMarshalFormSubmit(t *testing.T) {
	formIQ := NewIQ(Attrs{From: clientJid, To: serviceJid, Id: iqId, Type: IQTypeSet})
	formIQ.Payload = &PubSubOwner{
		OwnerUseCase: &ConfigureOwner{
			Node: serviceNode,
			Form: &Form{
				Type: FormTypeSubmit,
				Fields: []Field{
					{Var: "FORM_TYPE", Type: FieldTypeHidden, ValuesList: []string{"http://jabber.org/protocol/pubsub#node_config"}},
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
						Options: []Option{
							{ValuesList: []string{"normal"}},
							{ValuesList: []string{"headline"}},
						},
					},
				},
			},
		},
	}
	b, err := xml.Marshal(formIQ.Payload)
	if err != nil {
		t.Fatalf("Could not marshal formIQ : %v", err)
	}

	if strings.ReplaceAll(string(b), " ", "") != strings.ReplaceAll(formSubmit, " ", "") {
		t.Fatalf("Expected formIQ and marshalled one are different.\nExepected : %s\nMarshalled : %s", formSubmit, string(b))
	}

}

func TestUnmarshalFormSubmit(t *testing.T) {
	var f PubSubOwner
	mErr := xml.Unmarshal([]byte(formSubmit), &f)
	if mErr != nil {
		t.Fatalf("failed to unmarshal formSubmit ! %s", mErr)
	}

	data, err := xml.Marshal(&f)
	if err != nil {
		t.Fatalf("failed to marshal formSubmit")
	}

	if strings.ReplaceAll(string(data), " ", "") != strings.ReplaceAll(formSubmit, " ", "") {
		t.Fatalf("failed unmarshal/marshal for formSubmit : %s\n%s", string(data), formSubmit)
	}
}
