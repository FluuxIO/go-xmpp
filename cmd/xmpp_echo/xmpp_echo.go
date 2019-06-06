/*
xmpp_client is a demo client that connect on an XMPP server and echo message received back to original sender.
*/

package main

import (
	"fmt"
	"log"
	"os"

	"gosrc.io/xmpp"
)

func main() {
	config := xmpp.Config{
		Address:      "localhost:5222",
		Jid:          "test@localhost",
		Password:     "test",
		PacketLogger: os.Stdout,
		Insecure:     true,
	}

	client, err := xmpp.NewClient(config)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	cm := xmpp.NewClientManager(client, nil)
	cm.Start()
	// connection can be stopped with cm.Stop()
	// connection state can be checked by reading cm.Client.CurrentState

	// Iterator to receive packets coming from our XMPP connection
	for packet := range client.Recv() {
		switch packet := packet.(type) {
		case xmpp.Message:
			_, _ = fmt.Fprintf(os.Stdout, "Body = %s - from = %s\n", packet.Body, packet.From)
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From}, Body: packet.Body}
			_ = client.Send(reply)
		default:
			_, _ = fmt.Fprintf(os.Stdout, "Ignoring packet: %T\n", packet)
		}
	}
}

// TODO create default command line client to send message or to send an arbitrary XMPP sequence from a file,
//   (using templates ?)
