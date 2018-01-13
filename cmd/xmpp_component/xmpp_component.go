package main

import (
	"fmt"

	"fluux.io/xmpp"
)

func main() {
	component := xmpp.Component{Host: "mqtt.localhost", Secret: "mypass"}
	component.Connect("localhost:8888")

	for {
		packet, err := component.ReadPacket()
		if err != nil {
			fmt.Println("read error", err)
			return
		}

		switch p := packet.(type) {
		case xmpp.IQ:
			switch inner := p.Payload.(type) {
			case *xmpp.Node:
				fmt.Println("Node:", inner.XMLName.Space, inner.XMLName.Local)
			default:
				fmt.Println("default")
			}
		default:
			fmt.Println("Packet unhandled packet:", packet)
		}
	}
}
