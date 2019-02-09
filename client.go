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

// Client Metrics
// ============================================================================

type Metrics struct {
	startTime time.Time
	// ConnectTime returns the duration between client initiation of the TCP/IP
	// connection to the server and actual TCP/IP session establishment.
	// This time includes DNS resolution and can be slightly higher if the DNS
	// resolution result was not in cache.
	ConnectTime time.Duration
	// LoginTime returns the between client initiation of the TCP/IP
	// connection to the server and the return of the login result.
	// This includes ConnectTime, but also XMPP level protocol negociation
	// like starttls.
	LoginTime time.Duration
}

// initMetrics set metrics with default value and define the starting point
// for duration calculation (connect time, login time, etc).
func initMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

func (m *Metrics) setConnectTime() {
	m.ConnectTime = time.Since(m.startTime)
}

func (m *Metrics) setLoginTime() {
	m.LoginTime = time.Since(m.startTime)
}

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
	var tcpconn net.Conn
	var err error

	// TODO: Refactor = abstract retry loop in capped exponential back-off function
	var try = 0
	var success bool
	c.Metrics = initMetrics()
	for try <= c.config.Retry && !success {
		if tcpconn, err = net.DialTimeout("tcp", c.config.Address, time.Duration(c.config.ConnectTimeout)*time.Second); err == nil {
			c.Metrics.setConnectTime()
			success = true
		}
		try++
	}
	if err != nil {
		return nil, err
	}

	// Connection is ok, we now open XMPP session
	c.conn = tcpconn
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

func (c *Client) recv(receiver chan<- interface{}) (err error) {
	for {
		val, err := next(c.Session.decoder)
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
