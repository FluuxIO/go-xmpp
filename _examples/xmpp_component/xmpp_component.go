package main

import (
	"fmt"
	"log"

	"gosrc.io/xmpp"
)

func main() {
	opts := xmpp.ComponentOptions{
		Domain:   "service.localhost",
		Secret:   "mypass",
		Address:  "localhost:8888",
		Name:     "Test Component",
		Category: "gateway",
		Type:     "service",
	}
	component, err := xmpp.NewComponent(opts)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// If you pass the component to a connection manager, it will handle the reconnect policy
	// for you automatically.
	cm := xmpp.NewStreamManager(component, nil)
	err = cm.Start()
	if err != nil {
		log.Fatal(err)
	}

	// Iterator to receive packets coming from our XMPP connection
	for packet := range component.Recv() {
		switch p := packet.(type) {
		case xmpp.IQ:
			switch inner := p.Payload[0].(type) {
			case *xmpp.DiscoInfo:
				fmt.Println("DiscoInfo")
				if p.Type == "get" {
					discoResult(component, p.PacketAttrs, inner)
				}
			case *xmpp.DiscoItems:
				fmt.Println("DiscoItems")
				if p.Type == "get" {
					discoItems(component, p.PacketAttrs, inner)
				}
			default:
				fmt.Println("ignoring iq packet", inner)
				xError := xmpp.Err{
					Code:   501,
					Reason: "feature-not-implemented",
					Type:   "cancel",
				}
				reply := p.MakeError(xError)
				_ = component.Send(&reply)
			}

		case xmpp.Message:
			fmt.Println("Received message:", p.Body)

		case xmpp.Presence:
			fmt.Println("Received presence:", p.Type)

		default:
			fmt.Println("ignoring packet:", packet)
		}
	}
}

func discoResult(c *xmpp.Component, attrs xmpp.PacketAttrs, info *xmpp.DiscoInfo) {
	iq := xmpp.NewIQ("result", attrs.To, attrs.From, attrs.Id, "en")
	var identity xmpp.Identity
	if info.Node == "" {
		identity = xmpp.Identity{
			Name:     c.Name,
			Category: c.Category,
			Type:     c.Type,
		}
	}

	payload := xmpp.DiscoInfo{
		Identity: identity,
		Features: []xmpp.Feature{
			{Var: xmpp.NSDiscoInfo},
			{Var: xmpp.NSDiscoItems},
		},
	}
	iq.AddPayload(&payload)

	_ = c.Send(iq)
}

func discoItems(c *xmpp.Component, attrs xmpp.PacketAttrs, items *xmpp.DiscoItems) {
	iq := xmpp.NewIQ("result", attrs.To, attrs.From, attrs.Id, "en")

	var payload xmpp.DiscoItems
	if items.Node == "" {
		payload = xmpp.DiscoItems{
			Items: []xmpp.DiscoItem{
				{Name: "test node", JID: "service.localhost", Node: "node1"},
			},
		}
	}
	iq.AddPayload(&payload)
	_ = c.Send(iq)
}
