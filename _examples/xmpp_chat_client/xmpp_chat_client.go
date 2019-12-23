package main

/*
xmpp_chat_client is a demo client that connect on an XMPP server to chat with other members
*/

import (
	"context"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	infoFormat = "====== "
	// Default configuration
	defaultConfigFilePath = "./"

	configFileName = "config"
	configType     = "yaml"
	logStanzasOn   = "logger_on"
	logFilePath    = "logfile_path"
	// Keys in config
	serverAddressKey = "full_address"
	clientJid        = "jid"
	clientPass       = "pass"
	configContactSep = ";"
)

var (
	CorrespChan = make(chan string, 1)
	textChan    = make(chan string, 5)
	rawTextChan = make(chan string, 5)
	killChan    = make(chan error, 1)
	errChan     = make(chan error)
	rosterChan  = make(chan struct{})

	logger        *log.Logger
	disconnectErr = errors.New("disconnecting client")
)

type config struct {
	Server     map[string]string `mapstructure:"server"`
	Client     map[string]string `mapstructure:"client"`
	Contacts   string            `string:"contact"`
	LogStanzas map[string]string `mapstructure:"logstanzas"`
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

	//================================
	// Setup logger
	on, err := strconv.ParseBool(c.LogStanzas[logStanzasOn])
	if err != nil {
		log.Panicln(err)
	}
	if on {
		f, err := os.OpenFile(path.Join(c.LogStanzas[logFilePath], "logs.txt"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Panicln(err)
		}
		logger = log.New(f, "", log.Lshortfile|log.Ldate|log.Ltime)
		logger.SetOutput(f)
		defer f.Close()
	}

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
		if logger != nil {
			logger.Println(msg)
		}

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
			if len(strings.TrimSpace(msg.Body)) != 0 {
				_, err := fmt.Fprintf(v, "%s : %s \n", msg.From, msg.Body)
				return err
			}
			return nil
		})
	}

	router.HandleFunc("message", handlerWithGui)
	if client, err = xmpp.NewClient(clientCfg, router, errorHandler); err != nil {
		log.Panicln(fmt.Sprintf("Could not create a new client ! %s", err))

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
		fmt.Println("Failed to connect to server. Exiting...")
		errChan <- servConnFail
		return
	}

	// ==========================
	// Start working
	updateRosterFromConfig(g, config)
	// Sending the default contact in a channel. Default value is the first contact in the list from the config.
	viewState.currentContact = strings.Split(config.Contacts, configContactSep)[0]
	// Informing user of the default contact
	clw, _ := g.View(chatLogWindow)
	fmt.Fprintf(clw, infoFormat+"Now sending messages to "+viewState.currentContact+" in a private conversation\n")
	CorrespChan <- viewState.currentContact
	startMessaging(client, config, g)
}

func startMessaging(client xmpp.Sender, config *config, g *gocui.Gui) {
	var text string
	var correspondent string
	for {
		select {
		case err := <-killChan:
			if err == disconnectErr {
				sc := client.(xmpp.StreamClient)
				sc.Disconnect()
			} else {
				logger.Println(err)
			}
			return
		case text = <-textChan:
			reply := stanza.Message{Attrs: stanza.Attrs{To: correspondent, From: config.Client[clientJid], Type: stanza.MessageTypeChat}, Body: text}
			if logger != nil {
				raw, _ := xml.Marshal(reply)
				logger.Println(string(raw))
			}
			err := client.Send(reply)
			if err != nil {
				fmt.Printf("There was a problem sending the message : %v", reply)
				return
			}
		case text = <-rawTextChan:
			if logger != nil {
				logger.Println(text)
			}
			err := client.SendRaw(text)
			if err != nil {
				fmt.Printf("There was a problem sending the message : %v", text)
				return
			}
		case crrsp := <-CorrespChan:
			correspondent = crrsp
		case <-rosterChan:
			askForRoster(client, g, config)
		}

	}
}

// Only reads and parses the configuration
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

	// Check if we have contacts to message
	if len(strings.TrimSpace(config.Contacts)) == 0 {
		log.Panicln("You appear to have no contacts to message !")
	}
	// Check logging
	config.LogStanzas[logFilePath] = path.Clean(config.LogStanzas[logFilePath])
	on, err := strconv.ParseBool(config.LogStanzas[logStanzasOn])
	if err != nil {
		log.Panicln(err)
	}
	if d, e := isDirectory(config.LogStanzas[logFilePath]); (e != nil || !d) && on {
		log.Panicln("The log file path could not be found or is not a directory.")
	}

	return &config
}

// If an error occurs, this is used to kill the client
func errorHandler(err error) {
	killChan <- err
}

// Read the client roster from the config. This does not check with the server that the roster is correct.
// If user tries to send a message to someone not registered with the server, the server will return an error.
func updateRosterFromConfig(g *gocui.Gui, config *config) {
	viewState.contacts = append(strings.Split(config.Contacts, configContactSep), backFromContacts)
}

// Updates the menu panel of the view with the current user's roster, by asking the server.
func askForRoster(client xmpp.Sender, g *gocui.Gui, config *config) {
	// Craft a roster request
	req := stanza.NewIQ(stanza.Attrs{From: config.Client[clientJid], Type: stanza.IQTypeGet})
	req.RosterItems()
	if logger != nil {
		m, _ := xml.Marshal(req)
		logger.Println(string(m))
	}
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	// Send the roster request to the server
	c, err := client.SendIQ(ctx, req)
	if err != nil {
		logger.Panicln(err)
	}

	// Sending a IQ has a channel spawned to process the response once we receive it.
	// In order not to block the client, we spawn a goroutine to update the TUI once the server has responded.
	go func() {
		serverResp := <-c
		if logger != nil {
			m, _ := xml.Marshal(serverResp)
			logger.Println(string(m))
		}
		// Update contacts with the response from the server
		chlw, _ := g.View(chatLogWindow)
		if rosterItems, ok := serverResp.Payload.(*stanza.RosterItems); ok {
			viewState.contacts = []string{}
			for _, item := range rosterItems.Items {
				viewState.contacts = append(viewState.contacts, item.Jid)
			}
			viewState.contacts = append(viewState.contacts, backFromContacts)
			fmt.Fprintln(chlw, infoFormat+"Contacts list updated !")
			return
		}
		fmt.Fprintln(chlw, infoFormat+"Failed to update contact list !")
	}()
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), err
}
