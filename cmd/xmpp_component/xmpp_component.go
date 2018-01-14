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
			switch inner := p.Payload.(type) {
			case *xmpp.Node:
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
		result := fmt.Sprintf(`<iq type='result'
    from='%s'
    to='%s'
    id='%s'>
  <query xmlns='http://jabber.org/protocol/disco#info'>
    <identity
        category='%s'
        type='%s'
        name='%s'/>
    <feature var='http://jabber.org/protocol/disco#info'/>
    <feature var='http://jabber.org/protocol/disco#items'/>
  </query>
</iq>`, c.xmpp.Host, from, id, c.Category, c.Type, c.Name)
		c.xmpp.Send(result)
	default:
		iqErr := fmt.Sprintf(`<iq type='error'
    from='%s'
    to='%s'
    id='%s'>
     <error type="cancel" code="501">
      <feature-not-implemented xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"/>
     </error>
</iq>`, c.xmpp.Host, from, id)
		c.xmpp.Send(iqErr)
	}
}
