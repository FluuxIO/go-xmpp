package xmpp

import (
	"testing"
)

type params struct {
}

func TestParseAddr(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "ipv4-no-port-1", input: "localhost", want: "localhost:5222"},
		{name: "ipv4-with-port-1", input: "localhost:5555", want: "localhost:5555"},
		{name: "ipv4-no-port-2", input: "127.0.0.1", want: "127.0.0.1:5222"},
		{name: "ipv4-with-port-2", input: "127.0.0.1:5555", want: "127.0.0.1:5555"},
		{name: "ipv6-no-port-1", input: "::1", want: "[::1]:5222"},
		{name: "ipv6-no-port-2", input: "[::1]", want: "[::1]:5222"},
		{name: "ipv6-no-port-3", input: "2001::7334", want: "[2001::7334]:5222"},
		{name: "ipv6-no-port-4", input: "2001:db8:85a3:0:0:8a2e:370:7334", want: "[2001:db8:85a3:0:0:8a2e:370:7334]:5222"},
		{name: "ipv6-with-port-1", input: "[::1]:5555", want: "[::1]:5555"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(st *testing.T) {
			addr := ensurePort(tc.input, 5222)

			if addr != tc.want {
				st.Errorf("incorrect Result: %v (!= %v)", addr, tc.want)
			}
		})
	}
}
