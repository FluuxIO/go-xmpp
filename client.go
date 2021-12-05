package xmpp

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"gosrc.io/xmpp/stanza"
)

//=============================================================================
// EventManager

// SyncConnState represents the current connection state.
type SyncConnState struct {
	sync.RWMutex
	// Current state of the client. Please use the dedicated getter and setter for this field as they are thread safe.
	state ConnState
}
type ConnState = uint8

// getState is a thread-safe getter for the current state
func (scs *SyncConnState) getState() ConnState {
	var res ConnState
	scs.RLock()
	res = scs.state
	scs.RUnlock()
	return res
}

// setState is a thread-safe setter for the current
func (scs *SyncConnState) setState(cs ConnState) {
	scs.Lock()
	scs.state = cs
	scs.Unlock()
}

// This is a the list of events happening on the connection that the
// client can be notified about.
const (
	StateDisconnected ConnState = iota
	StateResuming
	StateSessionEstablished
	StateStreamError
	StatePermanentError
	InitialPresence = "<presence/>"
)

// Event is a structure use to convey event changes related to client state. This
// is for example used to notify the client when the client get disconnected.
type Event struct {
	State       SyncConnState
	Description string
	StreamError string
	SMState     SMState
}

// SMState holds Stream Management information regarding the session that can be
// used to resume session after disconnect
type SMState struct {
	// Stream Management ID
	Id string
	// Inbound stanza count
	Inbound uint

	// IP affinity
	preferredReconAddr string

	// Error
	StreamErrorGroup stanza.StanzaErrorGroup

	// Track sent stanzas
	*stanza.UnAckQueue

	// TODO Store max and timestamp, to check if we should retry resumption or not
}

// EventHandler is use to pass events about state of the connection to
// client implementation.
type EventHandler func(Event) error

type EventManager struct {
	// Store current state. Please use "getState" and "setState" to access and/or modify this.
	CurrentState SyncConnState

	// Callback used to propagate connection state changes
	Handler EventHandler
}

// updateState changes the CurrentState in the event manager. The state read is threadsafe but there is no guarantee
// regarding the triggered callback function.
func (em *EventManager) updateState(state ConnState) {
	em.CurrentState.setState(state)
	if em.Handler != nil {
		em.Handler(Event{State: em.CurrentState})
	}
}

// disconnected changes the CurrentState in the event manager to "disconnected". The state read is threadsafe but there is no guarantee
// regarding the triggered callback function.
func (em *EventManager) disconnected(state SMState) {
	em.CurrentState.setState(StateDisconnected)
	if em.Handler != nil {
		em.Handler(Event{State: em.CurrentState, SMState: state})
	}
}

// streamError changes the CurrentState in the event manager to "streamError". The state read is threadsafe but there is no guarantee
// regarding the triggered callback function.
func (em *EventManager) streamError(error, desc string) {
	em.CurrentState.setState(StateStreamError)
	if em.Handler != nil {
		em.Handler(Event{State: em.CurrentState, StreamError: error, Description: desc})
	}
}

// Client
// ============================================================================

var ErrCanOnlySendGetOrSetIq = errors.New("SendIQ can only send get and set IQ stanzas")

// Client is the main structure used to connect as a client on an XMPP
// server.
type Client struct {
	// Store user defined options and states
	config *Config
	// Session gather data that can be accessed by users of this library
	Session   *Session
	transport Transport
	// Router is used to dispatch packets
	router *Router
	// Track and broadcast connection state
	EventManager
	// Handle errors from client execution
	ErrorHandler func(error)

	// Post connection hook. This will be executed on first connection
	PostConnectHook func() error

	// Post resume hook. This will be executed after the client resumes a lost connection using StreamManagement (XEP-0198)
	PostResumeHook func() error
}

/*
Setting up the client / Checking the parameters
*/

// NewClient generates a new XMPP client, based on Config passed as parameters.
// If host is not specified, the DNS SRV should be used to find the host from the domain part of the Jid.
// Default the port to 5222.
func NewClient(config *Config, r *Router, errorHandler func(error)) (c *Client, err error) {
	if config.KeepaliveInterval == 0 {
		config.KeepaliveInterval = time.Second * 30
	}
	// Parse Jid
	if config.parsedJid, err = stanza.NewJid(config.Jid); err != nil {
		err = errors.New("missing jid")
		return nil, NewConnError(err, true)
	}

	if config.Credential.secret == "" {
		err = errors.New("missing credential")
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
	if config.Domain == "" {
		// Fallback to jid domain
		config.Domain = config.parsedJid.Domain
	}

	c = new(Client)
	c.config = config
	c.router = r
	c.ErrorHandler = errorHandler

	if c.config.ConnectTimeout == 0 {
		c.config.ConnectTimeout = 15 // 15 second as default
	}

	if config.TransportConfiguration.Domain == "" {
		config.TransportConfiguration.Domain = config.parsedJid.Domain
	}
	c.config.TransportConfiguration.ConnectTimeout = c.config.ConnectTimeout
	c.transport = NewClientTransport(c.config.TransportConfiguration)

	if config.StreamLogger != nil {
		c.transport.LogTraffic(config.StreamLogger)
	}

	return
}

// Connect establishes a first time connection to a XMPP server.
// It calls the PostConnectHook
func (c *Client) Connect() error {
	err := c.connect()
	if err != nil {
		return err
	}
	// TODO: Do we always want to send initial presence automatically ?
	// Do we need an option to avoid that or do we rely on client to send the presence itself ?
	err = c.sendWithWriter(c.transport, []byte(InitialPresence))
	// Execute the post first connection hook. Typically this holds "ask for roster" and this type of actions.
	if c.PostConnectHook != nil {
		err = c.PostConnectHook()
		if err != nil {
			return err
		}
	}

	// Start the keepalive go routine
	keepaliveQuit := make(chan struct{})
	go keepalive(c.transport, c.config.KeepaliveInterval, keepaliveQuit)
	// Start the receiver go routine
	go c.recv(keepaliveQuit)
	return err
}

// connect establishes an actual TCP connection, based on previously defined parameters, as well as a XMPP session
func (c *Client) connect() error {
	var state SMState
	var err error
	// This is the TCP connection
	streamId, err := c.transport.Connect()
	if err != nil {
		return err
	}

	// Client is ok, we now open XMPP session with TLS negotiation if possible and session resume or binding
	// depending on state.
	if c.Session, err = NewSession(c, state); err != nil {
		// Try to get the stream close tag from the server.
		go func() {
			for {
				val, err := stanza.NextPacket(c.transport.GetDecoder())
				if err != nil {
					c.ErrorHandler(err)
					c.disconnected(state)
					return
				}
				switch val.(type) {
				case stanza.StreamClosePacket:
					// TCP messages should arrive in order, so we can expect to get nothing more after this occurs
					c.transport.ReceivedStreamClose()
					return
				}
			}
		}()
		c.Disconnect()
		return err
	}
	c.Session.StreamId = streamId
	c.updateState(StateSessionEstablished)

	return err
}

// Resume attempts resuming  a Stream Managed session, based on the provided stream management
// state. See XEP-0198
func (c *Client) Resume() error {
	c.EventManager.updateState(StateResuming)
	err := c.connect()
	if err != nil {
		return err
	}
	// Execute post reconnect hook. This can be different from the first connection hook, and not trigger roster retrieval
	// for example.
	if c.PostResumeHook != nil {
		err = c.PostResumeHook()
	}
	return err
}

// Disconnect disconnects the client from the server, sending a stream close nonza and closing the TCP connection.
func (c *Client) Disconnect() error {
	if c.transport != nil {
		return c.transport.Close()
	}
	// No transport so no connection.
	return nil
}

func (c *Client) SetHandler(handler EventHandler) {
	c.Handler = handler
}

// Send marshals XMPP stanza and sends it to the server.
func (c *Client) Send(packet stanza.Packet) error {
	conn := c.transport
	if conn == nil {
		return errors.New("client is not connected")
	}

	data, err := xml.Marshal(packet)
	if err != nil {
		return errors.New("cannot marshal packet " + err.Error())
	}

	// Store stanza as non-acked as part of stream management
	// See https://xmpp.org/extensions/xep-0198.html#scenarios
	if c.config.StreamManagementEnable {
		if _, ok := packet.(stanza.SMRequest); !ok {
			toStore := stanza.UnAckedStz{Stz: string(data)}
			c.Session.SMState.UnAckQueue.Push(&toStore)
		}
	}

	return c.sendWithWriter(c.transport, data)
}

// SendIQ sends an IQ set or get stanza to the server. If a result is received
// the provided handler function will automatically be called.
//
// The provided context should have a timeout to prevent the client from waiting
// forever for an IQ result. For example:
//
//   ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)
//   result := <- client.SendIQ(ctx, iq)
//
func (c *Client) SendIQ(ctx context.Context, iq *stanza.IQ) (chan stanza.IQ, error) {
	if iq.Attrs.Type != stanza.IQTypeSet && iq.Attrs.Type != stanza.IQTypeGet {
		return nil, ErrCanOnlySendGetOrSetIq
	}
	if err := c.Send(iq); err != nil {
		return nil, err
	}
	return c.router.NewIQResultRoute(ctx, iq.Attrs.Id), nil
}

// SendRaw sends an XMPP stanza as a string to the server.
// It can be invalid XML or XMPP content. In that case, the server will
// disconnect the client. It is up to the user of this method to
// carefully craft the XML content to produce valid XMPP.
func (c *Client) SendRaw(packet string) error {
	conn := c.transport
	if conn == nil {
		return errors.New("client is not connected")
	}

	// Store stanza as non-acked as part of stream management
	// See https://xmpp.org/extensions/xep-0198.html#scenarios
	if c.config.StreamManagementEnable {
		toStore := stanza.UnAckedStz{Stz: packet}
		c.Session.SMState.UnAckQueue.Push(&toStore)
	}
	return c.sendWithWriter(c.transport, []byte(packet))
}

func (c *Client) sendWithWriter(writer io.Writer, packet []byte) error {
	var err error
	_, err = writer.Write(packet)
	return err
}

// ============================================================================
// Go routines

// Loop: Receive data from server
func (c *Client) recv(keepaliveQuit chan<- struct{}) {
	defer close(keepaliveQuit)

	for {
		val, err := stanza.NextPacket(c.transport.GetDecoder())
		if err != nil {
			c.ErrorHandler(err)
			c.disconnected(c.Session.SMState)
			return
		}

		// Handle stream errors
		switch packet := val.(type) {
		case stanza.StreamError:
			c.router.route(c, val)
			c.streamError(packet.Error.Local, packet.Text)
			c.ErrorHandler(errors.New("stream error: " + packet.Error.Local))
			// We don't return here, because we want to wait for the stream close tag from the server, or timeout.
			c.Disconnect()
		// Process Stream management nonzas
		case stanza.SMRequest:
			answer := stanza.SMAnswer{XMLName: xml.Name{
				Space: stanza.NSStreamManagement,
				Local: "a",
			}, H: c.Session.SMState.Inbound}
			err = c.Send(answer)
			if err != nil {
				c.ErrorHandler(err)
				return
			}
		case stanza.StreamClosePacket:
			// TCP messages should arrive in order, so we can expect to get nothing more after this occurs
			c.transport.ReceivedStreamClose()
			return
		default:
			c.Session.SMState.Inbound++
		}
		// Do normal route processing in a go-routine so we can immediately
		// start receiving other stanzas. This also allows route handlers to
		// send and receive more stanzas.
		go c.router.route(c, val)
	}
}

// Loop: send whitespace keepalive to server
// This is use to keep the connection open, but also to detect connection loss
// and trigger proper client connection shutdown.
func keepalive(transport Transport, interval time.Duration, quit <-chan struct{}) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			if err := transport.Ping(); err != nil {
				// When keepalive fails, we force close the transport. In all cases, the recv will also fail.
				ticker.Stop()
				_ = transport.Close()
				return
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
