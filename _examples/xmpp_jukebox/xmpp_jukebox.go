// Can be launched with:
//   ./xmpp_jukebox -jid=test@localhost/jukebox -password=test -address=localhost:5222
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/processone/mpg123"
	"github.com/processone/soundcloud"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

// Get the actual song Stream URL from SoundCloud website song URL and play it with mpg123 player.
const scClientID = "dde6a0075614ac4f3bea423863076b22"

func main() {
	jid := flag.String("jid", "", "jukebok XMPP JID, resource is optional")
	password := flag.String("password", "", "XMPP account password")
	address := flag.String("address", "", "If needed, XMPP server DNSName or IP and optional port (ie myserver:5222)")
	flag.Parse()

	// 1. Create mpg player
	player, err := mpg123.NewPlayer()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Prepare XMPP client
	config := xmpp.Config{
		Address:  *address,
		Jid:      *jid,
		Password: *password,
		// PacketLogger: os.Stdout,
		Insecure: true,
	}

	router := xmpp.NewRouter()
	router.NewRoute().
		Packet("message").
		HandlerFunc(func(s xmpp.Sender, p stanza.Packet) {
			handleMessage(s, p, player)
		})
	router.NewRoute().
		Packet("message").
		HandlerFunc(func(s xmpp.Sender, p stanza.Packet) {
			handleIQ(s, p, player)
		})

	client, err := xmpp.NewClient(config, router)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	cm := xmpp.NewStreamManager(client, nil)
	log.Fatal(cm.Run())
}

func handleMessage(s xmpp.Sender, p stanza.Packet, player *mpg123.Player) {
	msg, ok := p.(stanza.Message)
	if !ok {
		return
	}
	command := strings.Trim(msg.Body, " ")
	if command == "stop" {
		player.Stop()
	} else {
		playSCURL(player, command)
		sendUserTune(s, "Radiohead", "Spectre")
	}
}

func handleIQ(s xmpp.Sender, p stanza.Packet, player *mpg123.Player) {
	iq, ok := p.(stanza.IQ)
	if !ok {
		return
	}
	switch payload := iq.Payload.(type) {
	// We support IOT Control IQ
	case *stanza.ControlSet:
		var url string
		for _, element := range payload.Fields {
			if element.XMLName.Local == "string" && element.Name == "url" {
				url = strings.Trim(element.Value, " ")
				break
			}
		}

		playSCURL(player, url)
		setResponse := new(stanza.ControlSetResponse)
		// FIXME: Broken
		reply := stanza.IQ{Attrs: stanza.Attrs{To: iq.From, Type: "result", Id: iq.Id}, Payload: setResponse}
		_ = s.Send(reply)
		// TODO add Soundclound artist / title retrieval
		sendUserTune(s, "Radiohead", "Spectre")
	default:
		_, _ = fmt.Fprintf(os.Stdout, "Other IQ Payload: %T\n", iq.Payload)
	}
}

func sendUserTune(s xmpp.Sender, artist string, title string) {
	tune := stanza.Tune{Artist: artist, Title: title}
	iq := stanza.NewIQ(stanza.Attrs{Type: "set", Id: "usertune-1", Lang: "en"})
	payload := stanza.PubSub{Publish: &stanza.Publish{Node: "http://jabber.org/protocol/tune", Item: stanza.Item{Tune: &tune}}}
	iq.Payload = &payload
	_ = s.Send(iq)
}

func playSCURL(p *mpg123.Player, rawURL string) {
	songID, _ := soundcloud.GetSongID(rawURL)
	// TODO: Maybe we need to check the track itself to get the stream URL from reply ?
	url := soundcloud.FormatStreamURL(songID)

	_ = p.Play(url)
}

// TODO
// - Have a player API to play, play next, or add to queue
// - Have the ability to parse custom packet to play sound
// - Use PEP to display tunes status
// - Ability to "speak" messages
