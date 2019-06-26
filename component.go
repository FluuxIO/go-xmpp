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

	"gosrc.io/xmpp/stanza"
)

const componentStreamOpen = "<?xml version='1.0'?><stream:stream to='%s' xmlns='%s' xmlns:stream='%s'>"

type ComponentOptions struct {
	// =================================
	// Component Connection Info

	// Domain is the XMPP server subdomain that the component will handle
	Domain string
	// Secret is the "password" used by the XMPP server to secure component access
	Secret string
	// Address is the XMPP Host and port to connect to. Host is of
	// the form 'serverhost:port' i.e "localhost:8888"
	Address string

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

	// TCP level connection
	conn net.Conn

	// read / write
	socketProxy io.ReadWriter // TODO
	decoder     *xml.Decoder
}

func NewComponent(opts ComponentOptions, r *Router) (*Component, error) {
	c := Component{ComponentOptions: opts, router: r}
	return &c, nil
}

// Connect triggers component connection to XMPP server component port.
// TODO: Failed handshake should be a permanent error
func (c *Component) Connect() error {
	var conn net.Conn
	var err error
	if conn, err = net.DialTimeout("tcp", c.Address, time.Duration(5)*time.Second); err != nil {
		return err
	}
	c.conn = conn

	// 1. Send stream open tag
	if _, err := fmt.Fprintf(conn, componentStreamOpen, c.Domain, stanza.NSComponent, stanza.NSStream); err != nil {
		return errors.New("cannot send stream open " + err.Error())
	}
	c.decoder = xml.NewDecoder(conn)

	// 2. Initialize xml decoder and extract streamID from reply
	streamId, err := stanza.InitStream(c.decoder)
	if err != nil {
		return errors.New("cannot init decoder " + err.Error())
	}

	// 3. Authentication
	if _, err := fmt.Fprintf(conn, "<handshake>%s</handshake>", c.handshake(streamId)); err != nil {
		return errors.New("cannot send handshake " + err.Error())
	}

	// 4. Check server response for authentication
	val, err := stanza.NextPacket(c.decoder)
	if err != nil {
		return err
	}

	switch v := val.(type) {
	case stanza.StreamError:
		return errors.New("handshake failed " + v.Error.Local)
	case stanza.Handshake:
		// Start the receiver go routine
		go c.recv()
		return nil
	default:
		return errors.New("expecting handshake result, got " + v.Name())
	}
}

func (c *Component) Disconnect() {
	_ = c.SendRaw("</stream:stream>")
	// TODO: Add a way to wait for stream close acknowledgement from the server for clean disconnect
	_ = c.conn.Close()
}

func (c *Component) SetHandler(handler EventHandler) {
	c.Handler = handler
}

// Receiver Go routine receiver
func (c *Component) recv() (err error) {
	for {
		val, err := stanza.NextPacket(c.decoder)
		if err != nil {
			c.updateState(StateDisconnected)
			return err
		}

		// Handle stream errors
		switch p := val.(type) {
		case stanza.StreamError:
			c.router.route(c, val)
			c.streamError(p.Error.Local, p.Text)
			return errors.New("stream error: " + p.Error.Local)
		}
		c.router.route(c, val)
	}
}

// Send marshalls XMPP stanza and sends it to the server.
func (c *Component) Send(packet stanza.Packet) error {
	conn := c.conn
	if conn == nil {
		return errors.New("component is not connected")
	}

	data, err := xml.Marshal(packet)
	if err != nil {
		return errors.New("cannot marshal packet " + err.Error())
	}

	if _, err := fmt.Fprintf(conn, string(data)); err != nil {
		return errors.New("cannot send packet " + err.Error())
	}
	return nil
}

// SendRaw sends an XMPP stanza as a string to the server.
// It can be invalid XML or XMPP content. In that case, the server will
// disconnect the component. It is up to the user of this method to
// carefully craft the XML content to produce valid XMPP.
func (c *Component) SendRaw(packet string) error {
	conn := c.conn
	if conn == nil {
		return errors.New("component is not connected")
	}

	var err error
	_, err = fmt.Fprintf(c.conn, packet)
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
