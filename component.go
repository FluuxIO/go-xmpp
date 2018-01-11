package xmpp

import (
	"fmt"
	"net"
	"time"
)

const componentStreamOpen = "<?xml version='1.0'?><stream:stream to='%s' xmlns='%s' xmlns:stream='%s'>"

// Component implements an XMPP extension allowing to extend XMPP server
// using external components. Component specifications are defined
// in XEP-0114, XEP-0355 and XEP-0356.
type Component struct {
	// TCP level connection
	conn net.Conn
}

// TODO Helper to prepare connection string
func Open(connStr string) error {
	var conn net.Conn
	var err error

	if conn, err = net.DialTimeout("tcp", "localhost:8888", time.Duration(5)*time.Second); err != nil {
		return err
	}

	// TODO send stream open and check for reply
	// Send stream open tag
	componentHost := "mqtt.localhost"
	if _, err := fmt.Fprintf(conn, componentStreamOpen, componentHost, NSComponent, NSStream); err != nil {
		fmt.Println("Cannot send stream open.")
		return err
	}

	return nil
}
