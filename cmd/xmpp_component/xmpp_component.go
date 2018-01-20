package main

import (
	"fmt"

	"fluux.io/xmpp"
)

func main() {
	component := MyComponent{Name: "MQTT Component", Category: "gateway", Type: "mqtt"}
	component.xmpp = &xmpp.Component{Host: "mqtt.localhost", Secret: "mypass"}
	component.xmpp.Connect("localhost:8888")

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
					DiscoResult(component, p.From, p.To, p.Id)
				}

			default:
				fmt.Println("ignoring iq packet", inner)
				xerror := xmpp.Err{
					Code:   501,
					Reason: "feature-not-implemented",
					Type:   "cancel",
				}
				reply := p.MakeError(xerror)
				component.xmpp.Send(&reply)
			}
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

func DiscoResult(c MyComponent, from, to, id string) {
	iq := xmpp.NewIQ("result", to, from, id, "en")
	payload := xmpp.DiscoInfo{
		Identity: xmpp.Identity{
			Name:     c.Name,
			Category: c.Category,
			Type:     c.Type,
		},
		Features: []xmpp.Feature{
			{Var: "http://jabber.org/protocol/disco#info"},
			{Var: "http://jabber.org/protocol/disco#item"},
		},
	}
	iq.AddPayload(&payload)
	c.xmpp.Send(iq)
}
