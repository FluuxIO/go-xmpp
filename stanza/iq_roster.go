package stanza

import (
	"encoding/xml"
)

// ============================================================================
// Roster

const (
	// NSRoster is the Roster IQ namespace
	NSRoster = "jabber:iq:roster"
	// SubscriptionNone indicates the user does not have a subscription to
	// the contact's presence, and the contact does not have a subscription
	// to the user's presence; this is the default value, so if the subscription
	// attribute is not included then the state is to be understood as "none"
	SubscriptionNone = "none"

	// SubscriptionTo indicates the user has a subscription to the contact's
	// presence, but the contact does not have a subscription to the user's presence.
	SubscriptionTo = "to"

	// SubscriptionFrom indicates the contact has a subscription to the user's
	// presence, but the user does not have a subscription to the contact's presence
	SubscriptionFrom = "from"

	// SubscriptionBoth indicates the user and the contact have subscriptions to each
	// other's presence (also called a "mutual subscription")
	SubscriptionBoth = "both"
)

// ----------
// Namespaces

// Roster struct represents Roster IQs
type Roster struct {
	XMLName xml.Name `xml:"jabber:iq:roster query"`
	// Result sets
	ResultSet *ResultSet `xml:"set,omitempty"`
}

// Namespace defines the namespace for the RosterIQ
func (r *Roster) Namespace() string {
	return r.XMLName.Space
}
func (r *Roster) GetSet() *ResultSet {
	return r.ResultSet
}

// ---------------
// Builder helpers

// RosterIQ builds a default Roster payload
func (iq *IQ) RosterIQ() *Roster {
	r := Roster{
		XMLName: xml.Name{
			Space: NSRoster,
			Local: "query",
		},
	}
	iq.Payload = &r
	return &r
}

// -----------
// SubElements

// RosterItems represents the list of items in a roster IQ
type RosterItems struct {
	XMLName xml.Name     `xml:"jabber:iq:roster query"`
	Items   []RosterItem `xml:"item"`
	// Result sets
	ResultSet *ResultSet `xml:"set,omitempty"`
}

// Namespace lets RosterItems implement the IQPayload interface
func (r *RosterItems) Namespace() string {
	return r.XMLName.Space
}

func (r *RosterItems) GetSet() *ResultSet {
	return r.ResultSet
}

// RosterItem represents an item in the roster iq
type RosterItem struct {
	XMLName      xml.Name `xml:"jabber:iq:roster item"`
	Jid          string   `xml:"jid,attr"`
	Ask          string   `xml:"ask,attr,omitempty"`
	Name         string   `xml:"name,attr,omitempty"`
	Subscription string   `xml:"subscription,attr,omitempty"`
	Groups       []string `xml:"group"`
}

// ---------------
// Builder helpers

// RosterItems builds a default RosterItems payload
func (iq *IQ) RosterItems() *RosterItems {
	ri := RosterItems{
		XMLName: xml.Name{Space: "jabber:iq:roster", Local: "query"},
	}
	iq.Payload = &ri
	return &ri
}

// AddItem builds an item and ads it to the roster IQ
func (r *RosterItems) AddItem(jid, subscription, ask, name string, groups []string) *RosterItems {
	item := RosterItem{
		Jid:          jid,
		Name:         name,
		Groups:       groups,
		Subscription: subscription,
		Ask:          ask,
	}
	r.Items = append(r.Items, item)
	return r
}

// ============================================================================
// Registry init

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{Space: NSRoster, Local: "query"}, Roster{})
	TypeRegistry.MapExtension(PKTIQ, xml.Name{Space: NSRoster, Local: "query"}, RosterItems{})
}
