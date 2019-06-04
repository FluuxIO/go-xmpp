package xmpp // import "gosrc.io/xmpp"

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestRegistry_RegisterMsgExt(t *testing.T) {
	// Setup registry
	typeRegistry := newRegistry()

	// Register an element
	name := xml.Name{Space: "urn:xmpp:receipts", Local: "request"}
	typeRegistry.MapExtension(PKTMessage, name, ReceiptRequest{})

	// Match that element
	receipt := typeRegistry.GetMsgExtension(name)
	if receipt == nil {
		t.Error("cannot read element type from registry")
		return
	}

	switch r := receipt.(type) {
	case *ReceiptRequest:
	default:
		t.Errorf("Registry did not return expected type ReceiptRequest: %v", reflect.TypeOf(r))
	}
}

func BenchmarkRegistryGet(b *testing.B) {
	// Setup registry
	typeRegistry := newRegistry()

	// Register an element
	name := xml.Name{Space: "urn:xmpp:receipts", Local: "request"}
	typeRegistry.MapExtension(PKTMessage, name, ReceiptRequest{})

	for i := 0; i < b.N; i++ {
		// Match that element
		receipt := typeRegistry.GetExtensionType(PKTMessage, name)
		if receipt == nil {
			b.Error("cannot read element type from registry")
			return
		}
	}
}
