package stanza

import (
	"encoding/xml"
)

// Implementation of the http://jabber.org/protocol/pubsub#event namespace
type PubSubEvent struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/pubsub#event event"`
	MsgExtension
	EventElement EventElement
	//List ItemsEvent
}

func init() {
	TypeRegistry.MapExtension(PKTMessage, xml.Name{Space: "http://jabber.org/protocol/pubsub#event", Local: "event"}, PubSubEvent{})
}

type EventElement interface {
	Name() string
}

// *********************
// Collection
// *********************

const PubSubCollectionEventName = "Collection"

type CollectionEvent struct {
	AssocDisassoc AssocDisassoc
	Node          string `xml:"node,attr,omitempty"`
}

func (c CollectionEvent) Name() string {
	return PubSubCollectionEventName
}

// *********************
// Associate/Disassociate
// *********************
type AssocDisassoc interface {
	GetAssocDisassoc() string
}

// *********************
// Associate
// *********************
const Assoc = "Associate"

type AssociateEvent struct {
	XMLName xml.Name `xml:"associate"`
	Node    string   `xml:"node,attr"`
}

func (a *AssociateEvent) GetAssocDisassoc() string {
	return Assoc
}

// *********************
// Disassociate
// *********************
const Disassoc = "Disassociate"

type DisassociateEvent struct {
	XMLName xml.Name `xml:"disassociate"`
	Node    string   `xml:"node,attr"`
}

func (e *DisassociateEvent) GetAssocDisassoc() string {
	return Disassoc
}

// *********************
// Configuration
// *********************

const PubSubConfigEventName = "Configuration"

type ConfigurationEvent struct {
	Node string `xml:"node,attr,omitempty"`
	Form *Form
}

func (c ConfigurationEvent) Name() string {
	return PubSubConfigEventName
}

// *********************
// Delete
// *********************
const PubSubDeleteEventName = "Delete"

type DeleteEvent struct {
	Node     string         `xml:"node,attr"`
	Redirect *RedirectEvent `xml:"redirect"`
}

func (c DeleteEvent) Name() string {
	return PubSubConfigEventName
}

// *********************
// Redirect
// *********************
type RedirectEvent struct {
	URI string `xml:"uri,attr"`
}

// *********************
// List
// *********************

const PubSubItemsEventName = "List"

type ItemsEvent struct {
	XMLName xml.Name      `xml:"items"`
	Items   []ItemEvent   `xml:"item,omitempty"`
	Node    string        `xml:"node,attr"`
	Retract *RetractEvent `xml:"retract"`
}

type ItemEvent struct {
	XMLName   xml.Name `xml:"item"`
	Id        string   `xml:"id,attr,omitempty"`
	Publisher string   `xml:"publisher,attr,omitempty"`
	Any       *Node    `xml:",any"`
}

func (i ItemsEvent) Name() string {
	return PubSubItemsEventName
}

// *********************
// List
// *********************

type RetractEvent struct {
	XMLName xml.Name `xml:"retract"`
	ID      string   `xml:"node,attr"`
}

// *********************
// Purge
// *********************
const PubSubPurgeEventName = "Purge"

type PurgeEvent struct {
	XMLName xml.Name `xml:"purge"`
	Node    string   `xml:"node,attr"`
}

func (p PurgeEvent) Name() string {
	return PubSubPurgeEventName
}

// *********************
// Subscription
// *********************
const PubSubSubscriptionEventName = "Subscription"

type SubscriptionEvent struct {
	SubStatus string `xml:"subscription,attr,omitempty"`
	Expiry    string `xml:"expiry,attr,omitempty"`
	SubInfo   `xml:",omitempty"`
}

func (s SubscriptionEvent) Name() string {
	return PubSubSubscriptionEventName
}

func (pse *PubSubEvent) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	pse.XMLName = start.Name
	// decode inner elements
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}
		var ee EventElement
		switch tt := t.(type) {
		case xml.StartElement:
			switch tt.Name.Local {
			case "collection":
				ee = &CollectionEvent{}
			case "configuration":
				ee = &ConfigurationEvent{}
			case "delete":
				ee = &DeleteEvent{}
			case "items":
				ee = &ItemsEvent{}
			case "purge":
				ee = &PurgeEvent{}
			case "subscription":
				ee = &SubscriptionEvent{}
			default:
				ee = nil
			}
			// known child element found, decode it
			if ee != nil {
				err = d.DecodeElement(ee, &tt)
				if err != nil {
					return err
				}
				pse.EventElement = ee
			}
		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}

	}
}
