package pep

import (
	"encoding/xml"

	"github.com/processone/gox/xmpp"
)

type iq struct {
	XMLName     xml.Name `xml:"jabber:client iq"`
	C           pubSub   // c for "contains"
	xmpp.Packet          // Rename h for "header" ?
}

type pubSub struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/pubsub pubsub"`
	Publish publish
}

type publish struct {
	XMLName xml.Name `xml:"publish"`
	Node    string   `xml:"node,attr"`
	Item    item
}

type item struct {
	XMLName xml.Name `xml:"item"`
	Tune    Tune
}

type Tune struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/tune tune"`
	Artist  string   `xml:"artist,omitempty"`
	Length  int      `xml:"length,omitempty"`
	Rating  int      `xml:"rating,omitempty"`
	Source  string   `xml:"source,omitempty"`
	Title   string   `xml:"title,omitempty"`
	Track   string   `xml:"track,omitempty"`
	Uri     string   `xml:"uri,omitempty"`
}

/*
type PubsubPublish struct {
	XMLName xml.Name `xml:"publish"`
	node    string   `xml:"node,attr"`
	item    PubSubItem
}

type PubSubItem struct {
	xmlName xml.Name `xml:"item"`
}

type Thing2 struct {
	XMLName xml.Name `xml:"publish"`
	node    string   `xml:"node,attr"`
	tune    string   `xml:"http://jabber.org/protocol/tune item>tune"`
}

type Tune struct {
	artist string
	length int
	rating int
	source string
	title  string
	track  string
	uri    string
}
*/

func (t *Tune) XMPPFormat() (s string) {
	packet, _ := xml.Marshal(iq{Packet: xmpp.Packet{Id: "tunes", Type: "set"}, C: pubSub{Publish: publish{Node: "http://jabber.org/protocol/tune", Item: item{Tune: *t}}}})
	return string(packet)
}

/*
func (*Tune) XMPPFormat() string {
	return fmt.Sprintf(
		`<iq type='set' id='%s'>
 <pubsub xmlns='http://jabber.org/protocol/pubsub'>
  <publish node='http://jabber.org/protocol/tune'>
   <item>
    <tune xmlns='http://jabber.org/protocol/tune'>
     <artist>%s</artist>
     <length>%i</length>
     <rating>%i</rating>
     <source>%s</source>
     <title>%s</title>
     <track>%s</track>
     <uri>%s</uri>
    </tune>
   </item>
  </publish>
 </pubsub>
</iq>`)
}
*/
