package xmpp

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"nhooyr.io/websocket"
)

type WebsocketTransport struct {
	wsConn  *websocket.Conn
	netConn net.Conn
	ctx     context.Context
}

func (t *WebsocketTransport) Connect(address string, c Config) error {
	t.ctx = context.Background()

	ctx, cancel := context.WithTimeout(t.ctx, time.Duration(c.ConnectTimeout)*time.Second)
	defer cancel()

	if !c.Insecure && strings.HasPrefix(address, "wss:") {
		return errors.New("Websocket address is not secure")
	}
	wsConn, _, err := websocket.Dial(ctx, address, nil)
	if err != nil {
		t.wsConn = wsConn
		t.netConn = websocket.NetConn(t.ctx, t.wsConn, websocket.MessageText)
	}
	return err
}

func (t WebsocketTransport) DoesStartTLS() bool {
	return false
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
