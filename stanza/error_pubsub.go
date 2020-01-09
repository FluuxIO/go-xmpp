package stanza

import "encoding/xml"

// PubSubGeneric errors are nested in the pubsub payload of pubsub IQs. There can be many of them in a single
// payload.

type NotAuthorized struct {
	XMLName xml.Name `xml:"not-authorized"`
}

type ClosedNode struct {
	XMLName xml.Name `xml:"closed-node"`
}
type ConfigurationRequired struct {
	XMLName xml.Name `xml:"configuration-required"`
}

type InvalidJid struct {
	XMLName xml.Name `xml:"invalid-jid"`
}
type InvalidOptions struct {
	XMLName xml.Name `xml:"invalid-options"`
}
type InvalidPayload struct {
	XMLName xml.Name `xml:"invalid-payload"`
}
type InvalidSubid struct {
	XMLName xml.Name `xml:"invalid-subid"`
}
type ItemForbidden struct {
	XMLName xml.Name `xml:"item-forbidden"`
}
type ItemRequired struct {
	XMLName xml.Name `xml:"item-required"`
}
type JidRequired struct {
	XMLName xml.Name `xml:"jid-required"`
}
type MaxItemsExceeded struct {
	XMLName xml.Name `xml:"max-items-exceeded"`
}
type MaxNodesExceeded struct {
	XMLName xml.Name `xml:"max-nodes-exceeded"`
}
type NodeIdRequired struct {
	XMLName xml.Name `xml:"nodeid-required"`
}

type NotInRosterGroup struct {
	XMLName xml.Name `xml:"not-in-roster-group"`
}
type NotSubscribed struct {
	XMLName xml.Name `xml:"not-subscribed"`
}
type PayloadTooBig struct {
	XMLName xml.Name `xml:"payload-too-big"`
}
type PayloadRequired struct {
	XMLName xml.Name `xml:"payload-required"`
}
type PendingSubscription struct {
	XMLName xml.Name `xml:"pending-subscription"`
}
type PreconditionNotMet struct {
	XMLName xml.Name `xml:"precondition-not-met"`
}
type PresenceSubscriptionRequired struct {
	XMLName xml.Name `xml:"presence-subscription-required"`
}
type SubidRequired struct {
	XMLName xml.Name `xml:"subid-required"`
}
type TooManySubscriptions struct {
	XMLName xml.Name `xml:"too-many-subscriptions"`
}

// TODO: it's a complex type with sub elements
type Unsupported struct {
	XMLName xml.Name `xml:"unsupported"`
}
