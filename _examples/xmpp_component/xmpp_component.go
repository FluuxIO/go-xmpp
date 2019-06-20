package main

import (
	"encoding/xml"
	"fmt"
	"log"

	"gosrc.io/xmpp"
)

func main() {
	opts := xmpp.ComponentOptions{
		Domain:   "service2.localhost",
		Secret:   "mypass",
		Address:  "localhost:8888",
		Name:     "Test Component",
		Category: "gateway",
		Type:     "service",
	}

	router := xmpp.NewRouter()
	router.HandleFunc("message", handleMessage)
	router.NewRoute().
		IQNamespaces(xmpp.NSDiscoInfo).
		HandlerFunc(func(s xmpp.Sender, p xmpp.Packet) {
			discoInfo(s, p, opts)
		})
	router.NewRoute().
		IQNamespaces(xmpp.NSDiscoItems).
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

func handleMessage(_ xmpp.Sender, p xmpp.Packet) {
	msg, ok := p.(xmpp.Message)
	if !ok {
		return
	}
	fmt.Println("Received message:", msg.Body)
}

func discoInfo(c xmpp.Sender, p xmpp.Packet, opts xmpp.ComponentOptions) {
	// Type conversion & sanity checks
	iq, ok := p.(xmpp.IQ)
	if !ok {
		return
	}
	
	if iq.Type != "get" {
		return
	}

	iqResp := xmpp.NewIQ("result", iq.To, iq.From, iq.Id, "en")
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
	iqResp.Payload = &payload
	_ = c.Send(iqResp)
}

// TODO: Handle iq error responses
func discoItems(c xmpp.Sender, p xmpp.Packet) {
	// Type conversion & sanity checks
	iq, ok := p.(xmpp.IQ)
	if !ok {
		return
	}
	
	if iq.Type != "get" {
		return
	}

	discoItems, ok := iq.Payload.(*xmpp.DiscoItems)
	if !ok {
		return
	}

	iqResp := xmpp.NewIQ("result", iq.To, iq.From, iq.Id, "en")

	var payload xmpp.DiscoItems
	if discoItems.Node == "" {
		payload = xmpp.DiscoItems{
			Items: []xmpp.DiscoItem{
				{Name: "test node", JID: "service.localhost", Node: "node1"},
			},
		}
	}
	iqResp.Payload = &payload
	_ = c.Send(iqResp)
}

func handleVersion(c xmpp.Sender, p xmpp.Packet) {
	// Type conversion & sanity checks
	iq, ok := p.(xmpp.IQ)
	if !ok {
		return
	}

	iqResp := xmpp.NewIQ("result", iq.To, iq.From, iq.Id, "en")
	var payload xmpp.Version
	payload.Name = "Fluux XMPP Component"
	payload.Version = "0.0.1"
	iq.Payload = &payload
	_ = c.Send(iqResp)
}
