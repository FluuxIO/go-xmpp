package xmpp

import (
	"strings"
	"testing"
)

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

func TestEnsurePort(t *testing.T) {
	testAddresses := []string{
		"1ca3:6c07:ee3a:89ca:e065:9a70:71d:daad",
		"1ca3:6c07:ee3a:89ca:e065:9a70:71d:daad:5252",
		"[::1]",
		"127.0.0.1:5555",
		"127.0.0.1",
		"[::1]:5555",
	}

	for _, oldAddr := range testAddresses {
		t.Run(oldAddr, func(st *testing.T) {
			newAddr := ensurePort(oldAddr, 5222)

			if len(newAddr) < len(oldAddr) {
				st.Errorf("incorrect Result: transformed address is shorter than input : %v (old) > %v (new)", newAddr, oldAddr)
			}
			// If IPv6, the new address needs brackets to specify a port, like so : [2001:db8:85a3:0:0:8a2e:370:7334]:5222
			if strings.Count(newAddr, "[") < strings.Count(oldAddr, "[") ||
				strings.Count(newAddr, "]") < strings.Count(oldAddr, "]") {

				st.Errorf("incorrect Result. Transformed address seems to not have correct brakets : %v => %v", oldAddr, newAddr)
			}

			// Check if we messed up the colons, or didn't properly add a port
			if strings.Count(newAddr, ":") < strings.Count(oldAddr, ":") {
				st.Errorf("incorrect Result: transformed address doesn't seem to have a port %v (=> %v, no port ?)", oldAddr, newAddr)
			}
		})
	}

}
