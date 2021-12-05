# Fluux XMPP

[![GoDoc](https://godoc.org/gosrc.io/xmpp?status.svg)](https://godoc.org/gosrc.io/xmpp) [![GoReportCard](https://goreportcard.com/badge/gosrc.io/xmpp)](https://goreportcard.com/report/fluux.io/xmpp) [![Coverage Status](https://coveralls.io/repos/github/FluuxIO/go-xmpp/badge.svg?branch=master)](https://coveralls.io/github/FluuxIO/go-xmpp?branch=master)

Fluux XMPP is a Go XMPP library, focusing on simplicity, simple automation, and IoT.

The goal is to make simple to write simple XMPP clients and components:

- For automation (like for example monitoring of an XMPP service),
- For building connected "things" by plugging them on an XMPP server,
- For writing simple chatbot to control a service or a thing,
- For writing XMPP servers components.

The library is designed to have minimal dependencies. Currently it requires at least Go 1.13.

## Configuration and connection

### Allowing Insecure TLS connection during development

It is not recommended to disable the check for domain name and certificate chain. Doing so would open your client
to man-in-the-middle attacks.

However, in development, XMPP servers often use self-signed certificates. In that situation, it is better to add the
root CA that signed the certificate to your trusted list of root CA. It avoids changing the code and limit the risk
of shipping an insecure client to production.

That said, if you really want to allow your client to trust any TLS certificate, you can customize Go standard 
`tls.Config` and set it in Config struct.

Here is an example code to configure a client to allow connecting to a server with self-signed certificate. Note the 
`InsecureSkipVerify` option. When using this `tls.Config` option, all the checks on the certificate are skipped.

```go
config := xmpp.Config{
	Address:      "localhost:5222",
	Jid:          "test@localhost",
	Credential:   xmpp.Password("Test"),
	TLSConfig:    tls.Config{InsecureSkipVerify: true},
}
```

## Supported specifications

### Clients

- [RFC 6120: XMPP Core](https://xmpp.org/rfcs/rfc6120.html)
- [RFC 6121: XMPP Instant Messaging and Presence](https://xmpp.org/rfcs/rfc6121.html)

### Components

  - [XEP-0114: Jabber Component Protocol](https://xmpp.org/extensions/xep-0114.html)
  - [XEP-0355: Namespace Delegation](https://xmpp.org/extensions/xep-0355.html)
  - [XEP-0356: Privileged Entity](https://xmpp.org/extensions/xep-0356.html)

### Extensions 
  - [XEP-0060: Publish-Subscribe](https://xmpp.org/extensions/xep-0060.html)  
    Note : "6.5.4 Returning Some Items" requires support for [XEP-0059: Result Set Management](https://xmpp.org/extensions/xep-0059.html), 
    and is therefore not supported yet. 
  - [XEP-0004: Data Forms](https://xmpp.org/extensions/xep-0004.html)
  - [XEP-0050: Ad-Hoc Commands](https://xmpp.org/extensions/xep-0050.html)

## Package overview

### Stanza subpackage

XMPP stanzas are basic and extensible XML elements. Stanzas (or sometimes special stanzas called 'nonzas') are used to 
leverage the XMPP protocol features. During a session, a client (or a component) and a server will be exchanging stanzas
back and forth.

At a low-level, stanzas are XML fragments. However, Fluux XMPP library provides the building blocks to interact with
stanzas at a high-level, providing a Go-friendly API.

The `stanza` subpackage provides support for XMPP stream parsing, marshalling and unmarshalling of XMPP stanza. It is a
bridge between high-level Go structure and low-level XMPP protocol.

Parsing, marshalling and unmarshalling is automatically handled by Fluux XMPP client library. As a developer, you will
generally manipulates only the high-level structs provided by the stanza package.

The XMPP protocol, as the name implies is extensible. If your application is using custom stanza extensions, you can
implement your own extensions directly in your own application.

To learn more about the stanza package, you can read more in the
[stanza package documentation](https://github.com/FluuxIO/go-xmpp/blob/master/stanza/README.md).

### Router

TODO

### Getting IQ response from server

TODO

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
	"gosrc.io/xmpp/stanza"
)

func main() {
	config := xmpp.Config{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: "localhost:5222",
		},
		Jid:          "test@localhost",
		Credential:   xmpp.Password("test"),
		StreamLogger: os.Stdout,
		Insecure:     true,
		// TLSConfig: tls.Config{InsecureSkipVerify: true},
	}

	router := xmpp.NewRouter()
	router.HandleFunc("message", handleMessage)

	client, err := xmpp.NewClient(config, router, errorHandler)
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
	reply := stanza.Message{Attrs: stanza.Attrs{To: msg.From}, Body: msg.Body}
	_ = s.Send(reply)
}

func errorHandler(err error) {
	fmt.Println(err.Error())
}

```

## Reference documentation

The code documentation is available on GoDoc: [gosrc.io/xmpp](https://godoc.org/gosrc.io/xmpp)
