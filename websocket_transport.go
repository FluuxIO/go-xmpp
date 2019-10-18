package xmpp

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"gosrc.io/xmpp/stanza"
	"nhooyr.io/websocket"
)

const pingTimeout = time.Duration(5) * time.Second

var ServerDoesNotSupportXmppOverWebsocket = errors.New("The websocket server does not support the xmpp subprotocol")

type WebsocketTransport struct {
	Config  TransportConfiguration
	decoder *xml.Decoder
	wsConn  *websocket.Conn
	netConn net.Conn
	logFile io.Writer
}

func (t *WebsocketTransport) Connect() (string, error) {
	ctx := context.Background()

	if t.Config.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(t.Config.ConnectTimeout)*time.Second)
		defer cancel()
	}

	wsConn, response, err := websocket.Dial(ctx, t.Config.Address, &websocket.DialOptions{
		Subprotocols: []string{"xmpp"},
	})
	if err != nil {
		return "", NewConnError(err, true)
	}
	if response.Header.Get("Sec-WebSocket-Protocol") != "xmpp" {
		_ = wsConn.Close(websocket.StatusBadGateway, "Could not negotiate XMPP subprotocol")
		return "", NewConnError(ServerDoesNotSupportXmppOverWebsocket, true)
	}

	t.wsConn = wsConn
	t.netConn = websocket.NetConn(ctx, t.wsConn, websocket.MessageText)

	handshake := fmt.Sprintf("<open xmlns=\"urn:ietf:params:xml:ns:xmpp-framing\" to=\"%s\" version=\"1.0\" />", t.Config.Domain)
	if _, err = t.Write([]byte(handshake)); err != nil {
		_ = wsConn.Close(websocket.StatusBadGateway, "XMPP handshake error")
		return "", NewConnError(err, false)
	}

	handshakeResponse := make([]byte, 2048)
	if _, err = t.Read(handshakeResponse); err != nil {
		_ = wsConn.Close(websocket.StatusBadGateway, "XMPP handshake error")
		return "", NewConnError(err, false)
	}

	var openResponse = stanza.WebsocketOpen{}
	if err = xml.Unmarshal(handshakeResponse, &openResponse); err != nil {
		_ = wsConn.Close(websocket.StatusBadGateway, "XMPP handshake error")
		return "", NewConnError(err, false)
	}

	t.decoder = xml.NewDecoder(t)
	t.decoder.CharsetReader = t.Config.CharsetReader

	return openResponse.Id, nil
}

func (t WebsocketTransport) StartTLS() error {
	return TLSNotSupported
}

func (t WebsocketTransport) DoesStartTLS() bool {
	return false
}

func (t WebsocketTransport) GetDecoder() *xml.Decoder {
	return t.decoder
}

func (t WebsocketTransport) IsSecure() bool {
	return strings.HasPrefix(t.Config.Address, "wss:")
}

func (t WebsocketTransport) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	return t.wsConn.Ping(ctx)
}

func (t *WebsocketTransport) Read(p []byte) (n int, err error) {
	n, err = t.netConn.Read(p)
	if t.logFile != nil && n > 0 {
		_, _ = fmt.Fprintf(t.logFile, "RECV:\n%s\n\n", p)
	}
	return
}

func (t WebsocketTransport) Write(p []byte) (n int, err error) {
	if t.logFile != nil {
		_, _ = fmt.Fprintf(t.logFile, "SEND:\n%s\n\n", p)
	}
	return t.netConn.Write(p)
}

func (t WebsocketTransport) Close() error {
	t.Write([]byte("<close xmlns=\"urn:ietf:params:xml:ns:xmpp-framing\" />"))
	return t.netConn.Close()
}

func (t *WebsocketTransport) LogTraffic(logFile io.Writer) {
	t.logFile = logFile
}
