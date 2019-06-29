# XMPP Stanza

XMPP `stanza` package is used to parse, marshal and unmarshal XMPP stanzas and nonzas.

## Stanza creation

When creating stanzas, you can use two approaches:

1. You can create IQ, Presence or Message structs, set the fields and manually prepare extensions struct to add to the
stanza.
2. You can use `stanza` build helper to be guided when creating the stanza, and have more controls performed on the
final stanza.

The methods are equivalent and you can use whatever suits you best. The helpers will finally generate the same type of
struct that you can build by hand.

### Composing stanzas manually with structs

Here is for example how you would generate an IQ discovery result:

	iqResp := stanza.NewIQ(stanza.Attrs{Type: "result", From: iq.To, To: iq.From, Id: iq.Id})
	identity := stanza.Identity{
		Name:     opts.Name,
		Category: opts.Category,
		Type:     opts.Type,
	}
	payload := stanza.DiscoInfo{
		XMLName: xml.Name{
			Space: stanza.NSDiscoInfo,
			Local: "query",
		},
		Identity: []stanza.Identity{identity},
		Features: []stanza.Feature{
			{Var: stanza.NSDiscoInfo},
			{Var: stanza.NSDiscoItems},
			{Var: "jabber:iq:version"},
			{Var: "urn:xmpp:delegation:1"},
		},
	}
	iqResp.Payload = &payload

### Using helpers

Here is for example how you would generate an IQ discovery result using Builder:

	iq := stanza.NewIQ(stanza.Attrs{Type: "get", To: "service.localhost", Id: "disco-get-1"})
	disco := iq.DiscoInfo()
	disco.AddIdentity("Test Component", "gateway", "service")
	disco.AddFeatures(stanza.NSDiscoInfo, stanza.NSDiscoItems, "jabber:iq:version", "urn:xmpp:delegation:1")

## Payload and extensions

### Message

Here is the list of implemented message extensions:

- `Delegation`

- `Markable`
- `MarkAcknowledged`
- `MarkDisplayed`
- `MarkReceived`

- `StateActive`
- `StateComposing`
- `StateGone`
- `StateInactive`
- `StatePaused`

- `HTML`

- `OOB`

- `ReceiptReceived`
- `ReceiptRequest`

- `Mood`

### Presence

Here is the list of implemented presence extensions:

- `MucPresence`

### IQ

IQ (Information Queries) contain a payload associated with the request and possibly an error. The main difference with
Message and Presence extension is that you can only have one payload per IQ. The XMPP specification does not support
having multiple payloads.

Here is the list of structs implementing IQPayloads:

- `ControlSet`
- `ControlSetResponse`
- `Delegation`
- `DiscoInfo`
- `DiscoItems`
- `Pubsub`
- `Version`
- `Node`

Finally, when the payload of the parsed stanza is unknown, the parser will provide the unknown payload as a generic
`Node` element. You can also use the Node struct to add custom information on stanza generation. However, in both cases,
you may also consider [adding your own custom extensions on stanzas]().


## Adding your own custom extensions on stanzas

Extensions are registered on launch using the `Registry`. It can be used to register you own custom payload. You may
want to do so to support extensions we did not yet implement, or to add your own custom extensions to your XMPP stanzas.

To create an extension you need:
1. to create a struct for that extension. It need to have XMLName for consistency and to tagged at the struct level with
`xml` info.
2. It need to implement one or several extensions interface: stanza.IQPayload, stanza.MsgExtension and / or
stanza.PresExtension
3. Add that custom extension to the stanza.TypeRegistry during the file init.

Here an example code showing how to create a custom IQPayload. 

```go
package myclient

import (
	"encoding/xml"

	"gosrc.io/xmpp/stanza"
)

type CustomPayload struct {
	XMLName xml.Name `xml:"my:custom:payload query"`
	Node    string   `xml:"node,attr,omitempty"`
}

func (c CustomPayload) Namespace() string {
	return c.XMLName.Space
}

func init() {
	stanza.TypeRegistry.MapExtension(stanza.PKTIQ, xml.Name{"my:custom:payload", "query"}, CustomPayload{})
}
```