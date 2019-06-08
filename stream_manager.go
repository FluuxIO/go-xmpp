package xmpp // import "gosrc.io/xmpp"

import (
	"time"

	"golang.org/x/xerrors"
)

// The Fluux XMPP lib can manage client or component XMPP streams.
// The StreamManager handles the stream workflow handling the common
// stream events and doing the right operations.
//
// It can handle:
//     - Connection
//     - Stream establishment workflow
//     - Reconnection strategies, with exponential backoff. It also takes into account
//       permanent errors to avoid useless reconnection loops.
//     - Metrics processing

type StreamSession interface {
	Connect() error
	Disconnect()
	SetHandler(handler EventHandler)
}

// StreamManager supervises an XMPP client connection. Its role is to handle connection events and
// apply reconnection strategy.
type StreamManager struct {
	Client      *Client
	Session     *Session
	PostConnect PostConnect

	// Store low level metrics
	Metrics *Metrics
}

type PostConnect func(c *Client)

// NewStreamManager creates a new StreamManager structure, intended to support
// handling XMPP client state event changes and auto-trigger reconnection
// based on StreamManager configuration.
func NewStreamManager(client *Client, pc PostConnect) *StreamManager {
	return &StreamManager{
		Client:      client,
		PostConnect: pc,
	}
}

// Start launch the connection loop
func (cm *StreamManager) Start() error {
	cm.Client.Handler = func(e Event) {
		switch e.State {
		case StateConnected:
			cm.Metrics.setConnectTime()
		case StateSessionEstablished:
			cm.Metrics.setLoginTime()
		case StateDisconnected:
			// Reconnect on disconnection
			cm.connect()
		case StateStreamError:
			cm.Client.Disconnect()
			// Only try reconnecting if we have not been kicked by another session to avoid connection loop.
			if e.StreamError != "conflict" {
				cm.connect()
			}
		}
	}

	return cm.connect()
}

// Stop cancels pending operations and terminates existing XMPP client.
func (cm *StreamManager) Stop() {
	// Remove on disconnect handler to avoid triggering reconnect
	cm.Client.Handler = nil
	cm.Client.Disconnect()
}

// connect manages the reconnection loop and apply the define backoff to avoid overloading the server.
func (cm *StreamManager) connect() error {
	var backoff Backoff // TODO: Group backoff calculation features with connection manager?

	for {
		var err error
		// TODO: Make it possible to define logger to log disconnect and reconnection attempts
		cm.Metrics = initMetrics()

		if err = cm.Client.Connect(); err != nil {
			var actualErr ConnError
			if xerrors.As(err, &actualErr) {
				if actualErr.Permanent {
					return xerrors.Errorf("unrecoverable connect error %w", actualErr)
				}
			}
			backoff.Wait()
		} else { // We are connected, we can leave the retry loop
			break
		}
	}

	if cm.PostConnect != nil {
		cm.PostConnect(cm.Client)
	}
	return nil
}

// Stream Metrics
// ============================================================================

type Metrics struct {
	startTime time.Time
	// ConnectTime returns the duration between client initiation of the TCP/IP
	// connection to the server and actual TCP/IP session establishment.
	// This time includes DNS resolution and can be slightly higher if the DNS
	// resolution result was not in cache.
	ConnectTime time.Duration
	// LoginTime returns the between client initiation of the TCP/IP
	// connection to the server and the return of the login result.
	// This includes ConnectTime, but also XMPP level protocol negociation
	// like starttls.
	LoginTime time.Duration
}

// initMetrics set metrics with default value and define the starting point
// for duration calculation (connect time, login time, etc).
func initMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

func (m *Metrics) setConnectTime() {
	m.ConnectTime = time.Since(m.startTime)
}

func (m *Metrics) setLoginTime() {
	m.LoginTime = time.Since(m.startTime)
}
