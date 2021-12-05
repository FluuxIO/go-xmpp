package main

/*

Connect to an XMPP server using XEP 114 protocol, perform a discovery query on the server and print the response

*/

import (
	"context"
	"fmt"
	"log"
	"time"

	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

const (
	domain  = "mycomponent.localhost"
	address = "build.vpn.p1:8888"
)

// Init and return a component
func makeComponent() *xmpp.Component {
	opts := xmpp.ComponentOptions{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: address,
			Domain:  domain,
		},
		Domain: domain,
		Secret: "secret",
	}
	router := xmpp.NewRouter()
	c, err := xmpp.NewComponent(opts, router, handleError)
	if err != nil {
		panic(err)
	}
	return c
}

func handleError(err error) {
	fmt.Println(err.Error())
}

func main() {
	c := makeComponent()

	// Connect Component to the server
	fmt.Printf("Connecting to %v\n", address)
	err := c.Connect()
	if err != nil {
		panic(err)
	}

	// make a disco iq
	iqReq, err := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet,
		From: domain,
		To:   "localhost",
		Id:   "my-iq1"})
	if err != nil {
		log.Fatalf("failed to create IQ: %v", err)
	}
	disco := iqReq.DiscoInfo()
	iqReq.Payload = disco

	// res is the channel used to receive the result iq
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	res, _ := c.SendIQ(ctx, iqReq)

	select {
	case iqResponse := <-res:
		// Got response from server
		fmt.Print(iqResponse.Payload)
	case <-time.After(100 * time.Millisecond):
		cancel()
		panic("No iq response was received in time")
	}
}
