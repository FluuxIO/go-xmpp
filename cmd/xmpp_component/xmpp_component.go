package main

import (
	"fmt"

	"fluux.io/xmpp"
)

func main() {
	component := xmpp.Component{Host: "mqtt.localhost", Secret: "mypass"}
	component.Connect("localhost:8888")

	for {
		_, packet, err := component.ReadPacket()
		if err != nil {
			return
		}
		fmt.Println("Packet received: ", packet)
	}
}
