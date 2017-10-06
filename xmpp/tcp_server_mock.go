package xmpp

import (
	"net"
	"testing"
)

//=============================================================================
// TCP Server Mock

// ClientHandler is passed by the test client to provide custom behaviour to
// the TCP server mock. This allows customizing the server behaviour to allow
// testing clients under various scenarii.
type ClientHandler func(t *testing.T, conn net.Conn)

// ServerMock is a simple TCP server that can be use to mock basic server
// behaviour to test clients.
type ServerMock struct {
	t           *testing.T
	handler     ClientHandler
	listener    net.Listener
	connections []net.Conn
	done        chan struct{}
}

// Start launches the mock TCP server, listening to an actual address / port.
func (mock *ServerMock) Start(t *testing.T, addr string, handler ClientHandler) {
	mock.t = t
	mock.handler = handler
	if err := mock.init(addr); err != nil {
		return
	}
	go mock.loop()
}

func (mock *ServerMock) Stop() {
	close(mock.done)
	if mock.listener != nil {
		mock.listener.Close()
	}
	// Close all existing connections
	for _, c := range mock.connections {
		c.Close()
	}
}

//=============================================================================
// Mock Server internals

// init starts listener on the provided address.
func (mock *ServerMock) init(addr string) error {
	mock.done = make(chan struct{})

	l, err := net.Listen("tcp", addr)
	if err != nil {
		mock.t.Errorf("TCPServerMock cannot listen on address: %q", addr)
		return err
	}
	mock.listener = l
	return nil
}

// loop accepts connections and creates a go routine per connection.
// The go routine is running the client handler, that is used to provide the
// real TCP server behaviour.
func (mock *ServerMock) loop() {
	listener := mock.listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-mock.done:
				return
			default:
				mock.t.Error("TCPServerMock accept error:", err.Error())
			}
			return
		}
		mock.connections = append(mock.connections, conn)
		// TODO Create and pass a context to cancel the handler if they are still around = avoid possible leak on complex handlers
		go mock.handler(mock.t, conn)
	}
}
