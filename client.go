package xmpp

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"time"

	"gosrc.io/xmpp/stanza"
)

//=============================================================================
// EventManager

// ConnState represents the current connection state.
type ConnState = uint8

// This is a the list of events happening on the connection that the
// client can be notified about.
const (
	StateDisconnected ConnState = iota
	StateConnected
	StateSessionEstablished
	StateStreamError
)

// Event is a structure use to convey event changes related to client state. This
// is for example used to notify the client when the client get disconnected.
type Event struct {
	State       ConnState
	Description string
	StreamError string
}

// EventHandler is use to pass events about state of the connection to
// client implementation.
type EventHandler func(Event)

type EventManager struct {
	// Store current state
	CurrentState ConnState

	// Callback used to propagate connection state changes
	Handler EventHandler
}

func (em EventManager) updateState(state ConnState) {
	em.CurrentState = state
	if em.Handler != nil {
		em.Handler(Event{State: em.CurrentState})
	}
}

func (em EventManager) streamError(error, desc string) {
	em.CurrentState = StateStreamError
	if em.Handler != nil {
		em.Handler(Event{State: em.CurrentState, StreamError: error, Description: desc})
	}
}

// Client
// ============================================================================

// Client is the main structure used to connect as a client on an XMPP
// server.
type Client struct {
	// Store user defined options and states
	config Config
	// Session gather data that can be accessed by users of this library
	Session *Session
	// TCP level connection / can be replaced by a TLS session after starttls
	conn net.Conn
	// Router is used to dispatch packets
	router *Router
	// Track and broadcast connection state
	EventManager
}

/*
Setting up the client / Checking the parameters
*/

// NewClient generates a new XMPP client, based on Config passed as parameters.
// If host is not specified, the DNS SRV should be used to find the host from the domainpart of the JID.
// Default the port to 5222.
func NewClient(config Config, r *Router) (c *Client, err error) {
	// Parse JID
	if config.parsedJid, err = NewJid(config.Jid); err != nil {
		err = errors.New("missing jid")
		return nil, NewConnError(err, true)
	}

	if config.Password == "" {
		err = errors.New("missing password")
		return nil, NewConnError(err, true)
	}

	// Fallback to jid domain
	if config.Address == "" {
		config.Address = config.parsedJid.Domain

		// Fetch SRV DNS-Entries
		_, srvEntries, err := net.LookupSRV("xmpp-client", "tcp", config.parsedJid.Domain)

		if err == nil && len(srvEntries) > 0 {
			// If we found matching DNS records, use the entry with highest weight
			bestSrv := srvEntries[0]
			for _, srv := range srvEntries {
				if srv.Priority <= bestSrv.Priority && srv.Weight >= bestSrv.Weight {
					bestSrv = srv
					config.Address = ensurePort(srv.Target, int(srv.Port))
				}
			}
		}
	}
	config.Address = ensurePort(config.Address, 5222)

	c = new(Client)
	c.config = config
	c.router = r

	if c.config.ConnectTimeout == 0 {
		c.config.ConnectTimeout = 15 // 15 second as default
	}

	return
}

// Connect triggers actual TCP connection, based on previously defined parameters.
func (c *Client) Connect() error {
	var err error

	c.conn, err = net.DialTimeout("tcp", c.config.Address, time.Duration(c.config.ConnectTimeout)*time.Second)
	if err != nil {
		return err
	}
	c.updateState(StateConnected)

	// Client is ok, we now open XMPP session
	if c.conn, c.Session, err = NewSession(c.conn, c.config); err != nil {
		return err
	}
	c.updateState(StateSessionEstablished)

	// We're connected and can now receive and send messages.
	//fmt.Fprintf(client.conn, "<presence xml:lang='en'><show>%s</show><status>%s</status></presence>", "chat", "Online")
	// TODO: Do we always want to send initial presence automatically ?
	// Do we need an option to avoid that or do we rely on client to send the presence itself ?
	fmt.Fprintf(c.Session.streamLogger, "<presence/>")

	// Start the keepalive go routine
	keepaliveQuit := make(chan struct{})
	go keepalive(c.conn, keepaliveQuit)
	// Start the receiver go routine
	go c.recv(keepaliveQuit)

	return err
}

func (c *Client) Disconnect() {
	_ = c.SendRaw("</stream:stream>")
	// TODO: Add a way to wait for stream close acknowledgement from the server for clean disconnect
	_ = c.conn.Close()
}

func (c *Client) SetHandler(handler EventHandler) {
	c.Handler = handler
}

// Send marshals XMPP stanza and sends it to the server.
func (c *Client) Send(packet stanza.Packet) error {
	conn := c.conn
	if conn == nil {
		return errors.New("client is not connected")
	}

	data, err := xml.Marshal(packet)
	if err != nil {
		return errors.New("cannot marshal packet " + err.Error())
	}

	return c.sendWithLogger(string(data))
}

// SendRaw sends an XMPP stanza as a string to the server.
// It can be invalid XML or XMPP content. In that case, the server will
// disconnect the client. It is up to the user of this method to
// carefully craft the XML content to produce valid XMPP.
func (c *Client) SendRaw(packet string) error {
	conn := c.conn
	if conn == nil {
		return errors.New("client is not connected")
	}

	return c.sendWithLogger(packet)
}

func (c *Client) sendWithLogger(packet string) error {
	var err error
	_, err = fmt.Fprintf(c.Session.streamLogger, packet)
	return err
}

// ============================================================================
// Go routines

// Loop: Receive data from server
func (c *Client) recv(keepaliveQuit chan<- struct{}) (err error) {
	for {
		val, err := stanza.NextPacket(c.Session.decoder)
		if err != nil {
			close(keepaliveQuit)
			c.updateState(StateDisconnected)
			return err
		}

		// Handle stream errors
		switch packet := val.(type) {
		case stanza.StreamError:
			c.router.route(c, val)
			close(keepaliveQuit)
			c.streamError(packet.Error.Local, packet.Text)
			return errors.New("stream error: " + packet.Error.Local)
		}

		c.router.route(c, val)
	}
}

// Loop: send whitespace keepalive to server
// This is use to keep the connection open, but also to detect connection loss
// and trigger proper client connection shutdown.
func keepalive(conn net.Conn, quit <-chan struct{}) {
	// TODO: Make keepalive interval configurable
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			if n, err := fmt.Fprintf(conn, "\n"); err != nil || n != 1 {
				// When keep alive fails, we force close the connection. In all cases, the recv will also fail.
				ticker.Stop()
				_ = conn.Close()
				return
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
