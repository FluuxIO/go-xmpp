package stanza_test

import (
	"encoding/xml"
	"errors"
	"gosrc.io/xmpp/stanza"
	"testing"
)

// ******************************
// * 8.2 Configure a Node
// ******************************
func TestNewConfigureNode(t *testing.T) {
	expectedReq := "<iq type=\"get\" id=\"config1\" to=\"pubsub.shakespeare.lit\" > " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub#owner\"> <configure node=\"princely_musings\"></configure> " +
		"</pubsub> </iq>"

	subR, err := stanza.NewConfigureNode("pubsub.shakespeare.lit", "princely_musings")
	if err != nil {
		t.Fatalf("failed to create a configure node request: %v", err)
	}
	subR.Id = "config1"
	if err != nil {
		t.Fatalf("Could not create request : %s", err)
	}

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubOwner)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}

	if pubsub.OwnerUseCase == nil {
		t.Fatalf("owner use case is nil")
	}

	ownrUsecase, ok := pubsub.OwnerUseCase.(*stanza.ConfigureOwner)
	if !ok {
		t.Fatalf("owner use case is not a configure tag")
	}

	if ownrUsecase.Node == "" {
		t.Fatalf("could not parse node from config tag")
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestNewConfigureNodeResp(t *testing.T) {
	response := `
	<iq from="pubsub.shakespeare.lit" id="config1" to="hamlet@denmark.lit/elsinore" type="result">
  <pubsub xmlns="http://jabber.org/protocol/pubsub#owner">
    <configure node="princely_musings">
      <x type="form" xmlns="jabber:x:data">
        <field type="hidden" var="FORM_TYPE">
          <value>http://jabber.org/protocol/pubsub#node_config</value>
        </field>
        <field label="Purge all items when the relevant publisher goes offline?" type="boolean" var="pubsub#purge_offline">
          <value>0</value>
        </field>
        <field label="Max Payload size in bytes" type="text-single" var="pubsub#max_payload_size">
          <value>1028</value>
        </field>
        <field label="When to send the last published item" type="list-single" var="pubsub#send_last_published_item">
          <option label="Never">
            <value>never</value>
          </option>
          <option label="When a new subscription is processed">
            <value>on_sub</value>
          </option>
          <option label="When a new subscription is processed and whenever a subscriber comes online">
            <value>on_sub_and_presence</value>
          </option>
          <value>never</value>
        </field>
        <field label="Deliver event notifications only to available users" type="boolean" var="pubsub#presence_based_delivery">
          <value>0</value>
        </field>
        <field label="Specify the delivery style for event notifications" type="list-single" var="pubsub#notification_type">
          <option>
            <value>normal</value>
          </option>
          <option>
            <value>headline</value>
          </option>
          <value>headline</value>
        </field>
        <field label="Specify the type of payload data to be provided at this node" type="text-single" var="pubsub#type">
          <value>http://www.w3.org/2005/Atom</value>
        </field>
        <field label="Payload XSLT" type="text-single" var="pubsub#dataform_xslt"/>
      </x>
    </configure>
  </pubsub>
</iq>
`

	pubsub, err := getPubSubOwnerPayload(response)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if pubsub.OwnerUseCase == nil {
		t.Fatalf("owner use case is nil")
	}

	ownrUsecase, ok := pubsub.OwnerUseCase.(*stanza.ConfigureOwner)
	if !ok {
		t.Fatalf("owner use case is not a configure tag")
	}

	if ownrUsecase.Form == nil {
		t.Fatalf("form is nil in the parsed config tag")
	}

	if len(ownrUsecase.Form.Fields) != 8 {
		t.Fatalf("one or more fields in the response form could not be parsed correctly")
	}
}

// *************************************************
// * 8.3 Request Default Node Configuration Options
// *************************************************

func TestNewRequestDefaultConfig(t *testing.T) {
	expectedReq := "<iq type=\"get\" id=\"def1\" to=\"pubsub.shakespeare.lit\"> " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub#owner\"> <default></default> </pubsub> </iq>"

	subR, err := stanza.NewRequestDefaultConfig("pubsub.shakespeare.lit")
	if err != nil {
		t.Fatalf("failed to create a default config request: %v", err)
	}
	subR.Id = "def1"
	if err != nil {
		t.Fatalf("Could not create request : %s", err)
	}

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubOwner)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}

	if pubsub.OwnerUseCase == nil {
		t.Fatalf("owner use case is nil")
	}

	_, ok = pubsub.OwnerUseCase.(*stanza.DefaultOwner)
	if !ok {
		t.Fatalf("owner use case is not a default tag")
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestNewRequestDefaultConfigResp(t *testing.T) {
	response := `
	<iq from="pubsub.shakespeare.lit" id="config1" to="hamlet@denmark.lit/elsinore" type="result">
 <pubsub xmlns="http://jabber.org/protocol/pubsub#owner">
   <configure node="princely_musings">
     <x type="form" xmlns="jabber:x:data">
       <field type="hidden" var="FORM_TYPE">
         <value>http://jabber.org/protocol/pubsub#node_config</value>
       </field>
       <field label="Purge all items when the relevant publisher goes offline?" type="boolean" var="pubsub#purge_offline">
         <value>0</value>
       </field>
       <field label="Max Payload size in bytes" type="text-single" var="pubsub#max_payload_size">
         <value>1028</value>
       </field>
       <field label="When to send the last published item" type="list-single" var="pubsub#send_last_published_item">
         <option label="Never">
           <value>never</value>
         </option>
         <option label="When a new subscription is processed">
           <value>on_sub</value>
         </option>
         <option label="When a new subscription is processed and whenever a subscriber comes online">
           <value>on_sub_and_presence</value>
         </option>
         <value>never</value>
       </field>
       <field label="Deliver event notifications only to available users" type="boolean" var="pubsub#presence_based_delivery">
         <value>0</value>
       </field>
       <field label="Specify the delivery style for event notifications" type="list-single" var="pubsub#notification_type">
         <option>
           <value>normal</value>
         </option>
         <option>
           <value>headline</value>
         </option>
         <value>headline</value>
       </field>
       <field label="Specify the type of payload data to be provided at this node" type="text-single" var="pubsub#type">
         <value>http://www.w3.org/2005/Atom</value>
       </field>
       <field label="Payload XSLT" type="text-single" var="pubsub#dataform_xslt"/>
     </x>
   </configure>
 </pubsub>
</iq>
`

	pubsub, err := getPubSubOwnerPayload(response)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if pubsub.OwnerUseCase == nil {
		t.Fatalf("owner use case is nil")
	}

	ownrUsecase, ok := pubsub.OwnerUseCase.(*stanza.ConfigureOwner)
	if !ok {
		t.Fatalf("owner use case is not a configure tag")
	}

	if ownrUsecase.Form == nil {
		t.Fatalf("form is nil in the parsed config tag")
	}

	if len(ownrUsecase.Form.Fields) != 8 {
		t.Fatalf("one or more fields in the response form could not be parsed correctly")
	}
}

// ***********************
// * 8.4 Delete a Node
// ***********************

func TestNewDelNode(t *testing.T) {
	expectedReq := "<iq type=\"set\" id=\"delete1\" to=\"pubsub.shakespeare.lit\" >" +
		" <pubsub xmlns=\"http://jabber.org/protocol/pubsub#owner\"> " +
		"<delete node=\"princely_musings\"></delete> </pubsub> </iq>"

	subR, err := stanza.NewDelNode("pubsub.shakespeare.lit", "princely_musings")
	if err != nil {
		t.Fatalf("failed to create a node delete request: %v", err)
	}
	subR.Id = "delete1"
	if err != nil {
		t.Fatalf("Could not create request : %s", err)
	}

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubOwner)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}

	if pubsub.OwnerUseCase == nil {
		t.Fatalf("owner use case is nil")
	}

	_, ok = pubsub.OwnerUseCase.(*stanza.DeleteOwner)
	if !ok {
		t.Fatalf("owner use case is not a delete tag")
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestNewDelNodeResp(t *testing.T) {
	response := `
	<iq id="delete1" to="pubsub.shakespeare.lit" type="set">
    <pubsub xmlns="http://jabber.org/protocol/pubsub#owner">
        <delete node="princely_musings">
            <redirect uri="xmpp:hamlet@denmark.lit"/>
        </delete>
    </pubsub>
</iq>
`

	pubsub, err := getPubSubOwnerPayload(response)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if pubsub.OwnerUseCase == nil {
		t.Fatalf("owner use case is nil")
	}

	ownrUsecase, ok := pubsub.OwnerUseCase.(*stanza.DeleteOwner)
	if !ok {
		t.Fatalf("owner use case is not a configure tag")
	}

	if ownrUsecase.RedirectOwner == nil {
		t.Fatalf("redirect is nil in the delete tag")
	}

	if ownrUsecase.RedirectOwner.URI == "" {
		t.Fatalf("could not parse redirect uri")
	}
}

// ****************************
// * 8.5 Purge All Node Items
// ****************************

func TestNewPurgeAllItems(t *testing.T) {
	expectedReq := "<iq type=\"set\" id=\"purge1\" to=\"pubsub.shakespeare.lit\"> " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub#owner\"> " +
		"<purge node=\"princely_musings\"></purge> </pubsub> </iq>"

	subR, err := stanza.NewPurgeAllItems("pubsub.shakespeare.lit", "princely_musings")
	if err != nil {
		t.Fatalf("failed to create a purge all items request: %v", err)
	}
	subR.Id = "purge1"
	if err != nil {
		t.Fatalf("Could not create request : %s", err)
	}

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubOwner)
	if !ok {
		t.Fatalf("payload is not a pubsub !")
	}

	if pubsub.OwnerUseCase == nil {
		t.Fatalf("owner use case is nil")
	}

	purge, ok := pubsub.OwnerUseCase.(*stanza.PurgeOwner)
	if !ok {
		t.Fatalf("owner use case is not a delete tag")
	}

	if purge.Node == "" {
		t.Fatalf("could not parse purge targer node")
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

// ************************************
// * 8.6 Manage Subscription Requests
// ************************************
func TestNewApproveSubRequest(t *testing.T) {
	expectedReq := "<message id=\"approve1\" to=\"pubsub.shakespeare.lit\"> " +
		"<x xmlns=\"jabber:x:data\" type=\"submit\"> <field var=\"FORM_TYPE\" type=\"hidden\"> " +
		"<value>http://jabber.org/protocol/pubsub#subscribe_authorization</value> </field> <field var=\"pubsub#subid\">" +
		" <value>123-abc</value> </field> <field var=\"pubsub#node\"> <value>princely_musings</value> </field> " +
		"<field var=\"pubsub#subscriber_jid\"> <value>horatio@denmark.lit</value> </field> <field var=\"pubsub#allow\"> " +
		"<value>true</value> </field> </x> </message>"

	apprForm := &stanza.Form{
		Type: stanza.FormTypeSubmit,
		Fields: []*stanza.Field{
			{Var: "FORM_TYPE", Type: stanza.FieldTypeHidden, ValuesList: []string{"http://jabber.org/protocol/pubsub#subscribe_authorization"}},
			{Var: "pubsub#subid", ValuesList: []string{"123-abc"}},
			{Var: "pubsub#node", ValuesList: []string{"princely_musings"}},
			{Var: "pubsub#subscriber_jid", ValuesList: []string{"horatio@denmark.lit"}},
			{Var: "pubsub#allow", ValuesList: []string{"true"}},
		},
	}

	subR, err := stanza.NewApproveSubRequest("pubsub.shakespeare.lit", "approve1", apprForm)
	if err != nil {
		t.Fatalf("failed to create a sub approval request: %v", err)
	}
	subR.Id = "approve1"
	if err != nil {
		t.Fatalf("Could not create request : %s", err)
	}

	frm, ok := subR.Extensions[0].(*stanza.Form)
	if !ok {
		t.Fatalf("extension is not a from !")
	}

	var allowField *stanza.Field

	for _, f := range frm.Fields {
		if f.Var == "pubsub#allow" {
			allowField = f
		}
	}
	if allowField == nil || allowField.ValuesList[0] != "true" {
		t.Fatalf("could not correctly parse the allow field in the response from")
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

// ********************************************
// * 8.7 Process Pending Subscription Requests
// ********************************************

func TestNewGetPendingSubRequests(t *testing.T) {
	expectedReq := "<iq type=\"set\" id=\"pending1\" to=\"pubsub.shakespeare.lit\" > " +
		"<command xmlns=\"http://jabber.org/protocol/commands\"  action=\"execute\" node=\"http://jabber.org/protocol/pubsub#get-pending\" >" +
		"</command> </iq>"

	subR, err := stanza.NewGetPendingSubRequests("pubsub.shakespeare.lit")
	if err != nil {
		t.Fatalf("failed to create a get pending subs request: %v", err)
	}
	subR.Id = "pending1"
	if err != nil {
		t.Fatalf("Could not create request : %s", err)
	}

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	command, ok := subR.Payload.(*stanza.Command)
	if !ok {
		t.Fatalf("payload is not a command !")
	}

	if command.Action != stanza.CommandActionExecute {
		t.Fatalf("command should be execute !")
	}

	if command.Node != "http://jabber.org/protocol/pubsub#get-pending" {
		t.Fatalf("command node should be http://jabber.org/protocol/pubsub#get-pending !")
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestNewGetPendingSubRequestsResp(t *testing.T) {
	response := `
	<iq from="pubsub.shakespeare.lit" id="pending1" to="hamlet@denmark.lit/elsinore" type="result">
  <command action="execute" node="http://jabber.org/protocol/pubsub#get-pending" sessionid="pubsub-get-pending:20031021T150901Z-600" status="executing" xmlns="http://jabber.org/protocol/commands">
    <x type="form" xmlns="jabber:x:data">
      <field type="hidden" var="FORM_TYPE">
        <value>http://jabber.org/protocol/pubsub#subscribe_authorization</value>
      </field>
      <field type="list-single" var="pubsub#node">
        <option>
          <value>princely_musings</value>
        </option>
        <option>
          <value>news_from_elsinore</value>
        </option>
      </field>
    </x>
  </command>
</iq>
`

	var respIQ stanza.IQ
	err := xml.Unmarshal([]byte(response), &respIQ)
	if err != nil {
		t.Fatalf("could not parse iq")
	}

	_, ok := respIQ.Payload.(*stanza.Command)
	if !ok {
		t.Fatal("this iq payload is not a command")
	}

	fMap, err := respIQ.GetFormFields()
	if err != nil || len(fMap) != 2 {
		t.Fatal("could not parse command form fields")
	}

}

// ********************************************
// * 8.7 Process Pending Subscription Requests
// ********************************************

func TestNewApprovePendingSubRequest(t *testing.T) {
	expectedReq := "<iq type=\"set\" id=\"pending2\" to=\"pubsub.shakespeare.lit\"> " +
		"<command xmlns=\"http://jabber.org/protocol/commands\" action=\"execute\"" +
		"node=\"http://jabber.org/protocol/pubsub#get-pending\"sessionid=\"pubsub-get-pending:20031021T150901Z-600\"> " +
		"<x xmlns=\"jabber:x:data\" type=\"submit\"> <field xmlns=\"jabber:x:data\" var=\"pubsub#node\"> " +
		"<value xmlns=\"jabber:x:data\">princely_musings</value> </field> </x> </command> </iq>"

	subR, err := stanza.NewApprovePendingSubRequest("pubsub.shakespeare.lit",
		"pubsub-get-pending:20031021T150901Z-600",
		"princely_musings")
	if err != nil {
		t.Fatalf("failed to create a approve pending sub request: %v", err)
	}
	subR.Id = "pending2"
	if err != nil {
		t.Fatalf("Could not create request : %s", err)
	}

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	command, ok := subR.Payload.(*stanza.Command)
	if !ok {
		t.Fatalf("payload is not a command !")
	}

	if command.Action != stanza.CommandActionExecute {
		t.Fatalf("command should be execute !")
	}

	//if command.Node != "http://jabber.org/protocol/pubsub#get-pending"{
	//	t.Fatalf("command node should be http://jabber.org/protocol/pubsub#get-pending !")
	//}
	//

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

// ********************************************
// * 8.8.1 Retrieve Subscriptions List
// ********************************************

func TestNewSubListRqPl(t *testing.T) {
	expectedReq := "<iq type=\"get\" id=\"subman1\" to=\"pubsub.shakespeare.lit\" > " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub#owner\"> " +
		"<subscriptions node=\"princely_musings\"></subscriptions> </pubsub> </iq>"

	subR, err := stanza.NewSubListRqPl("pubsub.shakespeare.lit", "princely_musings")
	if err != nil {
		t.Fatalf("failed to create a sub list request: %v", err)
	}
	subR.Id = "subman1"
	if err != nil {
		t.Fatalf("Could not create request : %s", err)
	}

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubOwner)
	if !ok {
		t.Fatalf("payload is not a pubsub in namespace owner !")
	}

	subs, ok := pubsub.OwnerUseCase.(*stanza.SubscriptionsOwner)
	if !ok {
		t.Fatalf("pubsub doesn not contain a subscriptions node !")
	}

	if subs.Node != "princely_musings" {
		t.Fatalf("subs node attribute should be princely_musings. Found %s", subs.Node)
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestNewSubListRqPlResp(t *testing.T) {
	response := `
<iq from="pubsub.shakespeare.lit" id="subman1" to="hamlet@denmark.lit/elsinore" type="result">
  <pubsub xmlns="http://jabber.org/protocol/pubsub#owner">
    <subscriptions node="princely_musings">
      <subscription jid="hamlet@denmark.lit" subscription="subscribed"></subscription>
      <subscription jid="polonius@denmark.lit" subscription="unconfigured"></subscription>
      <subscription jid="bernardo@denmark.lit" subid="123-abc" subscription="subscribed"></subscription>
      <subscription jid="bernardo@denmark.lit" subid="004-yyy" subscription="subscribed"></subscription>
    </subscriptions>
  </pubsub>
</iq>
`

	var respIQ stanza.IQ
	err := xml.Unmarshal([]byte(response), &respIQ)
	if err != nil {
		t.Fatalf("could not parse iq")
	}

	pubsub, ok := respIQ.Payload.(*stanza.PubSubOwner)
	if !ok {
		t.Fatal("this iq payload is not a command")
	}

	subs, ok := pubsub.OwnerUseCase.(*stanza.SubscriptionsOwner)
	if !ok {
		t.Fatalf("pubsub doesn not contain a subscriptions node !")
	}

	if len(subs.Subscriptions) != 4 {
		t.Fatalf("expected to find 4 subscriptions but got %d", len(subs.Subscriptions))
	}

}

// ********************************************
// * 8.9.1 Retrieve Affiliations List
// ********************************************

func TestNewAffiliationListRequest(t *testing.T) {
	expectedReq := "<iq type=\"get\" id=\"ent1\" to=\"pubsub.shakespeare.lit\" > " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub#owner\"> " +
		"<affiliations node=\"princely_musings\"></affiliations> </pubsub> </iq>"

	subR, err := stanza.NewAffiliationListRequest("pubsub.shakespeare.lit", "princely_musings")
	if err != nil {
		t.Fatalf("failed to create an affiliations list request: %v", err)
	}
	subR.Id = "ent1"
	if err != nil {
		t.Fatalf("Could not create request : %s", err)
	}

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubOwner)
	if !ok {
		t.Fatalf("payload is not a pubsub in namespace owner !")
	}

	affils, ok := pubsub.OwnerUseCase.(*stanza.AffiliationsOwner)
	if !ok {
		t.Fatalf("pubsub doesn not contain an affiliations node !")
	}

	if affils.Node != "princely_musings" {
		t.Fatalf("affils node attribute should be princely_musings. Found %s", affils.Node)
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestNewAffiliationListRequestResp(t *testing.T) {
	response := `
<iq from="pubsub.shakespeare.lit" id="ent1" to="hamlet@denmark.lit/elsinore" type="result">
  <pubsub xmlns="http://jabber.org/protocol/pubsub#owner">
    <affiliations node="princely_musings">
      <affiliation affiliation="owner" jid="hamlet@denmark.lit"/>
      <affiliation affiliation="outcast" jid="polonius@denmark.lit"/>
    </affiliations>
  </pubsub>
</iq>
`

	var respIQ stanza.IQ
	err := xml.Unmarshal([]byte(response), &respIQ)
	if err != nil {
		t.Fatalf("could not parse iq")
	}

	pubsub, ok := respIQ.Payload.(*stanza.PubSubOwner)
	if !ok {
		t.Fatal("this iq payload is not a command")
	}

	affils, ok := pubsub.OwnerUseCase.(*stanza.AffiliationsOwner)
	if !ok {
		t.Fatalf("pubsub doesn not contain an affiliations node !")
	}

	if len(affils.Affiliations) != 2 {
		t.Fatalf("expected to find 2 subscriptions but got %d", len(affils.Affiliations))
	}

}

// ********************************************
// * 8.9.2 Modify Affiliation
// ********************************************

func TestNewModifAffiliationRequest(t *testing.T) {
	expectedReq := "<iq type=\"set\" id=\"ent3\" to=\"pubsub.shakespeare.lit\" > " +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub#owner\"> <affiliations node=\"princely_musings\"> " +
		"<affiliation affiliation=\"none\" jid=\"hamlet@denmark.lit\"></affiliation> " +
		"<affiliation affiliation=\"none\" jid=\"polonius@denmark.lit\"></affiliation> " +
		"<affiliation affiliation=\"publisher\" jid=\"bard@shakespeare.lit\"></affiliation> </affiliations> </pubsub> " +
		"</iq>"

	affils := []stanza.AffiliationOwner{
		{
			AffiliationStatus: stanza.AffiliationStatusNone,
			Jid:               "hamlet@denmark.lit",
		},
		{
			AffiliationStatus: stanza.AffiliationStatusNone,
			Jid:               "polonius@denmark.lit",
		},
		{
			AffiliationStatus: stanza.AffiliationStatusPublisher,
			Jid:               "bard@shakespeare.lit",
		},
	}

	subR, err := stanza.NewModifAffiliationRequest("pubsub.shakespeare.lit", "princely_musings", affils)
	if err != nil {
		t.Fatalf("failed to create a modif affiliation request: %v", err)
	}
	subR.Id = "ent3"
	if err != nil {
		t.Fatalf("Could not create request : %s", err)
	}

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubOwner)
	if !ok {
		t.Fatalf("payload is not a pubsub in namespace owner !")
	}

	as, ok := pubsub.OwnerUseCase.(*stanza.AffiliationsOwner)
	if !ok {
		t.Fatalf("pubsub doesn not contain an affiliations node !")
	}

	if as.Node != "princely_musings" {
		t.Fatalf("affils node attribute should be princely_musings. Found %s", as.Node)
	}
	if len(as.Affiliations) != 3 {
		t.Fatalf("expected 3 affiliations, found %d", len(as.Affiliations))
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestGetFormFields(t *testing.T) {
	response := `
	<iq from="pubsub.shakespeare.lit" id="config1" to="hamlet@denmark.lit/elsinore" type="result">
  <pubsub xmlns="http://jabber.org/protocol/pubsub#owner">
    <configure node="princely_musings">
      <x type="form" xmlns="jabber:x:data">
        <field type="hidden" var="FORM_TYPE">
          <value>http://jabber.org/protocol/pubsub#node_config</value>
        </field>
        <field label="Purge all items when the relevant publisher goes offline?" type="boolean" var="pubsub#purge_offline">
          <value>0</value>
        </field>
        <field label="Max Payload size in bytes" type="text-single" var="pubsub#max_payload_size">
          <value>1028</value>
        </field>
        <field label="When to send the last published item" type="list-single" var="pubsub#send_last_published_item">
          <option label="Never">
            <value>never</value>
          </option>
          <option label="When a new subscription is processed">
            <value>on_sub</value>
          </option>
          <option label="When a new subscription is processed and whenever a subscriber comes online">
            <value>on_sub_and_presence</value>
          </option>
          <value>never</value>
        </field>
        <field label="Deliver event notifications only to available users" type="boolean" var="pubsub#presence_based_delivery">
          <value>0</value>
        </field>
        <field label="Specify the delivery style for event notifications" type="list-single" var="pubsub#notification_type">
          <option>
            <value>normal</value>
          </option>
          <option>
            <value>headline</value>
          </option>
          <value>headline</value>
        </field>
        <field label="Specify the type of payload data to be provided at this node" type="text-single" var="pubsub#type">
          <value>http://www.w3.org/2005/Atom</value>
        </field>
        <field label="Payload XSLT" type="text-single" var="pubsub#dataform_xslt"/>
      </x>
    </configure>
  </pubsub>
</iq>
`
	var iq stanza.IQ
	err := xml.Unmarshal([]byte(response), &iq)
	if err != nil {
		t.Fatalf("could not parse IQ")
	}

	fields, err := iq.GetFormFields()
	if len(fields) != 8 {
		t.Fatalf("could not correctly parse fields. Expected 8, found : %v", len(fields))
	}

}

func TestGetFormFieldsCmd(t *testing.T) {
	response := `
	<iq from="pubsub.shakespeare.lit" id="pending1" to="hamlet@denmark.lit/elsinore" type="result">
  <command action="execute" node="http://jabber.org/protocol/pubsub#get-pending" sessionid="pubsub-get-pending:20031021T150901Z-600" status="executing" xmlns="http://jabber.org/protocol/commands">
    <x type="form" xmlns="jabber:x:data">
      <field type="hidden" var="FORM_TYPE">
        <value>http://jabber.org/protocol/pubsub#subscribe_authorization</value>
      </field>
      <field type="list-single" var="pubsub#node">
        <option>
          <value>princely_musings</value>
        </option>
        <option>
          <value>news_from_elsinore</value>
        </option>
      </field>
    </x>
  </command>
</iq>
`
	var iq stanza.IQ
	err := xml.Unmarshal([]byte(response), &iq)
	if err != nil {
		t.Fatalf("could not parse IQ")
	}

	fields, err := iq.GetFormFields()
	if len(fields) != 2 {
		t.Fatalf("could not correctly parse fields. Expected 2, found : %v", len(fields))
	}

}

func TestNewFormSubmissionOwner(t *testing.T) {
	expectedReq := "<iq type=\"set\" id=\"config2\" to=\"pubsub.shakespeare.lit\">" +
		"<pubsub xmlns=\"http://jabber.org/protocol/pubsub#owner\"> <configure node=\"princely_musings\"> " +
		"<x xmlns=\"jabber:x:data\" type=\"submit\" > <field var=\"FORM_TYPE\" type=\"hidden\"> " +
		"<value>http://jabber.org/protocol/pubsub#node_config</value> </field> <field var=\"pubsub#item_expire\"> " +
		"<value>604800</value> </field> <field var=\"pubsub#access_model\"> <value>roster</value> </field> " +
		"<field var=\"pubsub#roster_groups_allowed\"> <value>friends</value> <value>servants</value> " +
		"<value>courtiers</value> </field> </x> </configure> </pubsub> </iq>"

	subR, err := stanza.NewFormSubmissionOwner("pubsub.shakespeare.lit",
		"princely_musings",
		[]*stanza.Field{
			{Var: "FORM_TYPE", Type: stanza.FieldTypeHidden, ValuesList: []string{"http://jabber.org/protocol/pubsub#node_config"}},
			{Var: "pubsub#item_expire", ValuesList: []string{"604800"}},
			{Var: "pubsub#access_model", ValuesList: []string{"roster"}},
			{Var: "pubsub#roster_groups_allowed", ValuesList: []string{"friends", "servants", "courtiers"}},
		})
	if err != nil {
		t.Fatalf("failed to create a form submission request: %v", err)
	}

	subR.Id = "config2"
	if err != nil {
		t.Fatalf("Could not create request : %s", err)
	}

	if _, e := checkMarshalling(t, subR); e != nil {
		t.Fatalf("Failed to check marshalling for generated sub request : %s", e)
	}

	pubsub, ok := subR.Payload.(*stanza.PubSubOwner)
	if !ok {
		t.Fatalf("payload is not a pubsub in namespace owner !")
	}

	conf, ok := pubsub.OwnerUseCase.(*stanza.ConfigureOwner)
	if !ok {
		t.Fatalf("pubsub does not contain a configure node !")
	}

	if conf.Form == nil {
		t.Fatalf("the form is absent from the configuration submission !")
	}
	if len(conf.Form.Fields) != 4 {
		t.Fatalf("expected 4 fields, found %d", len(conf.Form.Fields))
	}
	if len(conf.Form.Fields[3].ValuesList) != 3 {
		t.Fatalf("expected 3 values in fourth field, found %d", len(conf.Form.Fields[3].ValuesList))
	}

	data, err := xml.Marshal(subR)
	if err := compareMarshal(expectedReq, string(data)); err != nil {
		t.Fatalf(err.Error())
	}
}

func getPubSubOwnerPayload(response string) (*stanza.PubSubOwner, error) {
	var respIQ stanza.IQ
	err := xml.Unmarshal([]byte(response), &respIQ)

	if err != nil {
		return &stanza.PubSubOwner{}, err
	}

	pubsub, ok := respIQ.Payload.(*stanza.PubSubOwner)
	if !ok {
		return nil, errors.New("this iq payload is not a pubsub of the owner namespace")
	}

	return pubsub, nil
}
