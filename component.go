package xmpp

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"gosrc.io/xmpp/stanza"
	"io"
)

type ComponentOptions struct {
	TransportConfiguration

	// =================================
	// Component Connection Info

	// Domain is the XMPP server subdomain that the component will handle
	Domain string
	// Secret is the "password" used by the XMPP server to secure component access
	Secret string

	// =================================
	// Component discovery

	// Component human readable name, that will be shown in XMPP discovery
	Name string
	// Typical categories and types: https://xmpp.org/registrar/disco-categories.html
	Category string
	Type     string

	// =================================
	// Communication with developer client / StreamManager

	// Track and broadcast connection state
	EventManager
}

// Component implements an XMPP extension allowing to extend XMPP server
// using external components. Component specifications are defined
// in XEP-0114, XEP-0355 and XEP-0356.
type Component struct {
	ComponentOptions
	router *Router

	transport Transport

	// read / write
	socketProxy  io.ReadWriter // TODO
	ErrorHandler func(error)
}

func NewComponent(opts ComponentOptions, r *Router, errorHandler func(error)) (*Component, error) {
	c := Component{ComponentOptions: opts, router: r, ErrorHandler: errorHandler}
	return &c, nil
}

// Connect triggers component connection to XMPP server component port.
// TODO: Failed handshake should be a permanent error
func (c *Component) Connect() error {
	return c.Resume()
}

func (c *Component) Resume() error {
	var err error
	var streamId string
	if c.ComponentOptions.TransportConfiguration.Domain == "" {
		c.ComponentOptions.TransportConfiguration.Domain = c.ComponentOptions.Domain
	}
	c.transport, err = NewComponentTransport(c.ComponentOptions.TransportConfiguration)
	if err != nil {
		c.updateState(StatePermanentError)
		return NewConnError(err, true)
	}

	if streamId, err = c.transport.Connect(); err != nil {
		c.updateState(StatePermanentError)
		return NewConnError(err, true)
	}

	// Authentication
	if err := c.sendWithWriter(c.transport, []byte(fmt.Sprintf("<handshake>%s</handshake>", c.handshake(streamId)))); err != nil {
		c.updateState(StateStreamError)

		return NewConnError(errors.New("cannot send handshake "+err.Error()), false)
	}

	// Check server response for authentication
	val, err := stanza.NextPacket(c.transport.GetDecoder())
	if err != nil {
		c.updateState(StatePermanentError)
		return NewConnError(err, true)
	}

	switch v := val.(type) {
	case stanza.StreamError:
		c.streamError("conflict", "no auth loop")
		return NewConnError(errors.New("handshake failed "+v.Error.Local), true)
	case stanza.Handshake:
		// Start the receiver go routine
		c.updateState(StateSessionEstablished)
		go c.recv()
		return err // Should be empty at this point
	default:
		c.updateState(StatePermanentError)
		return NewConnError(errors.New("expecting handshake result, got "+v.Name()), true)
	}
}

func (c *Component) Disconnect() error {
	// TODO: Add a way to wait for stream close acknowledgement from the server for clean disconnect
	if c.transport != nil {
		return c.transport.Close()
	}
	// No transport so no connection.
	return nil
}

func (c *Component) SetHandler(handler EventHandler) {
	c.Handler = handler
}

// Receiver Go routine receiver
func (c *Component) recv() {
	for {
		val, err := stanza.NextPacket(c.transport.GetDecoder())
		if err != nil {
			c.updateState(StateDisconnected)
			c.ErrorHandler(err)
			return
		}
		// Handle stream errors
		switch p := val.(type) {
		case stanza.StreamError:
			c.router.route(c, val)
			c.streamError(p.Error.Local, p.Text)
			c.ErrorHandler(errors.New("stream error: " + p.Error.Local))
			// We don't return here, because we want to wait for the stream close tag from the server, or timeout.
			c.Disconnect()
		case stanza.StreamClosePacket:
			// TCP messages should arrive in order, so we can expect to get nothing more after this occurs
			c.transport.ReceivedStreamClose()
			return
		}
		c.router.route(c, val)
	}
}

// Send marshalls XMPP stanza and sends it to the server.
func (c *Component) Send(packet stanza.Packet) error {
	transport := c.transport
	if transport == nil {
		return errors.New("component is not connected")
	}

	data, err := xml.Marshal(packet)
	if err != nil {
		return errors.New("cannot marshal packet " + err.Error())
	}

	if err := c.sendWithWriter(transport, data); err != nil {
		return errors.New("cannot send packet " + err.Error())
	}
	return nil
}

func (c *Component) sendWithWriter(writer io.Writer, packet []byte) error {
	var err error
	_, err = writer.Write(packet)
	return err
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
func (c *Component) SendIQ(ctx context.Context, iq *stanza.IQ) (chan stanza.IQ, error) {
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
// disconnect the component. It is up to the user of this method to
// carefully craft the XML content to produce valid XMPP.
func (c *Component) SendRaw(packet string) error {
	transport := c.transport
	if transport == nil {
		return errors.New("component is not connected")
	}

	var err error
	err = c.sendWithWriter(transport, []byte(packet))
	return err
}

// handshake generates an authentication token based on StreamID and shared secret.
func (c *Component) handshake(streamId string) string {
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

/*
TODO: Add support for discovery management directly in component
TODO: Support multiple identities on disco info
TODO: Support returning features on disco info
*/
