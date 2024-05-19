package xmpp

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/xerrors"
	"gosrc.io/xmpp/stanza"
)

// The Fluux XMPP lib can manage client or component XMPP streams.
// The StreamManager handles the stream workflow handling the common
// stream events and doing the right operations.
//
// It can handle:
//     - Client
//     - Stream establishment workflow
//     - Reconnection strategies, with exponential backoff. It also takes into account
//       permanent errors to avoid useless reconnection loops.
//     - Metrics processing

// StreamClient is an interface used by StreamManager to control Client lifecycle,
// set callback and trigger reconnection.
type StreamClient interface {
	Connect() error
	Resume() error
	Send(packet stanza.Packet) error
	SendIQ(ctx context.Context, iq *stanza.IQ) (chan stanza.IQ, error)
	SendRaw(packet string) error
	Disconnect() error
	SetHandler(handler EventHandler)
}

// Sender is an interface provided by Stream clients to allow sending XMPP data.
// It is mostly use in callback to pass a limited subset of the stream client interface
type Sender interface {
	Send(packet stanza.Packet) error
	SendIQ(ctx context.Context, iq *stanza.IQ) (chan stanza.IQ, error)
	SendRaw(packet string) error
}

// StreamManager supervises an XMPP client connection. Its role is to handle connection events and
// apply reconnection strategy.
type StreamManager struct {
	client      StreamClient
	PostConnect PostConnect

	// Store low level metrics
	Metrics *Metrics

	wg sync.WaitGroup
}

type PostConnect func(c Sender)

// NewStreamManager creates a new StreamManager structure, intended to support
// handling XMPP client state event changes and auto-trigger reconnection
// based on StreamManager configuration.
// TODO: Move parameters to Start and remove factory method
func NewStreamManager(client StreamClient, pc PostConnect) *StreamManager {
	return &StreamManager{
		client:      client,
		PostConnect: pc,
	}
}

// Run launches the connection of the underlying client or component
// and wait until Disconnect is called, or for the manager to terminate due
// to an unrecoverable error.
func (sm *StreamManager) Run() error {
	if sm.client == nil {
		return errors.New("missing stream client")
	}

	handler := func(e Event) error {
		switch e.State.state {
		case StateSessionEstablished:
			sm.Metrics.setLoginTime()
		case StateDisconnected:
			// Reconnect on disconnection
			return sm.resume()
		case StateStreamError:
			sm.client.Disconnect()
			// Only try reconnecting if we have not been kicked by another session to avoid connection loop.
			// TODO: Make this conflict exception a permanent error
			if e.StreamError != "conflict" {
				return sm.resume()
			}
		case StatePermanentError:
			// Do not attempt to reconnect
		}
		return nil
	}
	sm.client.SetHandler(handler)

	sm.wg.Add(1)
	if err := sm.connect(); err != nil {
		sm.wg.Done()
		return err
	}
	sm.wg.Wait()
	return nil
}

// Stop cancels pending operations and terminates existing XMPP client.
func (sm *StreamManager) Stop() {
	// Remove on disconnect handler to avoid triggering reconnect
	sm.client.SetHandler(nil)
	sm.client.Disconnect()
	sm.wg.Done()
}

func (sm *StreamManager) connect() error {
	if sm.client == nil {
		return errors.New("client is not set")
	}
	
	if c, ok := sm.client.(*Client); ok {
		if c.CurrentState.getState() != StateDisconnected {
			return errors.New("client is not disconnected")
		}
	}

	sm.Metrics = initMetrics()
	
	if err := sm.client.Connect(); err != nil {
		return err
	}
	
	if sm.PostConnect != nil {
		sm.PostConnect(sm.client)
	}
	
	return nil
}

// resume manages the reconnection loop and apply the define backoff to avoid overloading the server.
func (sm *StreamManager) resume() error {
	var backoff backoff // TODO: Group backoff calculation features with connection manager?

	for {
		var err error
		// TODO: Make it possible to define logger to log disconnect and reconnection attempts
		sm.Metrics = initMetrics()
		if err = sm.client.Resume(); err != nil {
			var actualErr ConnError
			if xerrors.As(err, &actualErr) {
				if actualErr.Permanent {
					return xerrors.Errorf("unrecoverable connect error %#v", actualErr)
				}
			}
			backoff.wait()
		} else { // We are connected, we can leave the retry loop
			break
		}
	}

	if sm.PostConnect != nil {
		sm.PostConnect(sm.client)
	}
	return nil
}

// Stream Metrics
// ============================================================================

type Metrics struct {
	startTime time.Time
	// LoginTime returns the between client initiation of the TCP/IP
	// connection to the server and the return of the login result.
	// This includes ConnectTime, but also XMPP level protocol negotiation
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

func (m *Metrics) setLoginTime() {
	m.LoginTime = time.Since(m.startTime)
}
