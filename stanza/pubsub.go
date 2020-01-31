package stanza

import (
	"encoding/xml"
	"errors"
	"strings"
)

type PubSubGeneric struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/pubsub pubsub"`

	Create    *Create    `xml:"create,omitempty"`
	Configure *Configure `xml:"configure,omitempty"`

	Subscribe  *SubInfo    `xml:"subscribe,omitempty"`
	SubOptions *SubOptions `xml:"options,omitempty"`

	Publish        *Publish        `xml:"publish,omitempty"`
	PublishOptions *PublishOptions `xml:"publish-options"`

	Affiliations *Affiliations `xml:"affiliations,omitempty"`
	Default      *Default      `xml:"default,omitempty"`

	Items        *Items        `xml:"items,omitempty"`
	Retract      *Retract      `xml:"retract,omitempty"`
	Subscription *Subscription `xml:"subscription,omitempty"`

	Subscriptions *Subscriptions `xml:"subscriptions,omitempty"`
	// To use in responses to sub/unsub for instance
	// Subscription options
	Unsubscribe *SubInfo `xml:"unsubscribe,omitempty"`

	// Result sets
	ResultSet *ResultSet `xml:"set,omitempty"`
}

func (p *PubSubGeneric) Namespace() string {
	return p.XMLName.Space
}

func (p *PubSubGeneric) GetSet() *ResultSet {
	return p.ResultSet
}

type Affiliations struct {
	List []Affiliation `xml:"affiliation"`
	Node string        `xml:"node,attr,omitempty"`
}

type Affiliation struct {
	AffiliationStatus string `xml:"affiliation"`
	Node              string `xml:"node,attr"`
}

type Create struct {
	Node string `xml:"node,attr,omitempty"`
}

type SubOptions struct {
	SubInfo
	Form *Form `xml:"x"`
}

type Configure struct {
	Form *Form `xml:"x"`
}
type Default struct {
	Node string `xml:"node,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
	Form *Form  `xml:"x"`
}

type Subscribe struct {
	XMLName xml.Name `xml:"subscribe"`
	SubInfo
}
type Unsubscribe struct {
	XMLName xml.Name `xml:"unsubscribe"`
	SubInfo
}

// SubInfo represents information about a subscription
// Node is the node related to the subscription
// Jid is the subscription JID of the subscribed entity
// SubID is the subscription ID
type SubInfo struct {
	Node string `xml:"node,attr,omitempty"`
	Jid  string `xml:"jid,attr,omitempty"`
	// Sub ID is optional
	SubId *string `xml:"subid,attr,omitempty"`
}

// validate checks if a node and a jid are present in the sub info, and if this jid is valid.
func (si *SubInfo) validate() error {
	// Requests MUST contain a valid JID
	if _, err := NewJid(si.Jid); err != nil {
		return err
	}
	// SubInfo must contain both a valid JID and a node. See XEP-0060
	if strings.TrimSpace(si.Node) == "" {
		return errors.New("SubInfo must contain the node AND the subscriber JID in subscription config options requests")
	}
	return nil
}

// Handles the "5.6 Retrieve Subscriptions" of XEP-0060
type Subscriptions struct {
	XMLName xml.Name       `xml:"subscriptions"`
	List    []Subscription `xml:"subscription,omitempty"`
}

// Handles the "5.6 Retrieve Subscriptions" and the 6.1 Subscribe to a Node and so on of XEP-0060
type Subscription struct {
	SubStatus string `xml:"subscription,attr,omitempty"`
	SubInfo   `xml:",omitempty"`
	// Seems like we can't marshal a self-closing tag for now : https://github.com/golang/go/issues/21399
	// subscribe-options should be like this as per XEP-0060:
	//    <subscribe-options>
	//        <required/>
	//    </subscribe-options>
	// Used to indicate if configuration options is required.
	Required *struct{}
}

type PublishOptions struct {
	XMLName xml.Name `xml:"publish-options"`
	Form    *Form
}

type Publish struct {
	XMLName xml.Name `xml:"publish"`
	Node    string   `xml:"node,attr"`
	Items   []Item   `xml:"item,omitempty"` // xsd says there can be many. See also 12.10 Batch Processing of XEP-0060
}

type Items struct {
	List     []Item `xml:"item,omitempty"`
	MaxItems int    `xml:"max_items,attr,omitempty"`
	Node     string `xml:"node,attr"`
	SubId    string `xml:"subid,attr,omitempty"`
}

type Item struct {
	XMLName   xml.Name `xml:"item"`
	Id        string   `xml:"id,attr,omitempty"`
	Publisher string   `xml:"publisher,attr,omitempty"`
	Any       *Node    `xml:",any"`
}

type Retract struct {
	XMLName xml.Name `xml:"retract"`
	Node    string   `xml:"node,attr"`
	Notify  *bool    `xml:"notify,attr,omitempty"`
	Items   []Item   `xml:"item"`
}

type PubSubOption struct {
	XMLName xml.Name `xml:"jabber:x:data options"`
	Form    `xml:"x"`
}

// NewSubRq builds a subscription request to a node at the given service.
// It's a Set type IQ.
// Information about the subscription and the requester are separated. subInfo contains information about the subscription.
// 6.1 Subscribe to a Node
func NewSubRq(serviceId string, subInfo SubInfo) (*IQ, error) {
	if e := subInfo.validate(); e != nil {
		return nil, e
	}

	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		Subscribe: &subInfo,
	}
	return iq, nil
}

// NewUnsubRq builds an unsub request to a node at the given service.
// It's a Set type IQ
// Information about the subscription and the requester are separated. subInfo contains information about the subscription.
// 6.2 Unsubscribe from a Node
func NewUnsubRq(serviceId string, subInfo SubInfo) (*IQ, error) {
	if e := subInfo.validate(); e != nil {
		return nil, e
	}

	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		Unsubscribe: &subInfo,
	}
	return iq, nil
}

// NewSubOptsRq builds a request for the subscription options.
// It's a Get type IQ
// Information about the subscription and the requester are separated. subInfo contains information about the subscription.
// 6.3 Configure Subscription Options
func NewSubOptsRq(serviceId string, subInfo SubInfo) (*IQ, error) {
	if e := subInfo.validate(); e != nil {
		return nil, e
	}

	iq, err := NewIQ(Attrs{Type: IQTypeGet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		SubOptions: &SubOptions{
			SubInfo: subInfo,
		},
	}
	return iq, nil
}

// NewFormSubmission builds a form submission pubsub IQ
// Information about the subscription and the requester are separated. subInfo contains information about the subscription.
// 6.3.5 Form Submission
func NewFormSubmission(serviceId string, subInfo SubInfo, form *Form) (*IQ, error) {
	if e := subInfo.validate(); e != nil {
		return nil, e
	}
	if form.Type != FormTypeSubmit {
		return nil, errors.New("form type was expected to be submit but was : " + form.Type)
	}

	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		SubOptions: &SubOptions{
			SubInfo: subInfo,
			Form:    form,
		},
	}
	return iq, nil
}

// NewSubAndConfig builds a subscribe request that contains configuration options for the service
// From XEP-0060 : The <options/> element MUST follow the <subscribe/> element and
// MUST NOT possess a 'node' attribute or 'jid' attribute,
// since the value of the <subscribe/> element's 'node' attribute specifies the desired NodeID and
// the value of the <subscribe/> element's 'jid' attribute specifies the subscriber's JID
// 6.3.7 Subscribe and Configure
func NewSubAndConfig(serviceId string, subInfo SubInfo, form *Form) (*IQ, error) {
	if e := subInfo.validate(); e != nil {
		return nil, e
	}
	if form.Type != FormTypeSubmit {
		return nil, errors.New("form type was expected to be submit but was : " + form.Type)
	}
	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		Subscribe: &subInfo,
		SubOptions: &SubOptions{
			SubInfo: SubInfo{SubId: subInfo.SubId},
			Form:    form,
		},
	}
	return iq, nil

}

// NewItemsRequest creates a request to query existing items from a node.
// Specify a "maxItems" value to request only the last maxItems items. If 0, requests all items.
// 6.5.2 Requesting All List AND 6.5.7 Requesting the Most Recent List
func NewItemsRequest(serviceId string, node string, maxItems int) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeGet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		Items: &Items{Node: node},
	}

	if maxItems != 0 {
		ps, _ := iq.Payload.(*PubSubGeneric)
		ps.Items.MaxItems = maxItems
	}
	return iq, nil
}

// NewItemsRequest creates a request to get a specific item from a node.
// 6.5.8 Requesting a Particular Item
func NewSpecificItemRequest(serviceId, node, itemId string) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeGet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		Items: &Items{Node: node,
			List: []Item{
				{
					Id: itemId,
				},
			},
		},
	}
	return iq, nil
}

// NewPublishItemRq creates a request to publish a single item to a node identified by its provided ID
func NewPublishItemRq(serviceId, nodeID, pubItemID string, item Item) (*IQ, error) {
	// "The <publish/> element MUST possess a 'node' attribute, specifying the NodeID of the node."
	if strings.TrimSpace(nodeID) == "" {
		return nil, errors.New("cannot publish without a target node ID")
	}

	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		Publish: &Publish{Node: nodeID, Items: []Item{item}},
	}

	// "The <item/> element provided by the publisher MAY possess an 'id' attribute,
	// specifying a unique ItemID for the item.
	// If an ItemID is not provided in the publish request,
	// the pubsub service MUST generate one and MUST ensure that it is unique for that node."
	if strings.TrimSpace(pubItemID) != "" {
		ps, _ := iq.Payload.(*PubSubGeneric)
		ps.Publish.Items[0].Id = pubItemID
	}
	return iq, nil
}

// NewPublishItemOptsRq creates a request to publish items to a node identified by its provided ID, along with configuration options
// A pubsub service MAY support the ability to specify options along with a publish request
//(if so, it MUST advertise support for the "http://jabber.org/protocol/pubsub#publish-options" feature).
func NewPublishItemOptsRq(serviceId, nodeID string, items []Item, options *PublishOptions) (*IQ, error) {
	// "The <publish/> element MUST possess a 'node' attribute, specifying the NodeID of the node."
	if strings.TrimSpace(nodeID) == "" {
		return nil, errors.New("cannot publish without a target node ID")
	}

	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		Publish:        &Publish{Node: nodeID, Items: items},
		PublishOptions: options,
	}

	return iq, nil
}

// NewDelItemFromNode creates a request to delete and item from a node, given its id.
// To delete an item, the publisher sends a retract request.
// This helper function follows 7.2 Delete an Item from a Node
func NewDelItemFromNode(serviceId, nodeID, itemId string, notify *bool) (*IQ, error) {
	// "The <retract/> element MUST possess a 'node' attribute, specifying the NodeID of the node."
	if strings.TrimSpace(nodeID) == "" {
		return nil, errors.New("cannot delete item without a target node ID")
	}

	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		Retract: &Retract{Node: nodeID, Items: []Item{{Id: itemId}}, Notify: notify},
	}
	return iq, nil
}

// NewCreateAndConfigNode makes a request for node creation that has the desired node configuration.
// See 8.1.3 Create and Configure a Node
func NewCreateAndConfigNode(serviceId, nodeID string, confForm *Form) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		Create:    &Create{Node: nodeID},
		Configure: &Configure{Form: confForm},
	}
	return iq, nil
}

// NewCreateNode builds a request to create a node on the service referenced by "serviceId"
// See 8.1 Create a Node
func NewCreateNode(serviceId, nodeName string) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeSet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		Create: &Create{Node: nodeName},
	}
	return iq, nil
}

// NewRetrieveAllSubsRequest builds a request to retrieve all subscriptions from all nodes
// In order to make the request, the requesting entity MUST send an IQ-get whose <pubsub/>
// child contains an empty <subscriptions/> element with no attributes.
func NewRetrieveAllSubsRequest(serviceId string) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeGet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		Subscriptions: &Subscriptions{},
	}
	return iq, nil
}

// NewRetrieveAllAffilsRequest builds a request to retrieve all affiliations from all nodes
// In order to make the request of the service, the requesting entity includes an empty <affiliations/> element with no attributes.
func NewRetrieveAllAffilsRequest(serviceId string) (*IQ, error) {
	iq, err := NewIQ(Attrs{Type: IQTypeGet, To: serviceId})
	if err != nil {
		return nil, err
	}
	iq.Payload = &PubSubGeneric{
		Affiliations: &Affiliations{},
	}
	return iq, nil
}

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{Space: "http://jabber.org/protocol/pubsub", Local: "pubsub"}, PubSubGeneric{})
}
