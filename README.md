# Fluux XMPP

[![Codeship Status for FluuxIO/xmpp](https://app.codeship.com/projects/dba7f300-d145-0135-6c51-26e28af241d2/status?branch=master)](https://app.codeship.com/projects/262399) [![GoDoc](https://godoc.org/gosrc.io/xmpp?status.svg)](https://godoc.org/gosrc.io/xmpp) [![GoReportCard](https://goreportcard.com/badge/gosrc.io/xmpp)](https://goreportcard.com/report/fluux.io/xmpp) [![codecov](https://codecov.io/gh/FluuxIO/go-xmpp/branch/master/graph/badge.svg)](https://codecov.io/gh/FluuxIO/go-xmpp)

Fluux XMPP is a Go XMPP library, focusing on simplicity, simple automation, and IoT.

The goal is to make simple to write simple adhoc XMPP clients:

- For automation (like for example monitoring of an XMPP service),
- For building connected "things" by plugging them on an XMPP server,
- For writing simple chatbot to control a service or a thing.
- For writing XMPP servers components.

The library is designed to have minimal dependencies. For now, the library does not depend on any other library.

## Example

Here is a demo "echo" client:

```go
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
	router.HandleFunc("message", HandleMessage)

	client, err := xmpp.NewClient(config, router)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// If you pass the client to a connection manager, it will handle the reconnect policy
	// for you automatically.
	cm := xmpp.NewStreamManager(client, nil)
	log.Fatal(cm.Run())
}

func HandleMessage(s xmpp.Sender, p xmpp.Packet) {
	msg, ok := p.(xmpp.Message)
	if !ok {
		_, _ = fmt.Fprintf(os.Stdout, "Ignoring packet: %T\n", p)
		return
	}

	_, _ = fmt.Fprintf(os.Stdout, "Body = %s - from = %s\n", msg.Body, msg.From)
	reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: msg.From}, Body: msg.Body}
	_ = s.Send(reply)
}
```

## Documentation

Please, check GoDoc for more information: [gosrc.io/xmpp](https://godoc.org/gosrc.io/xmpp)
