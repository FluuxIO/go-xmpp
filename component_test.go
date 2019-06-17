package xmpp // import "gosrc.io/xmpp"

import (
	"encoding/xml"
	"testing"
)

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

// We should be able to properly parse delegation confirmation messages
func TestParsingDelegationMessage(t *testing.T) {
	packetStr := `<message to='service.localhost' from='localhost'>
 <delegation xmlns='urn:xmpp:delegation:1'>
  <delegated namespace='http://jabber.org/protocol/pubsub'/>
 </delegation>
</message>`
	var msg Message
	data := []byte(packetStr)
	if err := xml.Unmarshal(data, &msg); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	// Check that we have extracted the delegation info as MsgExtension
	var nsDelegated string
	for _, ext := range msg.Extensions {
		if delegation, ok := ext.(*Delegation); ok {
			nsDelegated = delegation.Delegated.Namespace
		}
	}
	if nsDelegated != "http://jabber.org/protocol/pubsub" {
		t.Errorf("Could not find delegated namespace in delegation: %#v\n", msg)
	}
}

// Check that we can parse a delegation IQ.
// The most important thing is to be able to
func TestParsingDelegationIQ(t *testing.T) {
	packetStr := `<iq to='service.localhost' from='localhost' type='set' id='1'>
 <delegation xmlns='urn:xmpp:delegation:1'>
  <forwarded xmlns='urn:xmpp:forward:0'>
   <iq xml:lang='en' to='test1@localhost' from='test1@localhost/mremond-mbp' type='set' id='aaf3a' xmlns='jabber:client'>
    <pubsub xmlns='http://jabber.org/protocol/pubsub'>
     <publish node='http://jabber.org/protocol/mood'>
      <item id='current'>
       <mood xmlns='http://jabber.org/protocol/mood'>
        <excited/>
       </mood>
      </item>
     </publish>
    </pubsub>
   </iq>
  </forwarded>
 </delegation>
</iq>`
	var iq IQ
	data := []byte(packetStr)
	if err := xml.Unmarshal(data, &iq); err != nil {
		t.Errorf("Unmarshal(%s) returned error", data)
	}

	// Check that we have extracted the delegation info as IQPayload
	var node string
	for _, ext := range iq.Payload {
		if delegation, ok := ext.(*Delegation); ok {
			payload := delegation.Forwarded.IQ.Payload
			if len(payload) > 0 {
				payload := delegation.Forwarded.IQ.Payload[0]
				if pubsub, ok := payload.(*PubSub); ok {
					node = pubsub.Publish.Node
				}
			}
		}
	}
	if node != "http://jabber.org/protocol/mood" {
		t.Errorf("Could not find mood node name on delegated publish: %#v\n", iq)
	}
}
