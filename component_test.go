package xmpp

import (
	"fmt"
	"testing"
)

const testComponentDomain = "localhost"
const testComponentPort = "15222"

func TestHandshake(t *testing.T) {
	opts := ComponentOptions{
		Domain: "test.localhost",
		Secret: "mypass",
	}
	c := Component{ComponentOptions: opts}

	streamID := "1263952298440005243"
	expected := "c77e2ef0109fbbc5161e83b51629cd1353495332"

	result := c.handshake(streamID)
	if result != expected {
		t.Errorf("incorrect handshake calculation '%s' != '%s'", result, expected)
	}
}

func TestGenerateHandshake(t *testing.T) {
	// TODO
}

// Test that NewStreamManager can accept a Component.
//
// This validates that Component conforms to StreamClient interface.
func TestStreamManager(t *testing.T) {
	NewStreamManager(&Component{}, nil)
}

// Tests that the decoder is properly initialized when connecting a component to a server.
// The decoder is expected to be built after a valid connection
// Based on the xmpp_component example.
func TestDecoder(t *testing.T) {
	testComponentAddess := fmt.Sprintf("%s:%s", testComponentDomain, testComponentPort)
	mock := ServerMock{}
	mock.Start(t, testComponentAddess, handlerConnectSuccess)

	opts := ComponentOptions{
		TransportConfiguration: TransportConfiguration{
			Address: testComponentAddess,
			Domain:  "localhost",
		},
		Domain:   testComponentDomain,
		Secret:   "mypass",
		Name:     "Test Component",
		Category: "gateway",
		Type:     "service",
	}
	router := NewRouter()
	c, err := NewComponent(opts, router)
	if err != nil {
		t.Errorf("%+v", err)
	}
	c.transport, err = NewComponentTransport(c.ComponentOptions.TransportConfiguration)
	if err != nil {
		t.Errorf("%+v", err)
	}
	_, err = c.transport.Connect()
	if err != nil {
		t.Errorf("%+v", err)
	}
	if c.transport.GetDecoder() == nil {
		t.Errorf("Failed to initialize decoder. Decoder is nil.")
	}

}
