package main

/*
xmpp_chat_client is a demo client that connect on an XMPP server to chat with other members
Note that this example sends to a very specific user. User logic is not implemented here.
*/

import (
	"flag"
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
	"log"
	"strings"
)

const (
	infoFormat = "====== "
	// Default configuration
	defaultConfigFilePath = "./"

	configFileName = "config"
	configType     = "yaml"
	// Keys in config
	serverAddressKey = "full_address"
	clientJid        = "jid"
	clientPass       = "pass"
	configContactSep = ";"
)

var (
	CorrespChan = make(chan string, 1)
	textChan    = make(chan string, 5)
	killChan    = make(chan struct{}, 1)
)

type config struct {
	Server   map[string]string `mapstructure:"server"`
	Client   map[string]string `mapstructure:"client"`
	Contacts string            `string:"contact"`
}

func main() {
	// ============================================================
	// Parse the flag with the config directory path as argument
	flag.String("c", defaultConfigFilePath, "Provide a path to the directory that contains the configuration"+
		" file you want to use. Config file should be named \"config\" and be of YAML format..")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	// ==========================
	// Read configuration
	c := readConfig()

	// ==========================
	// Create TUI
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()
	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen
	g.SetManagerFunc(layout)
	setKeyBindings(g)

	// ==========================
	// Run TUI
	errChan := make(chan error)
	go func() {
		errChan <- g.MainLoop()
	}()

	// ==========================
	// Start XMPP client
	go startClient(g, c)

	select {
	case err := <-errChan:
		if err == gocui.ErrQuit {
			log.Println("Closing client.")
		} else {
			log.Panicln(err)
		}
	}
}

func startClient(g *gocui.Gui, config *config) {

	// ==========================
	// Client setup
	clientCfg := xmpp.Config{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: config.Server[serverAddressKey],
		},
		Jid:        config.Client[clientJid],
		Credential: xmpp.Password(config.Client[clientPass]),
		Insecure:   true}

	var client *xmpp.Client
	var err error
	router := xmpp.NewRouter()

	handlerWithGui := func(_ xmpp.Sender, p stanza.Packet) {
		msg, ok := p.(stanza.Message)
		v, err := g.View(chatLogWindow)
		if !ok {
			fmt.Fprintf(v, "%sIgnoring packet: %T\n", infoFormat, p)
			return
		}
		if err != nil {
			return
		}
		g.Update(func(g *gocui.Gui) error {
			if msg.Error.Code != 0 {
				_, err := fmt.Fprintf(v, "Error from server : %s : %s \n", msg.Error.Reason, msg.XMLName.Space)
				return err
			}
			_, err := fmt.Fprintf(v, "%s : %s \n", msg.From, msg.Body)
			return err
		})
	}

	router.HandleFunc("message", handlerWithGui)
	if client, err = xmpp.NewClient(clientCfg, router, errorHandler); err != nil {
		panic(fmt.Sprintf("Could not create a new client ! %s", err))

	}

	// ==========================
	// Client connection
	if err = client.Connect(); err != nil {
		msg := fmt.Sprintf("%sXMPP connection failed: %s", infoFormat, err)
		g.Update(func(g *gocui.Gui) error {
			v, err := g.View(chatLogWindow)
			fmt.Fprintf(v, msg)
			return err
		})
		return
	}

	// ==========================
	// Start working
	//askForRoster(client, g)
	updateRosterFromConfig(g, config)
	startMessaging(client, config)
}

func startMessaging(client xmpp.Sender, config *config) {
	var text string
	// Update this with a channel. Default value is the first contact in the list from the config.
	correspondent := strings.Split(config.Contacts, configContactSep)[0]
	for {
		select {
		case <-killChan:
			return
		case text = <-textChan:
			reply := stanza.Message{Attrs: stanza.Attrs{To: correspondent}, Body: text}
			err := client.Send(reply)
			if err != nil {
				fmt.Printf("There was a problem sending the message : %v", reply)
				return
			}
		case crrsp := <-CorrespChan:
			correspondent = crrsp
		}

	}
}

func readConfig() *config {
	viper.SetConfigName(configFileName) // name of config file (without extension)
	viper.BindPFlags(pflag.CommandLine)
	viper.AddConfigPath(viper.GetString("c")) // path to look for the config file in
	err := viper.ReadInConfig()               // Find and read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("%s %s", err, "Please make sure you give a path to the directory of the config and not to the config itself.")
		} else {
			log.Panicln(err)
		}
	}
	viper.SetConfigType(configType)
	var config config
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("Unable to decode Config: %s \n", err))
	}

	return &config
}

// If an error occurs, this is used to kill the client
func errorHandler(err error) {
	fmt.Printf("%v", err)
	killChan <- struct{}{}
}

// Read the client roster from the config. This does not check with the server that the roster is correct.
// If user tries to send a message to someone not registered with the server, the server will return an error.
func updateRosterFromConfig(g *gocui.Gui, config *config) {
	g.Update(func(g *gocui.Gui) error {
		menu, _ := g.View(menuWindow)
		for _, contact := range strings.Split(config.Contacts, configContactSep) {
			fmt.Fprintln(menu, contact)
		}
		return nil
	})
}

// Updates the menu panel of the view with the current user's roster.
// Need to add support for Roster IQ stanzas to make this work.
func askForRoster(client *xmpp.Client, g *gocui.Gui) {
	//ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	//iqReq := stanza.NewIQ(stanza.Attrs{Type: stanza.IQTypeGet, From: currentUserJid, To: "localhost", Lang: "en"})
	//disco := iqReq.DiscoInfo()
	//iqReq.Payload = disco
	//
	//// Handle a possible error
	//errChan := make(chan error)
	//errorHandler := func(err error) {
	//	errChan <- err
	//}
	//client.ErrorHandler = errorHandler
	//res, err := client.SendIQ(ctx, iqReq)
	//if err != nil {
	//	t.Errorf(err.Error())
	//}
	//
	//select {
	//case <-res:
	//}

	//roster := []string{"testuser1", "testuser2", "testuser3@localhost"}
	//
	//g.Update(func(g *gocui.Gui) error {
	//	menu, _ := g.View(menuWindow)
	//	for _, contact := range roster {
	//		fmt.Fprintln(menu, contact)
	//	}
	//	return nil
	//})
}
