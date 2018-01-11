package xmpp

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
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

	// read / write
	socketProxy io.ReadWriter
	decoder     *xml.Decoder
}

// TODO Helper to prepare connection string
func Open(connStr string) error {
	c := Component{}

	var conn net.Conn
	var err error
	if conn, err = net.DialTimeout("tcp", "localhost:8888", time.Duration(5)*time.Second); err != nil {
		return err
	}
	c.conn = conn

	// TODO send stream open and check for reply
	// Send stream open tag
	componentHost := connStr // TODO Fix me
	if _, err := fmt.Fprintf(conn, componentStreamOpen, componentHost, NSComponent, NSStream); err != nil {
		fmt.Println("cannot send stream open.")
		return err
	}
	c.decoder = xml.NewDecoder(conn)

	// Initialize xml decoder and extract streamID from reply
	streamId, err := initDecoder(c.decoder)
	if err != nil {
		fmt.Println("cannot init decoder")
		return err
	}

	fmt.Println("StreamID = ", streamId)

	// Authentication
	if _, err := fmt.Fprint(conn, "<handshake>aaee83c26aeeafcbabeabfcbcd50df997e0a2a1e</handshake>"); err != nil {
		fmt.Println("cannot send stream open.")
		return err
	}

	// Next message should be either success or failure.
	name, val, err := next(c.decoder)
	if err != nil {
		fmt.Println(err)
		return err
	}

	switch v := val.(type) {
	case *StreamError:
		fmt.Printf("error: %s", v.Error.Local)
	default:
		return errors.New("unexpected packet, got " + name.Local + " in " + name.Space)
	}

	return nil
}
