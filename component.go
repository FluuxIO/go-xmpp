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

// Connect triggers component connection to XMPP server component port.
// TODO Helper to prepare connection string
func (c *Component) Connect(connStr string) error {
	var conn net.Conn
	var err error
	if conn, err = net.DialTimeout("tcp", connStr, time.Duration(5)*time.Second); err != nil {
		return err
	}
	c.conn = conn

	// 1. Send stream open tag
	if _, err := fmt.Fprintf(conn, componentStreamOpen, c.Host, NSComponent, NSStream); err != nil {
		return errors.New("cannot send stream open " + err.Error())
	}
	c.decoder = xml.NewDecoder(conn)

	// 2. Initialize xml decoder and extract streamID from reply
	streamId, err := initDecoder(c.decoder)
	if err != nil {
		return errors.New("cannot init decoder " + err.Error())
	}

	// 3. Authentication
	if _, err := fmt.Fprintf(conn, "<handshake>%s</handshake>", c.handshake(streamId)); err != nil {
		return errors.New("cannot send handshake " + err.Error())
	}

	// 4. Check server response for authentication
	val, err := next(c.decoder)
	if err != nil {
		return err
	}

	switch v := val.(type) {
	case *StreamError:
		return errors.New("handshake failed " + v.Error.Local)
	case *Handshake:
		return nil
	default:
		return errors.New("unexpected packet, got " + v.Name())
	}
	panic("unreachable")
}

// ReadPacket reads next incoming XMPP packet
func (c *Component) ReadPacket() (Packet, error) {
	// TODO use defined interface Packet
	return next(c.decoder)
}

// Send marshalls XMPP stanza and sends it to the server.
func (c *Component) Send(packet Packet) error {
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
// disconnect the component. It is up to the user of this method to
// carefully craft the XML content to produce valid XMPP.
func (c *Component) SendRaw(packet string) error {
	fmt.Fprintf(c.conn, packet) // TODO handle errors
	return nil
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

// ============================================================================
// Handshake Stanza

// Handshake is a stanza used by XMPP components to authenticate on XMPP
// component port.
type Handshake struct {
	XMLName xml.Name `xml:"jabber:component:accept handshake"`
	// TODO Add handshake value with test for proper serialization
	// Value string     `xml:",innerxml"`
}

func (Handshake) Name() string {
	return "component:handshake"
}

// Handshake decoding wrapper

type handshakeDecoder struct{}

var handshake handshakeDecoder

func (handshakeDecoder) decode(p *xml.Decoder, se xml.StartElement) (Handshake, error) {
	var packet Handshake
	err := p.DecodeElement(&packet, &se)
	return packet, err
}
