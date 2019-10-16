package xmpp

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"nhooyr.io/websocket"
)

const pingTimeout = time.Duration(5) * time.Second

var ServerDoesNotSupportXmppOverWebsocket = errors.New("The websocket server does not support the xmpp subprotocol")

type WebsocketTransport struct {
	Config  TransportConfiguration
	wsConn  *websocket.Conn
	netConn net.Conn
	ctx     context.Context
}

func (t *WebsocketTransport) Connect() error {
	t.ctx = context.Background()

	if t.Config.ConnectTimeout > 0 {
		ctx, cancel := context.WithTimeout(t.ctx, time.Duration(t.Config.ConnectTimeout)*time.Second)
		t.ctx = ctx
		defer cancel()
	}

	wsConn, response, err := websocket.Dial(t.ctx, t.Config.Address, &websocket.DialOptions{
		Subprotocols: []string{"xmpp"},
	})
	if err != nil {
		return NewConnError(err, true)
	}
	if response.Header.Get("Sec-WebSocket-Protocol") != "xmpp" {
		return ServerDoesNotSupportXmppOverWebsocket
	}
	t.wsConn = wsConn
	t.netConn = websocket.NetConn(t.ctx, t.wsConn, websocket.MessageText)
	return nil
}

func (t WebsocketTransport) StartTLS(domain string) error {
	return TLSNotSupported
}

func (t WebsocketTransport) DoesStartTLS() bool {
	return false
}

func (t WebsocketTransport) IsSecure() bool {
	return strings.HasPrefix(t.Config.Address, "wss:")
}

func (t WebsocketTransport) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	// Note that we do not use wsConn.Ping(), because not all websocket servers
	// (ejabberd for example) implement ping frames
	return t.wsConn.Write(ctx, websocket.MessageText, []byte(" "))
}

func (t WebsocketTransport) Read(p []byte) (n int, err error) {
	return t.netConn.Read(p)
}

func (t WebsocketTransport) Write(p []byte) (n int, err error) {
	return t.netConn.Write(p)
}

func (t WebsocketTransport) Close() error {
	return t.netConn.Close()
}
