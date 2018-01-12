package xmpp

import (
	"crypto/sha1"
	"encoding/hex"
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
	Host   string
	Secret string

	// TCP level connection
	conn net.Conn

	// read / write
	socketProxy io.ReadWriter // TODO
	decoder     *xml.Decoder
}

type Handshake struct {
	XMLName xml.Name `xml:"jabber:component:accept handshake"`
}

// Handshake generates an authentication token based on StreamID and shared secret.
func (c *Component) Handshake(streamId string) string {
	// 1. Concatenate the Stream ID received from the server with the shared secret.
	concatStr := streamId + c.Secret

	// 2. Hash the concatenated string according to the SHA1 algorithm, i.e., SHA1( concat (sid, password)).
	h := sha1.New()
	h.Write([]byte(concatStr))
	hash := h.Sum(nil)

	// 3. Ensure that the hash output is in hexadecimal format, not binary or base64.
	// 4. Convert the hash output to all lowercase characters.
	encodedStr := hex.EncodeToString(hash)

	return encodedStr
}

// TODO Helper to prepare connection string
func Open(connStr string) error {
	c := Component{Host: connStr, Secret: "mypass"}

	var conn net.Conn
	var err error
	if conn, err = net.DialTimeout("tcp", "localhost:8888", time.Duration(5)*time.Second); err != nil {
		return err
	}
	c.conn = conn

	// TODO send stream open and check for reply
	// Send stream open tag
	componentHost := connStr // TODO Fix me: Extract componentID + secret
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
	if _, err := fmt.Fprintf(conn, "<handshake>%s</handshake>", c.Handshake(streamId)); err != nil {
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
	case *Handshake:
		fmt.Println("Component connected")
	default:
		return errors.New("unexpected packet, got " + name.Local + " in " + name.Space)
	}

	return nil
}
