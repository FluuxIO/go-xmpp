package main

import (
	"fmt"
	"log"

	"gosrc.io/xmpp"
)

func main() {
	component := MyComponent{Name: "Test Component", Category: "gateway", Type: "service"}
	component.xmpp = &xmpp.Component{Host: "service.localhost", Secret: "mypass"}
	if err := component.xmpp.Connect("localhost:8888"); err != nil {
		log.Fatal(err)
	}

	for {
		packet, err := component.xmpp.ReadPacket()
		if err != nil {
			fmt.Println("read error", err)
			return
		}

		switch p := packet.(type) {
		case xmpp.IQ:
			switch inner := p.Payload[0].(type) {
			case *xmpp.DiscoInfo:
				fmt.Println("Disco Info")
				if p.Type == "get" {
					DiscoResult(component, p.PacketAttrs, inner)
				}
			case *xmpp.DiscoItems:
				fmt.Println("DiscoItems")
				if p.Type == "get" {
					DiscoItems(component, p.PacketAttrs, inner)
				}
			default:
				fmt.Println("ignoring iq packet", inner)
				xError := xmpp.Err{
					Code:   501,
					Reason: "feature-not-implemented",
					Type:   "cancel",
				}
				reply := p.MakeError(xError)
				_ = component.xmpp.Send(&reply)
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

type MyComponent struct {
	Name string
	// Typical categories and types: https://xmpp.org/registrar/disco-categories.html
	Category string
	Type     string

	xmpp *xmpp.Component
}

func DiscoResult(c MyComponent, attrs xmpp.PacketAttrs, info *xmpp.DiscoInfo) {
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

	_ = c.xmpp.Send(iq)
}

func DiscoItems(c MyComponent, attrs xmpp.PacketAttrs, items *xmpp.DiscoItems) {
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
	_ = c.xmpp.Send(iq)
}
