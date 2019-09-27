package xmpp

import (
	"bytes"
	"testing"
)

func TestClient_Send(t *testing.T) {
	buffer := bytes.NewBufferString("")
	client := Client{}
	data := []byte("https://da.wikipedia.org/wiki/J%C3%A6vnd%C3%B8gn")
	if err := client.sendWithWriter(buffer, data); err != nil {
		t.Errorf("Writing failed: %v", err)
	}

	if buffer.String() != string(data) {
		t.Errorf("Incorrect value sent to buffer: '%s'", buffer.String())
	}
}
