/*
xmpp_client is a demo client that connect on an XMPP server and echo message received back to original sender.
*/

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mremond/gox/xmpp"
)

func main() {
	options := xmpp.Options{Address: "localhost:5222", Jid: "test@localhost", Password: "test", PacketLogger: os.Stdout}

	var client *xmpp.Client
	var err error
	if client, err = xmpp.NewClient(options); err != nil {
		log.Fatal("Error: ", err)
	}

	var session *xmpp.Session
	if session, err = client.Connect(); err != nil {
		log.Fatal("Error: ", err)
	}

	fmt.Println("Stream opened, we have streamID = ", session.StreamId)

	// Iterator to receive packets coming from our XMPP connection
	for packet := range client.Recv() {
		switch packet := packet.(type) {
		case *xmpp.ClientMessage:
			fmt.Fprintf(os.Stdout, "Body = %s - from = %s\n", packet.Body, packet.From)
			reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: packet.Body}
			client.Send(reply.XMPPFormat())
		default:
			fmt.Fprintf(os.Stdout, "Ignoring packet: %T\n", packet)
		}
	}
}

// TODO create default command line client to send message or to send an arbitrary XMPP sequence from a file,
// (using templates ?)

// TODO: autoreconnect when connection is lost
