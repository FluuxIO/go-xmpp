# Fluux XMPP

[![Codeship Status for FluuxIO/xmpp](https://app.codeship.com/projects/dba7f300-d145-0135-6c51-26e28af241d2/status?branch=master)](https://app.codeship.com/projects/262399) [![GoDoc](https://godoc.org/gosrc.io/xmpp?status.svg)](https://godoc.org/gosrc.io/xmpp) [![GoReportCard](https://goreportcard.com/badge/gosrc.io/xmpp)](https://goreportcard.com/report/fluux.io/xmpp) [![codecov](https://codecov.io/gh/FluuxIO/go-xmpp/branch/master/graph/badge.svg)](https://codecov.io/gh/FluuxIO/go-xmpp)

Fluux XMPP is a Go XMPP library, focusing on simplicity, simple automation, and IoT.

The goal is to make simple to write simple XMPP clients and components:

- For automation (like for example monitoring of an XMPP service),
- For building connected "things" by plugging them on an XMPP server,
- For writing simple chatbot to control a service or a thing.
- For writing XMPP servers components. Fluux XMPP supports:
  - [XEP-0114: Jabber Component Protocol](https://xmpp.org/extensions/xep-0114.html)
  - [XEP-0355: Namespace Delegation](https://xmpp.org/extensions/xep-0355.html)
  - [XEP-0356: Privileged Entity](https://xmpp.org/extensions/xep-0356.html)

The library is designed to have minimal dependencies. For now, the library does not depend on any other library.

## Examples

We have several [examples](https://github.com/FluuxIO/go-xmpp/tree/master/_examples) to help you get started using
Fluux XMPP library.

Here is the demo "echo" client:

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
```

## Reference documentation

The code documentation is available on GoDoc: [gosrc.io/xmpp](https://godoc.org/gosrc.io/xmpp)
