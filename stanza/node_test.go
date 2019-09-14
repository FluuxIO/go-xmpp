package stanza_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"gosrc.io/xmpp/stanza"
)

func TestNodeMarshalling(t *testing.T) {
	node := stanza.Node{
		XMLName:       xml.Name{Space: "go-xmpp", Local: "data"},
		CharacterData: "character data",
	}

	xmlData, err := xml.Marshal(node)
	if err != nil {
		t.Error("Can not marshal Node with character data")
		return
	}

	if !strings.Contains(string(xmlData), "character data") {
		t.Errorf("Character data not output: %v", string(xmlData))
		return
	}

	decodedNode := stanza.Node{}
	err = xml.Unmarshal(xmlData, &decodedNode)
	if decodedNode.CharacterData != "character data" {
		t.Error("Unmarshalling breaks character data")
	}
}
