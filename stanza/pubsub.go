package stanza

import (
	"encoding/xml"
)

type PubSub struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/pubsub pubsub"`
	Publish *Publish
	Retract *Retract
	// TODO <configure/>
}

func (p *PubSub) Namespace() string {
	return p.XMLName.Space
}

type Publish struct {
	XMLName xml.Name `xml:"publish"`
	Node    string   `xml:"node,attr"`
	Item    Item
}

type Item struct {
	XMLName xml.Name `xml:"item"`
	Id      string   `xml:"id,attr,omitempty"`
	Tune    *Tune
	Mood    *Mood
}

type Retract struct {
	XMLName xml.Name `xml:"retract"`
	Node    string   `xml:"node,attr"`
	Notify  string   `xml:"notify,attr"`
	Item    Item
}

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{"http://jabber.org/protocol/pubsub", "pubsub"}, PubSub{})
}
