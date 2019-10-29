/*
xmpp_websocket is a demo client that connect on an XMPP server using websocket and prints received messages.ß
*/

package main

import (
	"fmt"
	"log"
	"os"

	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

func main() {
	config := xmpp.Config{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: "wss://localhost:5443/ws",
		},
		Jid:          "test@localhost",
		Credential:   xmpp.Password("test"),
		StreamLogger: os.Stdout,
	}

	router := xmpp.NewRouter()
	router.HandleFunc("message", handleMessage)

	client, err := xmpp.NewClient(config, router)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// If you pass the client to a connection manager, it will handle the reconnect policy
	// for you automatically.
	cm := xmpp.NewStreamManager(client, nil)
	log.Fatal(cm.Run())
}

func handleMessage(s xmpp.Sender, p stanza.Packet) {
	msg, ok := p.(stanza.Message)
	if !ok {
		_, _ = fmt.Fprintf(os.Stdout, "Ignoring packet: %T\n", p)
		return
	}

	_, _ = fmt.Fprintf(os.Stdout, "Body = %s - from = %s\n", msg.Body, msg.From)
}
