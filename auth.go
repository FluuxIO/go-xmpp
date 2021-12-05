package xmpp

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	"gosrc.io/xmpp/stanza"
)

// Credential is used to pass the type of secret that will be used to connect to XMPP server.
// It can be either a password or an OAuth 2 bearer token.
type Credential struct {
	secret     string
	mechanisms []string
}

func Password(pwd string) Credential {
	credential := Credential{
		secret:     pwd,
		mechanisms: []string{"PLAIN"},
	}
	return credential
}

func OAuthToken(token string) Credential {
	credential := Credential{
		secret:     token,
		mechanisms: []string{"X-OAUTH2"},
	}
	return credential
}

// ============================================================================
// Authentication flow for SASL mechanisms

func authSASL(socket io.ReadWriter, decoder *xml.Decoder, f stanza.StreamFeatures, user string, credential Credential) (err error) {
	var matchingMech string
	for _, mech := range credential.mechanisms {
		if isSupportedMech(mech, f.Mechanisms.Mechanism) {
			matchingMech = mech
			break
		}
	}

	switch matchingMech {
	case "PLAIN", "X-OAUTH2":
		// TODO: Implement other type of SASL mechanisms
		return authPlain(socket, decoder, matchingMech, user, credential.secret)
	default:
		err := fmt.Errorf("no matching authentication (%v) supported by server: %v", credential.mechanisms, f.Mechanisms.Mechanism)
		return NewConnError(err, true)
	}
}

// Plain authentication: send base64-encoded \x00 user \x00 password
func authPlain(socket io.ReadWriter, decoder *xml.Decoder, mech string, user string, secret string) error {
	raw := "\x00" + user + "\x00" + secret
	enc := make([]byte, base64.StdEncoding.EncodedLen(len(raw)))
	base64.StdEncoding.Encode(enc, []byte(raw))

	a := stanza.SASLAuth{
		Mechanism: mech,
		Value:     string(enc),
	}
	data, err := xml.Marshal(a)
	if err != nil {
		return err
	}
	n, err := socket.Write(data)
	if err != nil {
		return err
	} else if n == 0 {
		return errors.New("failed to write authSASL nonza to socket : wrote 0 bytes")
	}

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

// isSupportedMech returns true if the mechanism is supported in the provided list.
func isSupportedMech(mech string, mechanisms []string) bool {
	for _, m := range mechanisms {
		if mech == m {
			return true
		}
	}
	return false
}
