package xmpp // import "fluux.io/xmpp"

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
	panic("unreachable")
}

// Scan XML token stream to find next StartElement.
func nextStart(p *xml.Decoder) (xml.StartElement, error) {
	for {
		t, err := p.Token()
		if err == io.EOF {
			return xml.StartElement{}, nil
		}
		if err != nil {
			return xml.StartElement{}, fmt.Errorf("nextStart %s", err)
		}
		switch t := t.(type) {
		case xml.StartElement:
			return t, nil
		}
	}
	panic("unreachable")
}

// next scans XML token stream for next element and then assign a structure to decode
// that elements.
// TODO Use an interface to return packets interface xmppDecoder
func next(p *xml.Decoder) (xml.Name, interface{}, error) {
	// Read start element to find out what type we want.
	se, err := nextStart(p)
	if err != nil {
		return xml.Name{}, nil, err
	}

	// Put it in an interface and allocate the right structure
	var nv interface{}
	// TODO: general case = Parse IQ / presence / message => split SASL Stream and component cases
	switch se.Name.Space {
	case NSStream:
		if nv, err = decodeStream(se); err != nil {
			return xml.Name{}, nil, err
		}
	case nsSASL:
		if nv, err = decodeSASL(se); err != nil {
			return xml.Name{}, nil, err
		}
	case NSClient:
		if nv, err = decodeClient(se); err != nil {
			return xml.Name{}, nil, err
		}
	case NSComponent:
		if nv, err = decodeComponent(se); err != nil {
			return xml.Name{}, nil, err
		}
	default:
		return xml.Name{}, nil, errors.New("unknown namespace " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}

	// Decode element into pointer storage
	if err = p.DecodeElement(nv, &se); err != nil {
		return xml.Name{}, nil, err
	}
	return se.Name, nv, err
}

func decodeStream(se xml.StartElement) (interface{}, error) {
	switch se.Name.Local {
	case "error":
		return &StreamError{}, nil
	default:
		return nil, errors.New("unexpected XMPP packet " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}
}

func decodeSASL(se xml.StartElement) (interface{}, error) {
	switch se.Name.Local {
	case "success":
		return &saslSuccess{}, nil
	case "failure":
		return &saslFailure{}, nil
	default:
		return nil, errors.New("unexpected XMPP packet " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}
}

func decodeClient(se xml.StartElement) (interface{}, error) {
	switch se.Name.Local {
	case "message":
		return &ClientMessage{}, nil
	case "presence":
		return &ClientPresence{}, nil
	case "iq":
		return &ClientIQ{}, nil
	default:
		return nil, errors.New("unexpected XMPP packet " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}
}

func decodeComponent(se xml.StartElement) (interface{}, error) {
	switch se.Name.Local {
	case "handshake":
		return &Handshake{}, nil
	case "iq":
		return &ClientIQ{}, nil
	default:
		return nil, errors.New("unexpected XMPP packet " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}
}
