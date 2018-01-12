package main

import "fluux.io/xmpp"

func main() {
	component := xmpp.Component{Host: "mqtt.localhost", Secret: "mypass"}
	component.Connect("localhost:8888")
}
