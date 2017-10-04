package xmpp

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// Client is the main structure use to connect as a client on an XMPP
// server.
type Client struct {
	// Store user defined options
	options Options
	// Session gather data that can be accessed by users of this library
	Session *Session
	// TCP level connection / can be replace by a TLS session after starttls
	conn net.Conn
}

/*
Setting up the client / Checking the parameters
*/

// NewClient generates a new XMPP client, based on Options passed as parameters.
// If host is not specified, the  DNS SRV should be used to find the host from the domainpart of the JID.
// Default the port to 5222.
// TODO: better options checks
func NewClient(options Options) (c *Client, err error) {
	// TODO: If option address is nil, use the Jid domain to compose the address
	if options.Address, err = checkAddress(options.Address); err != nil {
		return
	}

	if options.Password == "" {
		err = errors.New("missing password")
		return
	}

	c = new(Client)
	c.options = options

	// Parse JID
	if c.options.parsedJid, err = NewJid(c.options.Jid); err != nil {
		return
	}

	if c.options.ConnectTimeout == 0 {
		c.options.ConnectTimeout = 15 // 15 second as default
	}
	return
}

// TODO Pass JID to be able to add default address based on JID, if
// addr is empty
func checkAddress(addr string) (string, error) {
	var err error
	hostport := strings.Split(addr, ":")
	if len(hostport) > 2 {
		err = errors.New("too many colons in xmpp server address")
		return addr, err
	}

	// Address is composed of two parts, we are good
	if len(hostport) == 2 && hostport[1] != "" {
		return addr, err
	}

	// Port was not passed, we append XMPP default port:
	return strings.Join([]string{hostport[0], "5222"}, ":"), err
}

// Connect triggers actual TCP connection, based on previously defined parameters.
func (c *Client) Connect() (*Session, error) {
	var tcpconn net.Conn
	var err error

	// TODO: Refactor = abstract retry loop in capped exponential back-off function
	var try = 0
	var success bool
	for try <= c.options.Retry || !success {
		if tcpconn, err = net.DialTimeout("tcp", c.options.Address, time.Duration(c.options.ConnectTimeout)*time.Second); err == nil {
			success = true
		}
		try++
	}
	if err != nil {
		return nil, err
	}

	// Connection is ok, we now open XMPP session
	c.conn = tcpconn
	if c.conn, c.Session, err = NewSession(c.conn, c.options); err != nil {
		return c.Session, err
	}

	// We're connected and can now receive and send messages.
	//fmt.Fprintf(client.conn, "<presence xml:lang='en'><show>%s</show><status>%s</status></presence>", "chat", "Online")
	// TODO: Do we always want to send initial presence automatically ?
	// Do we need an option to avoid that or do we rely on client to send the presence itself ?
	fmt.Fprintf(c.Session.socketProxy, "<presence/>")

	return c.Session, err
}

func (c *Client) recv(receiver chan<- interface{}) (err error) {
	for {
		_, val, err := next(c.Session.decoder)
		if err != nil {
			return err
		}
		receiver <- val
		val = nil
	}
	panic("unreachable")
}

// Recv abstracts receiving preparsed XMPP packets from a channel.
// Channel allow client to receive / dispatch packets in for range loop.
func (c *Client) Recv() <-chan interface{} {
	ch := make(chan interface{})
	go c.recv(ch)
	return ch
}

// Send sends message text.
func (c *Client) Send(packet string) error {
	fmt.Fprintf(c.Session.socketProxy, packet)
	return nil
}

func xmlEscape(s string) string {
	var b bytes.Buffer
	xml.Escape(&b, []byte(s))
	return b.String()
}
