package stanza

import (
	"encoding/xml"
	"errors"
	"strings"
)

type PubSubOwner struct {
	XMLName      xml.Name `xml:"http://jabber.org/protocol/pubsub#owner pubsub"`
	OwnerUseCase OwnerUseCase
	// Result sets
	ResultSet *ResultSet `xml:"set,omitempty"`
}

func (pso *PubSubOwner) Namespace() string {
	return pso.XMLName.Space
}

func (pso *PubSubOwner) GetSet() *ResultSet {
	return pso.ResultSet
}

type OwnerUseCase interface {
	UseCase() string
}

type AffiliationsOwner struct {
	XMLName      xml.Name           `xml:"affiliations"`
	Affiliations []AffiliationOwner `xml:"affiliation,omitempty"`
	Node         string             `xml:"node,attr"`
}

func (AffiliationsOwner) UseCase() string {
	return "affiliations"
}

type AffiliationOwner struct {
	XMLName           xml.Name `xml:"affiliation"`
	AffiliationStatus string   `xml:"affiliation,attr"`
	Jid               string   `xml:"jid,attr"`
}

const (
	AffiliationStatusMember      = "member"
	AffiliationStatusNone        = "none"
	AffiliationStatusOutcast     = "outcast"
	AffiliationStatusOwner       = "owner"
	AffiliationStatusPublisher   = "publisher"
	AffiliationStatusPublishOnly = "publish-only"
)

type ConfigureOwner struct {
	XMLName xml.Name `xml:"configure"`
	Node    string   `xml:"node,attr,omitempty"`
	Form    *Form    `xml:"x,omitempty"`
}

func (*ConfigureOwner) UseCase() string {
	return "configure"
}

type DefaultOwner struct {
	XMLName xml.Name `xml:"default"`
	Form    *Form    `xml:"x,omitempty"`
}

func (*DefaultOwner) UseCase() string {
	return "default"
}

type DeleteOwner struct {
	XMLName       xml.Name       `xml:"delete"`
	RedirectOwner *RedirectOwner `xml:"redirect,omitempty"`
	Node          string         `xml:"node,attr,omitempty"`
}

func (*DeleteOwner) UseCase() string {
	return "delete"
}

type RedirectOwner struct {
	XMLName xml.Name `xml:"redirect"`
	URI     string   `xml:"uri,attr"`
}

type PurgeOwner struct {
	XMLName xml.Name `xml:"purge"`
	Node    string   `xml:"node,attr"`
}

func (*PurgeOwner) UseCase() string {
	return "purge"
}

type SubscriptionsOwner struct {
	XMLName       xml.Name            `xml:"subscriptions"`
	Subscriptions []SubscriptionOwner `xml:"subscription"`
	Node          string              `xml:"node,attr"`
}

func (*SubscriptionsOwner) UseCase() string {
	return "subscriptions"
}

type SubscriptionOwner struct {
	SubscriptionStatus string `xml:"subscription"`
	Jid                string `xml:"jid,attr"`
}

const (
	SubscriptionStatusNone         = "none"
	SubscriptionStatusPending      = "pending"
	SubscriptionStatusSubscribed   = "subscribed"
	SubscriptionStatusUnconfigured = "unconfigured"
)

// NewConfigureNode creates a request to configure a node on the given service.
// A form will be returned by the service, to which the user must respond using for instance the NewFormSubmission function.
// See 8.2 Configure a Node
func NewConfigureNode(serviceId, nodeName string) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeGet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubOwner{
		OwnerUseCase: &ConfigureOwner{Node: nodeName},
	}
	return iq, nil
}

// NewDelNode creates a request to delete node "nodeID" from the "serviceId" service
// See 8.4 Delete a Node
func NewDelNode(serviceId, nodeID string) (*IQ, error) {
	if strings.TrimSpace(nodeID) == "" {
		return nil, errors.New("cannot delete a node without a target node ID")
	}
	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubOwner{
		OwnerUseCase: &DeleteOwner{Node: nodeID},
	}
	return iq, nil
}

// NewPurgeAllItems creates a new purge request for the "nodeId" node, at "serviceId" service
// See 8.5 Purge All Node Items
func NewPurgeAllItems(serviceId, nodeId string) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubOwner{
		OwnerUseCase: &PurgeOwner{Node: nodeId},
	}
	return iq, nil
}

// NewRequestDefaultConfig build a request to ask the service for the default config of its nodes
// See 8.3 Request Default Node Configuration Options
func NewRequestDefaultConfig(serviceId string) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeGet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubOwner{
		OwnerUseCase: &DefaultOwner{},
	}
	return iq, nil
}

// NewApproveSubRequest creates a new sub approval response to a request from the service to the owner of the node
// In order to approve the request, the owner shall submit the form and set the "pubsub#allow" field to a value of "1" or "true"
// For tracking purposes the message MUST reflect the 'id' attribute originally provided in the request.
// See 8.6 Manage Subscription Requests
func NewApproveSubRequest(serviceId, reqID string, apprForm *Form) (Message, error) {
	if serviceId == "" {
		return Message{}, errors.New("need a target service serviceId send approval serviceId")
	}
	if reqID == "" {
		return Message{}, errors.New("the request ID is empty but must be used for the approval")
	}
	if apprForm == nil {
		return Message{}, errors.New("approval form is nil")
	}
	apprMess := NewMessage(Attrs{To: serviceId})
	apprMess.Extensions = []MsgExtension{apprForm}
	apprMess.Id = reqID

	return apprMess, nil
}

// NewGetPendingSubRequests creates a new request for all pending subscriptions to all their nodes at a service
// This feature MUST be implemented using the Ad-Hoc Commands (XEP-0050) protocol
// 8.7 Process Pending Subscription Requests
func NewGetPendingSubRequests(serviceId string) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &Command{
		//  the command name ('node' attribute of the command element) MUST have a value of "http://jabber.org/protocol/pubsub#get-pending"
		Node:   "http://jabber.org/protocol/pubsub#get-pending",
		Action: CommandActionExecute,
	}
	return iq, nil
}

// NewGetPendingSubRequests creates a new request for all pending subscriptions to be approved on a given node
// Upon receiving the data form for managing subscription requests, the owner then MAY request pending subscription
// approval requests for a given node.
// See 8.7.4 Per-Node Request
func NewApprovePendingSubRequest(serviceId, sessionId, nodeId string) (*IQ, error) {
	if sessionId == "" {
		return nil, errors.New("the sessionId must be maintained for the command")
	}

	form := &Form{
		Type:   FormTypeSubmit,
		Fields: []*Field{{Var: "pubsub#node", ValuesList: []string{nodeId}}},
	}
	data, err := xml.Marshal(form)
	if err != nil {
		return nil, err
	}
	var n Node
	err = xml.Unmarshal(data, &n)
	if err != nil {
		return nil, err
	}

	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &Command{
		//  the command name ('node' attribute of the command element) MUST have a value of "http://jabber.org/protocol/pubsub#get-pending"
		Node:           "http://jabber.org/protocol/pubsub#get-pending",
		Action:         CommandActionExecute,
		SessionId:      sessionId,
		CommandElement: &n,
	}
	return iq, nil
}

// NewSubListRequest creates a request to list subscriptions of the client, for all nodes at the service.
// It's a Get type IQ
// 8.8.1 Retrieve Subscriptions
func NewSubListRqPl(serviceId, nodeID string) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeGet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubOwner{
		OwnerUseCase: &SubscriptionsOwner{Node: nodeID},
	}
	return iq, nil
}

func NewSubsForEntitiesRequest(serviceId, nodeID string, subs []SubscriptionOwner) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubOwner{
		OwnerUseCase: &SubscriptionsOwner{Node: nodeID, Subscriptions: subs},
	}
	return iq, nil
}

// NewModifAffiliationRequest creates a request to either modify one or more affiliations, or delete one or more affiliations
// 8.9.2 Modify Affiliation & 8.9.2.4 Multiple Simultaneous Modifications & 8.9.3 Delete an Entity (just set the status to "none")
func NewModifAffiliationRequest(serviceId, nodeID string, newAffils []AffiliationOwner) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubOwner{
		OwnerUseCase: &AffiliationsOwner{
			Node:         nodeID,
			Affiliations: newAffils,
		},
	}
	return iq, nil
}

// NewAffiliationListRequest creates a request to list all affiliated entities
// See 8.9.1 Retrieve List List
func NewAffiliationListRequest(serviceId, nodeID string) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeGet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubOwner{
		OwnerUseCase: &AffiliationsOwner{
			Node: nodeID,
		},
	}
	return iq, nil
}

// NewFormSubmission builds a form submission pubsub IQ, in the Owner namespace
// This is typically used to respond to a form issued by the server when configuring a node.
// See 8.2.4 Form Submission
func NewFormSubmissionOwner(serviceId, nodeName string, fields []*Field) (*IQ, error) {
	if serviceId == "" || nodeName == "" {
		return nil, errors.New("serviceId and nodeName must be filled for this request to be valid")
	}

	submitConf, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	submitConf.Payload = &PubSubOwner{
		OwnerUseCase: &ConfigureOwner{
			Node: nodeName,
			Form: NewForm(fields,
				FormTypeSubmit)},
	}

	return submitConf, nil
}

// GetFormFields gets the fields from a form in a IQ stanza of type result, as a map.
// Key is the "var" attribute of the field, and field is the value.
// The user can then select and modify the fields they want to alter, and submit a new form to the service using the
// NewFormSubmission function to build the IQ.
// TODO : remove restriction on IQ type ?
func (iq *IQ) GetFormFields() (map[string]*Field, error) {
	if iq.Type != IQTypeResult {
		return nil, errors.New("this IQ is not a result type IQ. Cannot extract the form from it")
	}
	switch payload := iq.Payload.(type) {
	// We support IOT Control IQ
	case *PubSubGeneric:
		fieldMap := make(map[string]*Field)
		for _, elt := range payload.Configure.Form.Fields {
			fieldMap[elt.Var] = elt
		}
		return fieldMap, nil
	case *PubSubOwner:
		fieldMap := make(map[string]*Field)
		co, ok := payload.OwnerUseCase.(*ConfigureOwner)
		if !ok {
			return nil, errors.New("this IQ does not contain a PubSub payload with a configure tag for the owner namespace")
		}
		for _, elt := range co.Form.Fields {
			fieldMap[elt.Var] = elt
		}
		return fieldMap, nil

	case *Command:
		fieldMap := make(map[string]*Field)
		co, ok := payload.CommandElement.(*Form)
		if !ok {
			return nil, errors.New("this IQ does not contain a command payload with a form")
		}
		for _, elt := range co.Fields {
			fieldMap[elt.Var] = elt
		}
		return fieldMap, nil
	default:
		if iq.Any != nil {
			fieldMap := make(map[string]*Field)
			if iq.Any.XMLName.Local != "command" {
				return nil, errors.New("this IQ does not contain a form")
			}

			for _, nde := range iq.Any.Nodes {
				if nde.XMLName.Local == "x" {
					for _, n := range nde.Nodes {
						if n.XMLName.Local == "field" {
							f := Field{}
							data, err := xml.Marshal(n)
							if err != nil {
								continue
							}
							err = xml.Unmarshal(data, &f)
							if err == nil {
								fieldMap[f.Var] = &f
							}
						}
					}
				}
			}
			return fieldMap, nil
		}
		return nil, errors.New("this IQ does not contain a form")
	}
}

func (pso *PubSubOwner) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	pso.XMLName = start.Name
	// decode inner elements
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch tt := t.(type) {

		case xml.StartElement:
			// Decode sub-elements
			var err error
			switch tt.Name.Local {

			case "affiliations":
				aff := AffiliationsOwner{}
				err = d.DecodeElement(&aff, &tt)
				pso.OwnerUseCase = &aff
			case "configure":
				co := ConfigureOwner{}
				err = d.DecodeElement(&co, &tt)
				pso.OwnerUseCase = &co
			case "default":
				def := DefaultOwner{}
				err = d.DecodeElement(&def, &tt)
				pso.OwnerUseCase = &def
			case "delete":
				del := DeleteOwner{}
				err = d.DecodeElement(&del, &tt)
				pso.OwnerUseCase = &del
			case "purge":
				pu := PurgeOwner{}
				err = d.DecodeElement(&pu, &tt)
				pso.OwnerUseCase = &pu
			case "subscriptions":
				subs := SubscriptionsOwner{}
				err = d.DecodeElement(&subs, &tt)
				pso.OwnerUseCase = &subs
				if err != nil {
					return err
				}
			}
			if err != nil {
				return err
			}
		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}
	}
}

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{Space: "http://jabber.org/protocol/pubsub#owner", Local: "pubsub"}, PubSubOwner{})
}
