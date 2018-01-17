package main

import (
	"encoding/xml"
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
			case *xmpp.Node:
				fmt.Printf("%q\n", inner)

				data, err := xml.Marshal(inner)
				if err != nil {
					fmt.Println("cannot marshall payload")
				}
				fmt.Println("data=", string(data))
				component.processIQ(p.Type, p.Id, p.From, inner)
			default:
				fmt.Println("default")
			}
		default:
			fmt.Println("Packet unhandled packet:", packet)
		}
	}
}

const (
	NSDiscoInfo = "http://jabber.org/protocol/disco#info"
)

type MyComponent struct {
	Name string
	// Typical categories and types: https://xmpp.org/registrar/disco-categories.html
	Category string
	Type     string

	xmpp *xmpp.Component
}

func (c MyComponent) processIQ(iqType, id, from string, inner *xmpp.Node) {
	fmt.Println("Node:", inner.XMLName.Space, inner.XMLName.Local)
	switch inner.XMLName.Space + " " + iqType {
	case NSDiscoInfo + " get":
		fmt.Println("Send Disco Info")

		iq := xmpp.NewIQ("result", "admin@localhost", "test@localhost", "1", "en")
		payload := xmpp.DiscoInfo{
			Identity: xmpp.Identity{
				Name:     "Test Gateway",
				Category: "gateway",
				Type:     "mqtt",
			},
			Features: []xmpp.Feature{
				{Var: "http://jabber.org/protocol/disco#info"},
				{Var: "http://jabber.org/protocol/disco#item"},
			},
		}
		iq.AddPayload(&payload)
		c.xmpp.Send(iq)
	default:
		iqErr := fmt.Sprintf(`<iq type='error'
    from='%s'
    to='%s'
    id='%s'>
     <error type="cancel" code="501">
      <feature-not-implemented xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"/>
     </error>
</iq>`, c.xmpp.Host, from, id)
		c.xmpp.SendOld(iqErr) // FIXME Remove that method
	}
}
