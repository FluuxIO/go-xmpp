package xmpp

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	"gosrc.io/xmpp/stanza"
)

func authSASL(socket io.ReadWriter, decoder *xml.Decoder, f stanza.StreamFeatures, user string, password string) (err error) {
	// TODO: Implement other type of SASL Authentication
	havePlain := false
	for _, m := range f.Mechanisms.Mechanism {
		if m == "PLAIN" {
			havePlain = true
			break
		}
	}
	if !havePlain {
		err := fmt.Errorf("PLAIN authentication is not supported by server: %v", f.Mechanisms.Mechanism)
		return NewConnError(err, true)
	}

	return authPlain(socket, decoder, user, password)
}

// Plain authentication: send base64-encoded \x00 user \x00 password
func authPlain(socket io.ReadWriter, decoder *xml.Decoder, user string, password string) error {
	raw := "\x00" + user + "\x00" + password
	enc := make([]byte, base64.StdEncoding.EncodedLen(len(raw)))
	base64.StdEncoding.Encode(enc, []byte(raw))
	fmt.Fprintf(socket, "<auth xmlns='%s' mechanism='PLAIN'>%s</auth>", stanza.NSSASL, enc)

	// Next message should be either success or failure.
	val, err := stanza.NextPacket(decoder)
	if err != nil {
		return err
	}

	switch v := val.(type) {
	case stanza.SASLSuccess:
	case stanza.SASLFailure:
		// v.Any is type of sub-element in failure, which gives a description of what failed.
		err := errors.New("auth failure: " + v.Any.Local)
		return NewConnError(err, true)
	default:
		return errors.New("expected SASL success or failure, got " + v.Name())
	}
	return err
}
