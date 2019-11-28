package main

import (
	"fmt"
	"log"

	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

func main() {
	opts := xmpp.ComponentOptions{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: "localhost:8888",
			Domain:  "service2.localhost",
		},
		Domain:   "service2.localhost",
		Secret:   "mypass",
		Name:     "Test Component",
		Category: "gateway",
		Type:     "service",
	}

	router := xmpp.NewRouter()
	router.HandleFunc("message", handleMessage)
	router.NewRoute().
		IQNamespaces(stanza.NSDiscoInfo).
		HandlerFunc(func(s xmpp.Sender, p stanza.Packet) {
			discoInfo(s, p, opts)
		})
	router.NewRoute().
		IQNamespaces(stanza.NSDiscoItems).
		HandlerFunc(discoItems)
	router.NewRoute().
		IQNamespaces("jabber:iq:version").
		HandlerFunc(handleVersion)

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

func handleMessage(_ xmpp.Sender, p stanza.Packet) {
	msg, ok := p.(stanza.Message)
	if !ok {
		return
	}
	fmt.Println("Received message:", msg.Body)
}

func discoInfo(c xmpp.Sender, p stanza.Packet, opts xmpp.ComponentOptions) {
	// Type conversion & sanity checks
	iq, ok := p.(stanza.IQ)
	if !ok || iq.Type != stanza.IQTypeGet {
		return
	}

	iqResp := stanza.NewIQ(stanza.Attrs{Type: "result", From: iq.To, To: iq.From, Id: iq.Id, Lang: "en"})
	disco := iqResp.DiscoInfo()
	disco.AddIdentity(opts.Name, opts.Category, opts.Type)
	disco.AddFeatures(stanza.NSDiscoInfo, stanza.NSDiscoItems, "jabber:iq:version", "urn:xmpp:delegation:1")
	_ = c.Send(iqResp)
}

// TODO: Handle iq error responses
func discoItems(c xmpp.Sender, p stanza.Packet) {
	// Type conversion & sanity checks
	iq, ok := p.(stanza.IQ)
	if !ok || iq.Type != stanza.IQTypeGet {
		return
	}

	discoItems, ok := iq.Payload.(*stanza.DiscoItems)
	if !ok {
		return
	}

	iqResp := stanza.NewIQ(stanza.Attrs{Type: "result", From: iq.To, To: iq.From, Id: iq.Id, Lang: "en"})
	items := iqResp.DiscoItems()

	if discoItems.Node == "" {
		items.AddItem("service.localhost", "node1", "test node")
	}
	_ = c.Send(iqResp)
}

func handleVersion(c xmpp.Sender, p stanza.Packet) {
	// Type conversion & sanity checks
	iq, ok := p.(stanza.IQ)
	if !ok {
		return
	}

	iqResp := stanza.NewIQ(stanza.Attrs{Type: "result", From: iq.To, To: iq.From, Id: iq.Id, Lang: "en"})
	iqResp.Version().SetInfo("Fluux XMPP Component", "0.0.1", "")
	_ = c.Send(iqResp)
}
