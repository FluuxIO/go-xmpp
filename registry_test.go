package xmpp // import "gosrc.io/xmpp"

import (
	"reflect"
	"testing"
)

func TestRegistry_RegisterMsgExt(t *testing.T) {
	// Setup registry
	typeRegistry := newRegistry()

	// Register an element
	req := ReceiptRequest{}
	res := ReceiptReceived{}

	typeRegistry.MapExtension(PKTMessage, req)
	typeRegistry.MapExtension(PKTMessage, res)

	// Match that element
	receipt := typeRegistry.GetMsgExtension(req.Name())
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
	req := ReceiptRequest{}
	typeRegistry.MapExtension(PKTMessage, req)

	for i := 0; i < b.N; i++ {
		// Match that element
		receipt := typeRegistry.GetExtensionType(PKTMessage, req.Name())
		if receipt == nil {
			b.Error("cannot read element type from registry")
			return
		}
	}
}
