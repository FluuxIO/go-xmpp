package main

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/bdlm/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gosrc.io/xmpp"
)

var configFile = ""

var jid = ""
var password = ""

var receiverMUC = false

var cmd = &cobra.Command{
	Use:     "sendxmpp <recieve,> [message]",
	Example: `sendxmpp to@chat.sum7.eu "Hello World!"`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		receiver := strings.Split(args[0], ",")
		msgText := args[1]

		var err error
		client, err := xmpp.NewClient(xmpp.Config{
			Jid:      viper.GetString("jid"),
			Address:  viper.GetString("addr"),
			Password: viper.GetString("password"),
		}, xmpp.NewRouter())

		if err != nil {
			log.Errorf("error on startup xmpp client: %s", err)
			return
		}

		wg := sync.WaitGroup{}
		wg.Add(1)

		cm := xmpp.NewStreamManager(client, func(c xmpp.Sender) {
			defer wg.Done()

			log.Info("client connected")

			if receiverMUC {
				for _, muc := range receiver {
					joinMUC(c, muc, "sendxmpp")
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

		leaveMUCs(client)
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	cmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is ~/.config/fluxxmpp.yml)")

	cmd.Flags().StringP("jid", "", "", "using jid (required)")
	viper.BindPFlag("jid", cmd.Flags().Lookup("jid"))

	cmd.Flags().StringP("password", "", "", "using password for your jid (required)")
	viper.BindPFlag("password", cmd.Flags().Lookup("password"))

	cmd.Flags().StringP("addr", "", "", "host[:port]")
	viper.BindPFlag("addr", cmd.Flags().Lookup("addr"))

	cmd.Flags().BoolVarP(&receiverMUC, "muc", "m", false, "reciever is a muc (join it before sending messages)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	}

	viper.SetConfigName("fluxxmpp")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("FLUXXMPP")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("no configuration found (somebody could read your password from progress argument list): %s", err)
	}
}
