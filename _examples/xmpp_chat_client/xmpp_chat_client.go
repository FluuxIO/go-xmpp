package main

/*
xmpp_chat_client is a demo client that connect on an XMPP server to chat with other members
Note that this example sends to a very specific user. User logic is not implemented here.
*/

import (
	. "bufio"
	"fmt"
	"os"

	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
)

const (
	currentUserAddress = "localhost:5222"
	currentUserJid     = "testuser@localhost"
	currentUserPass    = "testpass"
	correspondantJid   = "testuser2@localhost"
)

func main() {
	config := xmpp.Config{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: currentUserAddress,
		},
		Jid:        currentUserJid,
		Credential: xmpp.Password(currentUserPass),
		Insecure:   true}

	var client *xmpp.Client
	var err error
	router := xmpp.NewRouter()
	router.HandleFunc("message", handleMessage)
	if client, err = xmpp.NewClient(config, router, errorHandler); err != nil {
		fmt.Println("Error new client")
	}

	// Connecting client and handling messages
	// To use a stream manager, just write something like this instead :
	//cm := xmpp.NewStreamManager(client, startMessaging)
	//log.Fatal(cm.Run()) //=> this will lock the calling goroutine

	if err = client.Connect(); err != nil {
		fmt.Printf("XMPP connection failed: %s", err)
		return
	}
	startMessaging(client)

}

func startMessaging(client xmpp.Sender) {
	reader := NewReader(os.Stdin)
	textChan := make(chan string)
	var text string
	for {
		fmt.Print("Enter text: ")
		go readInput(reader, textChan)
		select {
		case <-killChan:
			return
		case text = <-textChan:
			reply := stanza.Message{Attrs: stanza.Attrs{To: correspondantJid}, Body: text}
			err := client.Send(reply)
			if err != nil {
				fmt.Printf("There was a problem sending the message : %v", reply)
				return
			}
		}
	}
}

func readInput(reader *Reader, textChan chan string) {
	text, _ := reader.ReadString('\n')
	textChan <- text
}

var killChan = make(chan struct{})

// If an error occurs, this is used
func errorHandler(err error) {
	fmt.Printf("%v", err)
	killChan <- struct{}{}
}

func handleMessage(s xmpp.Sender, p stanza.Packet) {
	msg, ok := p.(stanza.Message)
	if !ok {
		_, _ = fmt.Fprintf(os.Stdout, "Ignoring packet: %T\n", p)
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, "Body = %s - from = %s\n", msg.Body, msg.From)
}
