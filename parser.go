package xmpp // import "gosrc.io/xmpp"

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

// Reads and checks the opening XMPP stream element.
// TODO It returns a stream structure containing:
// - Host: You can check the host against the host you were expecting to connect to
// - Id: the Stream ID is a temporary shared secret used for some hash calculation. It is also used by ProcessOne
//       reattach features (allowing to resume an existing stream at the point the connection was interrupted, without
//       getting through the authentication process.
// TODO We should handle stream error from XEP-0114 ( <conflict/> or <host-unknown/> )
func initDecoder(p *xml.Decoder) (sessionID string, err error) {
	for {
		var t xml.Token
		t, err = p.Token()
		if err != nil {
			return
		}

		switch elem := t.(type) {
		case xml.StartElement:
			if elem.Name.Space != NSStream || elem.Name.Local != "stream" {
				err = errors.New("xmpp: expected <stream> but got <" + elem.Name.Local + "> in " + elem.Name.Space)
				return
			}

			// Parse Stream attributes
			for _, attrs := range elem.Attr {
				switch attrs.Name.Local {
				case "id":
					sessionID = attrs.Value
				}
			}
			return
		}
	}
}

// Scan XML token stream to find next StartElement.
func nextStart(p *xml.Decoder) (xml.StartElement, error) {
	for {
		t, err := p.Token()
		if err == io.EOF {
			return xml.StartElement{}, errors.New("connection closed")
		}
		if err != nil {
			return xml.StartElement{}, fmt.Errorf("nextStart %s", err)
		}
		switch t := t.(type) {
		case xml.StartElement:
			return t, nil
		}
	}
}

// next scans XML token stream for next element and then assign a structure to decode
// that elements.
// TODO Use an interface to return packets interface xmppDecoder
func next(p *xml.Decoder) (Packet, error) {
	// Read start element to find out how we want to parse the XMPP packet
	se, err := nextStart(p)
	if err != nil {
		return nil, err
	}

	// Decode one of the top level XMPP namespace
	switch se.Name.Space {
	case NSStream:
		return decodeStream(p, se)
	case nsSASL:
		return decodeSASL(p, se)
	case NSClient:
		return decodeClient(p, se)
	case NSComponent:
		return decodeComponent(p, se)
	default:
		return nil, errors.New("unknown namespace " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}
}

func decodeStream(p *xml.Decoder, se xml.StartElement) (Packet, error) {
	switch se.Name.Local {
	case "error":
		return streamError.decode(p, se)
	case "features":
		return streamFeatures.decode(p, se)
	default:
		return nil, errors.New("unexpected XMPP packet " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}
}

func decodeSASL(p *xml.Decoder, se xml.StartElement) (Packet, error) {
	switch se.Name.Local {
	case "success":
		return saslSuccess.decode(p, se)
	case "failure":
		return saslFailure.decode(p, se)
	default:
		return nil, errors.New("unexpected XMPP packet " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}
}

func decodeClient(p *xml.Decoder, se xml.StartElement) (Packet, error) {
	switch se.Name.Local {
	case "message":
		return message.decode(p, se)
	case "presence":
		return presence.decode(p, se)
	case "iq":
		return iq.decode(p, se)
	default:
		return nil, errors.New("unexpected XMPP packet " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}
}

func decodeComponent(p *xml.Decoder, se xml.StartElement) (Packet, error) {
	switch se.Name.Local {
	case "handshake":
		return handshake.decode(p, se)
	case "message":
		return message.decode(p, se)
	case "presence":
		return presence.decode(p, se)
	case "iq":
		return iq.decode(p, se)
	default:
		return nil, errors.New("unexpected XMPP packet " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}
}
