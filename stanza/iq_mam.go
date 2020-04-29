package stanza

import (
	"encoding/xml"
	"github.com/google/uuid"
)

// ----------
// Namespaces

const (
	// NSRoster is the Roster IQ namespace
	NSMam = "urn:xmpp:mam:2"
)

// Roster struct represents Roster IQs
type MamQuery struct {
	XMLName xml.Name `xml:"urn:xmpp:mam:2 query"`
	QueryId string   `xml:"queryid,attr"`
}

// Namespace defines the namespace for the RosterIQ
func (mq *MamQuery) Namespace() string {
	return mq.XMLName.Space
}
func (mq *MamQuery) GetQueryId() string {
	return mq.QueryId
}

// To implement IqPayload interface only
func (mq *MamQuery) GetSet() *ResultSet {
	return nil
}

// ---------------
// Builder helpers

// RosterIQ builds a default Roster payload
func (iq *IQ) NewMamIQ() *MamQuery {
	mq := MamQuery{
		XMLName: xml.Name{
			Space: NSMam,
			Local: "query",
		},
	}
	if id, err := uuid.NewRandom(); err == nil {
		mq.QueryId = id.String()
	}

	iq.Payload = &mq
	return &mq
}
