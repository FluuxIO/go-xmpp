package main // import "gosrc.io/xmpp"

import (
	"log"
	"os"

	"gosrc.io/xmpp"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		log.Fatal("usage: xmpp-check host[:port] [domain]")
	}

	var address string
	var domain string
	if len(args) >= 1 {
		address = args[0]
	}

	if len(args) >= 2 {
		domain = args[1]
	}

	runCheck(address, domain)
}

func runCheck(address, domain string) {
	client, err := xmpp.NewChecker(address, domain)

	if err != nil {
		log.Fatal("Error: ", err)
	}

	if err = client.Check(); err != nil {
		log.Fatal("Failed connection check: ", err)
	}

	log.Println("All checks passed")
}
