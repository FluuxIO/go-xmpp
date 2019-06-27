package stanza_test

import (
	"encoding/xml"

	"github.com/google/go-cmp/cmp"
)

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
