package xmpp

import (
	"context"
	"net"
	"strings"
	"time"

	"nhooyr.io/websocket"
)

type WebsocketTransport struct {
	Config  TransportConfiguration
	wsConn  *websocket.Conn
	netConn net.Conn
	ctx     context.Context
}

func (t *WebsocketTransport) Connect() error {
	t.ctx = context.Background()

	ctx, cancel := context.WithTimeout(t.ctx, time.Duration(t.Config.ConnectTimeout)*time.Second)
	defer cancel()

	wsConn, _, err := websocket.Dial(ctx, t.Config.Address, nil)
	if err != nil {
		t.wsConn = wsConn
		t.netConn = websocket.NetConn(t.ctx, t.wsConn, websocket.MessageText)
	}
	return err
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

func (t WebsocketTransport) Read(p []byte) (n int, err error) {
	return t.netConn.Read(p)
}

func (t WebsocketTransport) Write(p []byte) (n int, err error) {
	return t.netConn.Write(p)
}

func (t WebsocketTransport) Close() error {
	return t.netConn.Close()
}
