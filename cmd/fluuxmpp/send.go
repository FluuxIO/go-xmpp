package main

import (
	"bufio"
	"gosrc.io/xmpp/stanza"
	"os"
	"strings"
	"sync"

	"github.com/bdlm/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gosrc.io/xmpp"
)

var configFile = ""

// FIXME: Remove global variables
var isMUCRecipient = false

var cmdSend = &cobra.Command{
	Use:     "send <recipient,> [message]",
	Short:   "is a command-line tool to send to send XMPP messages to users",
	Example: `fluuxmpp send to@chat.sum7.eu "Hello World!"`,
	Args:    cobra.ExactArgs(2),
	Run:     sendxmpp,
}

func sendxmpp(cmd *cobra.Command, args []string) {
	receiver := strings.Split(args[0], ",")
	msgText := args[1]

	var err error
	client, err := xmpp.NewClient(&xmpp.Config{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: viper.GetString("addr"),
		},
		Jid:        viper.GetString("jid"),
		Credential: xmpp.Password(viper.GetString("password")),
	},
		xmpp.NewRouter(),
		func(err error) {
			log.Println(err)
		})

	if err != nil {
		log.Errorf("error when starting xmpp client: %s", err)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	// FIXME: Remove global variables
	var mucsToLeave []*stanza.Jid

	cm := xmpp.NewStreamManager(client, func(c xmpp.Sender) {
		defer wg.Done()

		log.Info("client connected")

		if isMUCRecipient {
			for _, muc := range receiver {
				jid, err := stanza.NewJid(muc)
				if err != nil {
					log.WithField("muc", muc).Errorf("skipping invalid muc jid: %w", err)
					continue
				}
				jid.Resource = "sendxmpp"

				if err := joinMUC(c, jid); err != nil {
					log.WithField("muc", muc).Errorf("error joining muc: %w", err)
					continue
				}
				mucsToLeave = append(mucsToLeave, jid)
			}
		}

		if msgText != "-" {
			send(c, receiver, msgText)
			return
		}

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			send(c, receiver, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Errorf("error on reading stdin: %s", err)
		}
	})
	go func() {
		err := cm.Run()
		log.Panic("closed connection:", err)
		wg.Done()
	}()

	wg.Wait()

	leaveMUCs(client, mucsToLeave)
}

func init() {
	cmdRoot.AddCommand(cmdSend)

	cobra.OnInitialize(initConfigFile)
	cmdSend.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is ~/.config/fluuxmpp.yml)")

	cmdSend.Flags().StringP("jid", "", "", "using jid (required)")
	viper.BindPFlag("jid", cmdSend.Flags().Lookup("jid"))

	cmdSend.Flags().StringP("password", "", "", "using password for your jid (required)")
	viper.BindPFlag("password", cmdSend.Flags().Lookup("password"))

	cmdSend.Flags().StringP("addr", "", "", "host[:port]")
	viper.BindPFlag("addr", cmdSend.Flags().Lookup("addr"))

	cmdSend.Flags().BoolVarP(&isMUCRecipient, "muc", "m", false, "recipient is a muc (join it before sending messages)")
}

// initConfig reads in config file and ENV variables if set.
func initConfigFile() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	}

	viper.SetConfigName("fluuxmpp")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("FLUXXMPP")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("no configuration found (somebody could read your password from process argument list): %s", err)
	}
}
