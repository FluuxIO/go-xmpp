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
)

// Get the actual song Stream URL from SoundCloud website song URL and play it with mpg123 player.
const scClientID = "dde6a0075614ac4f3bea423863076b22"

func main() {
	jid := flag.String("jid", "", "jukebok XMPP JID, resource is optional")
	password := flag.String("password", "", "XMPP account password")
	address := flag.String("address", "", "If needed, XMPP server DNSName or IP and optional port (ie myserver:5222)")
	flag.Parse()

	var client *xmpp.Client
	var err error
	if client, err = connectXmpp(*jid, *password, *address); err != nil {
		log.Fatal("Could not connect to XMPP: ", err)
	}

	p, err := mpg123.NewPlayer()
	if err != nil {
		log.Fatal(err)
	}

	// Iterator to receive packets coming from our XMPP connection
	for packet := range client.Recv() {

		switch packet := packet.(type) {
		case xmpp.Message:
			processMessage(client, p, &packet)
		case xmpp.IQ:
			processIq(client, p, &packet)
		case xmpp.Presence:
			// Do nothing with received presence
		default:
			_, _ = fmt.Fprintf(os.Stdout, "Ignoring packet: %T\n", packet)
		}
	}
}

func processMessage(client *xmpp.Client, p *mpg123.Player, packet *xmpp.Message) {
	command := strings.Trim(packet.Body, " ")
	if command == "stop" {
		p.Stop()
	} else {
		playSCURL(p, command)
		sendUserTune(client, "Radiohead", "Spectre")
	}
}

func processIq(client *xmpp.Client, p *mpg123.Player, packet *xmpp.IQ) {
	switch payload := packet.Payload[0].(type) {
	// We support IOT Control IQ
	case *xmpp.ControlSet:
		var url string
		for _, element := range payload.Fields {
			if element.XMLName.Local == "string" && element.Name == "url" {
				url = strings.Trim(element.Value, " ")
				break
			}
		}

		playSCURL(p, url)
		setResponse := new(xmpp.ControlSetResponse)
		// FIXME: Broken
		reply := xmpp.IQ{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "result", Id: packet.Id}, Payload: []xmpp.IQPayload{setResponse}}
		_ = client.Send(reply)
		// TODO add Soundclound artist / title retrieval
		sendUserTune(client, "Radiohead", "Spectre")
	default:
		_, _ = fmt.Fprintf(os.Stdout, "Other IQ Payload: %T\n", packet.Payload)
	}
}

func sendUserTune(client *xmpp.Client, artist string, title string) {
	tune := xmpp.Tune{Artist: artist, Title: title}
	iq := xmpp.NewIQ("set", "", "", "usertune-1", "en")
	payload := xmpp.PubSub{Publish: &xmpp.Publish{Node: "http://jabber.org/protocol/tune", Item: xmpp.Item{Tune: &tune}}}
	iq.AddPayload(&payload)
	_ = client.Send(iq)
}

func playSCURL(p *mpg123.Player, rawURL string) {
	songID, _ := soundcloud.GetSongID(rawURL)
	// TODO: Maybe we need to check the track itself to get the stream URL from reply ?
	url := soundcloud.FormatStreamURL(songID)

	_ = p.Play(url)
}

func connectXmpp(jid string, password string, address string) (client *xmpp.Client, err error) {
	xmppConfig := xmpp.Config{Address: address,
		Jid: jid, Password: password, PacketLogger: os.Stdout, Insecure: true}

	if client, err = xmpp.NewClient(xmppConfig); err != nil {
		return
	}

	if err = client.Connect(); err != nil {
		return
	}
	return
}

// TODO
// - Have a player API to play, play next, or add to queue
// - Have the ability to parse custom packet to play sound
// - Use PEP to display tunes status
// - Ability to "speak" messages
