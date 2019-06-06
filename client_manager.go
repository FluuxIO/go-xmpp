package xmpp // import "gosrc.io/xmpp"

import (
	"log"
	"time"
)

type PostConnect func(c *Client)

// ClientManager supervises an XMPP client connection. Its role is to handle connection events and
// apply reconnection strategy.
type ClientManager struct {
	Client      *Client
	Session     *Session
	PostConnect PostConnect

	// Store low level metrics
	Metrics *Metrics
}

// NewClientManager creates a new client manager structure, intended to support
// handling XMPP client state event changes and auto-trigger reconnection
// based on ClientManager configuration.
func NewClientManager(client *Client, pc PostConnect) *ClientManager {
	return &ClientManager{
		Client:      client,
		PostConnect: pc,
	}
}

// Start launch the connection loop
func (cm *ClientManager) Start() {
	cm.Client.Handler = func(e Event) {
		switch e.State {
		case StateConnected:
			cm.Metrics.setConnectTime()
		case StateSessionEstablished:
			cm.Metrics.setLoginTime()
		case StateDisconnected:
			// Reconnect on disconnection
			cm.connect()
		}
	}
	cm.connect()
}

// Stop cancels pending operations and terminates existing XMPP client.
func (cm *ClientManager) Stop() {
	// Remove on disconnect handler to avoid triggering reconnect
	cm.Client.Handler = nil
	cm.Client.Disconnect()
}

// connect manages the reconnection loop and apply the define backoff to avoid overloading the server.
func (cm *ClientManager) connect() {
	var backoff Backoff // TODO: Group backoff calculation features with connection manager?

	for {
		var err error
		cm.Metrics = initMetrics()

		if cm.Client.Session, err = cm.Client.Connect(); err != nil {
			log.Printf("Connection error: %v\n", err)
			backoff.Wait()
		} else {
			break
		}
	}

	if cm.PostConnect != nil {
		cm.PostConnect(cm.Client)
	}
}

// Client Metrics
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
