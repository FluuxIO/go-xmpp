/*
xmpp_echo is a demo client that connect on an XMPP server and echo message received back to original sender.
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

func handleMessage(s xmpp.Sender, p xmpp.Packet) {
	msg, ok := p.(xmpp.Message)
	if !ok {
		_, _ = fmt.Fprintf(os.Stdout, "Ignoring packet: %T\n", p)
		return
	}

	_, _ = fmt.Fprintf(os.Stdout, "Body = %s - from = %s\n", msg.Body, msg.From)
	reply := xmpp.Message{Attrs: xmpp.Attrs{To: msg.From}, Body: msg.Body}
	_ = s.Send(reply)
}

// TODO create default command line client to send message or to send an arbitrary XMPP sequence from a file,
//   (using templates ?)
