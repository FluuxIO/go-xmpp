package main

import (
	"encoding/xml"
	"fmt"
	"log"

	"gosrc.io/xmpp"
)

func main() {
	opts := xmpp.ComponentOptions{
		Domain:  "service.localhost",
		Secret:  "mypass",
		Address: "localhost:9999",

		// TODO: Move that part to a component discovery handler
		Name:     "Test Component",
		Category: "gateway",
		Type:     "service",
	}

	router := xmpp.NewRouter()
	router.HandleFunc("message", HandleMessage)
	router.NewRoute().
		IQNamespaces(xmpp.NSDiscoInfo).
		HandlerFunc(func(s xmpp.Sender, p xmpp.Packet) {
			DiscoInfo(s, p, opts)
		})
	router.NewRoute().
		IQNamespaces("urn:xmpp:delegation:1").
		HandlerFunc(HandleDelegation)

	component, err := xmpp.NewComponent(opts, router)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// If you pass the component to a stream manager, it will handle the reconnect policy
	// for you automatically.
	// TODO: Post Connect could be a feature of the router or the client. Move it somewhere else.
	cm := xmpp.NewStreamManager(component, nil)
	log.Fatal(cm.Run())
}

func HandleMessage(_ xmpp.Sender, p xmpp.Packet) {
	msg, ok := p.(xmpp.Message)
	if !ok {
		return
	}
	var msgProcessed bool
	for _, ext := range msg.Extensions {
		delegation, ok := ext.(*xmpp.Delegation)
		if ok {
			msgProcessed = true
			fmt.Printf("Delegation confirmed for namespace %s\n", delegation.Delegated.Namespace)
		}
	}
	// TODO: Decode privilege message
	// <message to='service.localhost' from='localhost'><privilege xmlns='urn:xmpp:privilege:1'><perm type='outgoing' access='message'/><perm type='get' access='roster'/><perm type='managed_entity' access='presence'/></privilege></message>

	if !msgProcessed {
		fmt.Printf("Ignored received message, not related to delegation: %v\n", msg)
	}
}

const (
	pubsubNode = "urn:xmpp:delegation:1::http://jabber.org/protocol/pubsub"
	pepNode    = "urn:xmpp:delegation:1:bare:http://jabber.org/protocol/pubsub"
)

// TODO: replace xmpp.Sender by ctx xmpp.Context ?
// ctx.Stream.Send / SendRaw
// ctx.Opts
func DiscoInfo(c xmpp.Sender, p xmpp.Packet, opts xmpp.ComponentOptions) {
	// Type conversion & sanity checks
	iq, ok := p.(xmpp.IQ)
	if !ok {
		return
	}
	info, ok := iq.Payload[0].(*xmpp.DiscoInfo)
	if !ok {
		return
	}

	iqResp := xmpp.NewIQ("result", iq.To, iq.From, iq.Id, "en")

	switch info.Node {
	case "":
		DiscoInfoRoot(&iqResp, opts)
	case pubsubNode:
		DiscoInfoPubSub(&iqResp)
	case pepNode:
		DiscoInfoPEP(&iqResp)
	}

	_ = c.Send(iqResp)
}

func DiscoInfoRoot(iqResp *xmpp.IQ, opts xmpp.ComponentOptions) {
	// Higher level discovery
	identity := xmpp.Identity{
		Name:     opts.Name,
		Category: opts.Category,
		Type:     opts.Type,
	}
	payload := xmpp.DiscoInfo{
		XMLName: xml.Name{
			Space: xmpp.NSDiscoInfo,
			Local: "query",
		},
		Identity: identity,
		Features: []xmpp.Feature{
			{Var: xmpp.NSDiscoInfo},
			{Var: xmpp.NSDiscoItems},
			{Var: "jabber:iq:version"},
			{Var: "urn:xmpp:delegation:1"},
		},
	}
	iqResp.AddPayload(&payload)
}

func DiscoInfoPubSub(iqResp *xmpp.IQ) {
	payload := xmpp.DiscoInfo{
		XMLName: xml.Name{
			Space: xmpp.NSDiscoInfo,
			Local: "query",
		},
		Node: pubsubNode,
		Features: []xmpp.Feature{
			{Var: "http://jabber.org/protocol/pubsub"},
			{Var: "http://jabber.org/protocol/pubsub#publish"},
			{Var: "http://jabber.org/protocol/pubsub#subscribe"},
			{Var: "http://jabber.org/protocol/pubsub#publish-options"},
		},
	}
	iqResp.AddPayload(&payload)
}

func DiscoInfoPEP(iqResp *xmpp.IQ) {
	identity := xmpp.Identity{
		Category: "pubsub",
		Type:     "pep",
	}
	payload := xmpp.DiscoInfo{
		XMLName: xml.Name{
			Space: xmpp.NSDiscoInfo,
			Local: "query",
		},
		Identity: identity,
		Node:     pepNode,
		Features: []xmpp.Feature{
			{Var: "http://jabber.org/protocol/pubsub#access-presence"},
			{Var: "http://jabber.org/protocol/pubsub#auto-create"},
			{Var: "http://jabber.org/protocol/pubsub#auto-subscribe"},
			{Var: "http://jabber.org/protocol/pubsub#config-node"},
			{Var: "http://jabber.org/protocol/pubsub#create-and-configure"},
			{Var: "http://jabber.org/protocol/pubsub#create-nodes"},
			{Var: "http://jabber.org/protocol/pubsub#filtered-notifications"},
			{Var: "http://jabber.org/protocol/pubsub#persistent-items"},
			{Var: "http://jabber.org/protocol/pubsub#publish"},
			{Var: "http://jabber.org/protocol/pubsub#retrieve-items"},
			{Var: "http://jabber.org/protocol/pubsub#subscribe"},
		},
	}
	iqResp.AddPayload(&payload)
}

func HandleDelegation(s xmpp.Sender, p xmpp.Packet) {
	// Type conversion & sanity checks
	iq, ok := p.(xmpp.IQ)
	if !ok {
		return
	}

	payload1 := iq.Payload[0]
	delegation, ok := payload1.(*xmpp.Delegation)
	if !ok {
		return
	}
	forwardedPacket := delegation.Forwarded.Stanza
	fmt.Println(forwardedPacket)
	forwardedIQ, ok := forwardedPacket.(xmpp.IQ)
	if !ok {
		return
	}
	payload := forwardedIQ.Payload
	if len(payload) == 0 {
		return
	}

	pubsub, ok := payload[0].(*xmpp.PubSub)
	if !ok {
		// We only support pubsub delegation
		return
	}

	if pubsub.Publish.XMLName.Local == "publish" {
		// Prepare pubsub IQ reply
		iqResp := xmpp.NewIQ("result", forwardedIQ.To, forwardedIQ.From, forwardedIQ.Id, "en")
		payload := xmpp.PubSub{
			XMLName: xml.Name{
				Space: "http://jabber.org/protocol/pubsub",
				Local: "pubsub",
			},
		}
		iqResp.AddPayload(&payload)
		// Wrap the reply in delegation 'forward'
		iqForward := xmpp.NewIQ("result", iq.To, iq.From, iq.Id, "en")
		delegPayload := xmpp.Delegation{
			XMLName: xml.Name{
				Space: "urn:xmpp:delegation:1",
				Local: "delegation",
			},
			Forwarded: &xmpp.Forwarded{
				XMLName: xml.Name{
					Space: "urn:xmpp:forward:0",
					Local: "forward",
				},
				Stanza: iqResp,
			},
		}
		iqForward.AddPayload(&delegPayload)
		_ = s.Send(iqForward)
		// TODO: The component should actually broadcast the mood to subscribers
	}
}
