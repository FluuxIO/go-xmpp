package xmpp

import (
	"bufio"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"gosrc.io/xmpp/stanza"
	"nhooyr.io/websocket"
)

const maxPacketSize = 32768

const pingTimeout = time.Duration(5) * time.Second

var ServerDoesNotSupportXmppOverWebsocket = errors.New("the websocket server does not support the xmpp subprotocol")

// The decoder is expected to be initialized after connecting to a server.
type WebsocketTransport struct {
	Config  TransportConfiguration
	decoder *xml.Decoder
	wsConn  *websocket.Conn
	queue   chan []byte
	logFile io.Writer

	closeCtx  context.Context
	closeFunc context.CancelFunc
}

func (t *WebsocketTransport) Connect() (string, error) {
	t.queue = make(chan []byte, 256)
	t.closeCtx, t.closeFunc = context.WithCancel(context.Background())

	var ctx context.Context
	ctx = context.Background()
	if t.Config.ConnectTimeout > 0 {
		var cancelConnect context.CancelFunc
		ctx, cancelConnect = context.WithTimeout(t.closeCtx, time.Duration(t.Config.ConnectTimeout)*time.Second)
		defer cancelConnect()
	}

	wsConn, response, err := websocket.Dial(ctx, t.Config.Address, &websocket.DialOptions{
		Subprotocols: []string{"xmpp"},
	})

	if err != nil {
		return "", NewConnError(err, true)
	}
	if response.Header.Get("Sec-WebSocket-Protocol") != "xmpp" {
		t.cleanup(websocket.StatusBadGateway)
		return "", NewConnError(ServerDoesNotSupportXmppOverWebsocket, true)
	}

	wsConn.SetReadLimit(maxPacketSize)
	t.wsConn = wsConn
	t.startReader()

	t.decoder = xml.NewDecoder(bufio.NewReaderSize(t, maxPacketSize))
	t.decoder.CharsetReader = t.Config.CharsetReader

	return t.StartStream()
}

func (t WebsocketTransport) StartStream() (string, error) {
	if _, err := fmt.Fprintf(t, `<open xmlns="urn:ietf:params:xml:ns:xmpp-framing" to="%s" version="1.0" />`, t.Config.Domain); err != nil {
		t.cleanup(websocket.StatusBadGateway)
		return "", NewConnError(err, true)
	}

	sessionID, err := stanza.InitStream(t.GetDecoder())
	if err != nil {
		t.Close()
		return "", NewConnError(err, false)
	}
	return sessionID, nil
}

// startReader runs a go function that keeps reading from the websocket. This
// is required to allow Ping() to work: Ping requires a Reader to be running
// to process incoming control frames.
func (t WebsocketTransport) startReader() {
	go func() {
		buffer := make([]byte, maxPacketSize)
		for {
			_, reader, err := t.wsConn.Reader(t.closeCtx)
			if err != nil {
				return
			}
			n, err := reader.Read(buffer)
			if err != nil && err != io.EOF {
				return
			}
			if n > 0 {
				// We need to make a copy, otherwise we will overwrite the slice content
				// on the next iteration of the for loop.
				tmp := make([]byte, n)
				copy(tmp, buffer)
				t.queue <- tmp
			}
		}
	}()
}

func (t WebsocketTransport) StartTLS() error {
	return ErrTLSNotSupported
}

func (t WebsocketTransport) DoesStartTLS() bool {
	return false
}

func (t WebsocketTransport) GetDomain() string {
	return t.Config.Domain
}

func (t WebsocketTransport) GetDecoder() *xml.Decoder {
	return t.decoder
}

func (t WebsocketTransport) IsSecure() bool {
	return strings.HasPrefix(t.Config.Address, "wss:")
}

func (t WebsocketTransport) Ping() error {
	ctx, cancel := context.WithTimeout(t.closeCtx, pingTimeout)
	defer cancel()
	return t.wsConn.Ping(ctx)
}

func (t *WebsocketTransport) Read(p []byte) (int, error) {
	select {
	case <-t.closeCtx.Done():
		return 0, t.closeCtx.Err()
	case data := <-t.queue:
		if t.logFile != nil && len(data) > 0 {
			_, _ = fmt.Fprintf(t.logFile, "RECV:\n%s\n\n", data)
		}
		copy(p, data)
		return len(data), nil
	}
}

func (t WebsocketTransport) Write(p []byte) (int, error) {
	if t.logFile != nil {
		_, _ = fmt.Fprintf(t.logFile, "SEND:\n%s\n\n", p)
	}
	return len(p), t.wsConn.Write(t.closeCtx, websocket.MessageText, p)
}

func (t WebsocketTransport) Close() error {
	t.Write([]byte("<close xmlns=\"urn:ietf:params:xml:ns:xmpp-framing\" />"))
	return t.cleanup(websocket.StatusGoingAway)
}

func (t *WebsocketTransport) LogTraffic(logFile io.Writer) {
	t.logFile = logFile
}

func (t *WebsocketTransport) cleanup(code websocket.StatusCode) error {
	var err error
	if t.queue != nil {
		close(t.queue)
		t.queue = nil
	}
	if t.wsConn != nil {
		err = t.wsConn.Close(code, "Done")
		t.wsConn = nil
	}
	if t.closeFunc != nil {
		t.closeFunc()
		t.closeFunc = nil
		t.closeCtx = nil
	}
	return err
}

// ReceivedStreamClose is not used for websockets for now
func (t *WebsocketTransport) ReceivedStreamClose() {
	return
}
