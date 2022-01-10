package stanza

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
func InitStream(p *xml.Decoder) (sessionID string, err error) {
	for {
		var t xml.Token
		t, err = p.Token()
		if err != nil {
			return sessionID, err
		}

		switch elem := t.(type) {
		case xml.StartElement:
			isStreamOpen := elem.Name.Space == NSStream && elem.Name.Local == "stream"
			isFrameOpen := elem.Name.Space == NSFraming && elem.Name.Local == "open"
			if !isStreamOpen && !isFrameOpen {
				err = errors.New("xmpp: expected <stream> or <open> but got <" + elem.Name.Local + "> in " + elem.Name.Space)
				return sessionID, err
			}

			// Parse XMPP stream attributes
			for _, attrs := range elem.Attr {
				switch attrs.Name.Local {
				case "id":
					sessionID = attrs.Value
				}
			}
			return sessionID, err
		}
	}
}

// NextPacket scans XML token stream for next complete XMPP stanza.
// Once the type of stanza has been identified, a structure is created to decode
// that stanza and returned.
// TODO Use an interface to return packets interface xmppDecoder
// TODO make auth and bind use NextPacket instead of directly NextStart
func NextPacket(p *xml.Decoder) (Packet, error) {
	// Read start element to find out how we want to parse the XMPP packet
	t, err := NextXmppToken(p)
	if err != nil {
		return nil, err
	}

	if ee, ok := t.(xml.EndElement); ok {
		return decodeStream(p, ee)
	}

	// If not an end element, then must be a start
	se, ok := t.(xml.StartElement)
	if !ok {
		return nil, errors.New("unknown token ")
	}
	// Decode one of the top level XMPP namespace
	switch se.Name.Space {
	case NSStream:
		return decodeStream(p, se)
	case NSSASL:
		return decodeSASL(p, se)
	case NSClient:
		return decodeClient(p, se)
	case NSComponent:
		return decodeComponent(p, se)
	case NSStreamManagement:
		return sm.decode(p, se)
	default:
		return nil, errors.New("unknown namespace " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}
}

// NextXmppToken scans XML token stream to find next StartElement or stream EndElement.
// We need the EndElement scan, because we must register stream close tags
func NextXmppToken(p *xml.Decoder) (xml.Token, error) {
	for {
		t, err := p.Token()
		if err == io.EOF {
			return xml.StartElement{}, errors.New("connection closed")
		}
		if err != nil {
			return xml.StartElement{}, fmt.Errorf("NextStart: %w", err)
		}
		switch t := t.(type) {
		case xml.StartElement:
			return t, nil
		case xml.EndElement:
			if t.Name.Space == NSStream && t.Name.Local == "stream" {
				return t, nil
			}
		}
	}
}

// NextStart scans XML token stream to find next StartElement.
func NextStart(p *xml.Decoder) (xml.StartElement, error) {
	for {
		t, err := p.Token()
		if err == io.EOF {
			return xml.StartElement{}, errors.New("connection closed")
		}
		if err != nil {
			return xml.StartElement{}, fmt.Errorf("NextStart: %w", err)
		}
		switch t := t.(type) {
		case xml.StartElement:
			return t, nil
		}
	}
}

/*
TODO: From all the decoder, we can return a pointer to the actual concrete type, instead of directly that
   type.
   That way, we have a consistent way to do type assertion, always matching against pointers.
*/

// decodeStream will fully decode a stream packet
func decodeStream(p *xml.Decoder, t xml.Token) (Packet, error) {
	if se, ok := t.(xml.StartElement); ok {
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

	if ee, ok := t.(xml.EndElement); ok {
		if ee.Name.Local == "stream" {
			return streamClose.decode(ee), nil
		}
		return nil, errors.New("unexpected XMPP packet " +
			ee.Name.Space + " <" + ee.Name.Local + "/>")
	}

	// Should not happen
	return nil, errors.New("unexpected XML token ")
}

// decodeSASL decodes a packet related to SASL authentication.
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

// decodeClient decodes all known packets in the client namespace.
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

// decodeComponent decodes all known packets in the component namespace.
func decodeComponent(p *xml.Decoder, se xml.StartElement) (Packet, error) {
	switch se.Name.Local {
	case "handshake": // handshake is used to authenticate components
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
