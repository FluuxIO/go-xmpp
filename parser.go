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

// Scan XML token stream for next element and save into val.
// If val == nil, allocate new element based on proto map.
// Either way, return val.
func next(p *xml.Decoder) (xml.Name, interface{}, error) {
	// Read start element to find out what type we want.
	se, err := nextStart(p)
	if err != nil {
		return xml.Name{}, nil, err
	}

	// Put it in an interface and allocate one.
	var nv interface{}
	switch se.Name.Space + " " + se.Name.Local {
	case NSStream + " error":
		nv = &StreamError{}
		// TODO: general case = Parse IQ / presence / message => split SASL case
	case nsSASL + " success":
		nv = &saslSuccess{}
	case nsSASL + " failure":
		nv = &saslFailure{}
	case NSClient + " message":
		nv = &ClientMessage{}
	case NSClient + " presence":
		nv = &ClientPresence{}
	case NSClient + " iq":
		nv = &ClientIQ{}
	default:
		return xml.Name{}, nil, errors.New("unexpected XMPP message " +
			se.Name.Space + " <" + se.Name.Local + "/>")
	}

	// Decode element into pointer storage
	if err = p.DecodeElement(nv, &se); err != nil {
		return xml.Name{}, nil, err
	}
	return se.Name, nv, err
}
