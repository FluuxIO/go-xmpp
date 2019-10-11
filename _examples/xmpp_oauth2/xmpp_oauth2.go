/*
xmpp_oauth2 is a demo client that connect on an XMPP server using OAuth2 and prints received messages.
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
			Address: "localhost:5222",
			// TLSConfig: tls.Config{InsecureSkipVerify: true},
		},
		Jid:          "test@localhost",
		Credential:   xmpp.OAuthToken("OdAIsBlY83SLBaqQoClAn7vrZSHxixT8"),
		StreamLogger: os.Stdout,
		// Insecure:     true,
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
