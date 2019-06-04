package xmpp // import "gosrc.io/xmpp/pep"

import (
	"encoding/xml"
)

type PubSub struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/pubsub pubsub"`
	Publish Publish
}

type Publish struct {
	XMLName xml.Name `xml:"publish"`
	Node    string   `xml:"node,attr"`
	Item    Item
}

type Item struct {
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
