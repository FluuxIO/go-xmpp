package xmpp // import "gosrc.io/xmpp"

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

//=============================================================================

// ConnState represents the current connection state.
type ConnState = uint8

// This is a the list of events happening on the connection that the
// client can be notified about.
const (
	StateDisconnected ConnState = iota
	StateConnected
	StateSessionEstablished
)

// Event is a structure use to convey event changes related to client state. This
// is for example used to notify the client when the client get disconnected.
type Event struct {
	State       ConnState
	Description string
}

// EventHandler is use to pass events about state of the connection to
// client implementation.
type EventHandler func(Event)

// Client
// ============================================================================

// Client is the main structure used to connect as a client on an XMPP
// server.
type Client struct {
	// Store user defined options
	config Config
	// Session gather data that can be accessed by users of this library
	Session *Session
	// TCP level connection / can be replaced by a TLS session after starttls
	conn net.Conn

	// TODO: Move to ClientManager
	// store low level metrics
	Metrics *Metrics
}

/*
Setting up the client / Checking the parameters
*/

// NewClient generates a new XMPP client, based on Config passed as parameters.
// If host is not specified, the DNS SRV should be used to find the host from the domainpart of the JID.
// Default the port to 5222.
// TODO: better config checks
func NewClient(config Config) (c *Client, err error) {
	// TODO: If option address is nil, use the Jid domain to compose the address
	if config.Address, err = checkAddress(config.Address); err != nil {
		return
	}

	if config.Password == "" {
		err = errors.New("missing password")
		return
	}

	c = new(Client)
	c.config = config

	// Parse JID
	if c.config.parsedJid, err = NewJid(c.config.Jid); err != nil {
		return
	}

	if c.config.ConnectTimeout == 0 {
		c.config.ConnectTimeout = 15 // 15 second as default
	}
	return
}

// TODO Pass JID to be able to add default address based on JID, if addr is empty
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
	var err error

	// TODO: Refactor = abstract retry loop in capped exponential back-off function
	c.Metrics = initMetrics()
	c.conn, err = net.DialTimeout("tcp", c.config.Address, time.Duration(c.config.ConnectTimeout)*time.Second)
	if err != nil {
		return nil, err
	}
	if c.config.Handler != nil {
		c.config.Handler(Event{State: StateConnected})
	}

	// Connection is ok, we now open XMPP session
	if c.conn, c.Session, err = NewSession(c.conn, c.config); err != nil {
		return c.Session, err
	}

	c.Metrics.setLoginTime()
	// We're connected and can now receive and send messages.
	//fmt.Fprintf(client.conn, "<presence xml:lang='en'><show>%s</show><status>%s</status></presence>", "chat", "Online")
	// TODO: Do we always want to send initial presence automatically ?
	// Do we need an option to avoid that or do we rely on client to send the presence itself ?
	fmt.Fprintf(c.Session.socketProxy, "<presence/>")

	return c.Session, err
}

func (c *Client) Disconnect() {
	_ = c.SendRaw("</stream:stream>")
	_ = c.conn.Close()
}

func (c *Client) recv(receiver chan<- interface{}) (err error) {
	for {
		val, err := next(c.Session.decoder)
		if err != nil {
			if c.config.Handler != nil {
				c.config.Handler(Event{State: StateDisconnected})
			}
			close(receiver)
			return err
		}
		receiver <- val
		val = nil
	}
}

// Recv abstracts receiving preparsed XMPP packets from a channel.
// Channel allow client to receive / dispatch packets in for range loop.
// FIXME: The code will not work fine if the XMPP client calls Recv several times.
func (c *Client) Recv() <-chan interface{} {
	ch := make(chan interface{})
	go c.recv(ch)
	return ch
}

// Send marshalls XMPP stanza and sends it to the server.
func (c *Client) Send(packet Packet) error {
	data, err := xml.Marshal(packet)
	if err != nil {
		return errors.New("cannot marshal packet " + err.Error())
	}

	if _, err := fmt.Fprintf(c.conn, string(data)); err != nil {
		return errors.New("cannot send packet " + err.Error())
	}
	return nil
}

// SendRaw sends an XMPP stanza as a string to the server.
// It can be invalid XML or XMPP content. In that case, the server will
// disconnect the client. It is up to the user of this method to
// carefully craft the XML content to produce valid XMPP.
func (c *Client) SendRaw(packet string) error {
	fmt.Fprintf(c.Session.socketProxy, packet) // TODO handle errors
	return nil
}

func xmlEscape(s string) string {
	var b bytes.Buffer
	xml.Escape(&b, []byte(s))
	return b.String()
}
