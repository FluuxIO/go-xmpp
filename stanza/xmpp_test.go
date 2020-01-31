package stanza_test

import (
	"encoding/xml"
	"errors"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gosrc.io/xmpp/stanza"
)

var reLeadcloseWhtsp = regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
var reInsideWhtsp = regexp.MustCompile(`[\s\p{Zs}]`)

// ============================================================================
// Marshaller / unmarshaller test

func checkMarshalling(t *testing.T, iq *stanza.IQ) (*stanza.IQ, error) {
	// Marshall
	data, err := xml.Marshal(iq)
	if err != nil {
		t.Errorf("cannot marshal iq: %s\n%#v", err, iq)
		return nil, err
	}

	// Unmarshall
	var parsedIQ stanza.IQ
	err = xml.Unmarshal(data, &parsedIQ)
	if err != nil {
		t.Errorf("Unmarshal returned error: %s\n%s", err, data)
	}
	return &parsedIQ, err
}

// ============================================================================
// XML structs comparison

// Compare iq structure but ignore empty namespace as they are set properly on
// marshal / unmarshal. There is no need to manage them on the manually
// crafted structure.
func xmlEqual(x, y interface{}) bool {
	return cmp.Equal(x, y, xmlOpts())
}

// xmlDiff compares xml structures ignoring namespace preferences
func xmlDiff(x, y interface{}) string {
	return cmp.Diff(x, y, xmlOpts())
}

func xmlOpts() cmp.Options {
	alwaysEqual := cmp.Comparer(func(_, _ interface{}) bool { return true })
	opts := cmp.Options{
		cmp.FilterValues(func(x, y interface{}) bool {
			xx, xok := x.(xml.Name)
			yy, yok := y.(xml.Name)
			if xok && yok {
				zero := xml.Name{}
				if xx == zero || yy == zero {
					return true
				}
				if xx.Space == "" || yy.Space == "" {
					return true
				}
			}
			return false
		}, alwaysEqual),
	}
	return opts
}

func delSpaces(s string) string {
	return reInsideWhtsp.ReplaceAllString(reLeadcloseWhtsp.ReplaceAllString(s, ""), "")
}

func compareMarshal(expected, data string) error {
	if delSpaces(expected) != delSpaces(data) {
		return errors.New("failed to verify unmarshal->marshal. Expected :" + expected + "\ngot: " + data)
	}
	return nil
}
