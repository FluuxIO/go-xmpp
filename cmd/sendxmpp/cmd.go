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

var jid = ""
var password = ""

var receiverMUC = false
var stdIn = false

var cmd = &cobra.Command{
	Use:  "sendxmpp",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		receiver := strings.Split(args[0], ",")
		msgText := ""

		if !stdIn && len(args) < 2 {
			log.Error("no message to send")
			return
		} else if !stdIn {
			msgText = args[1]
		}

		var err error
		client, err := xmpp.NewClient(xmpp.Config{
			Jid:      jid,
			Password: password,
		}, xmpp.NewRouter())

		if err != nil {
			log.Panicf("error on startup xmpp client: %s", err)
		}

		wg := sync.WaitGroup{}
		wg.Add(1)

		cm := xmpp.NewStreamManager(client, func(c xmpp.Sender) {
			log.Info("client connected")
			if receiverMUC {
				for _, muc := range receiver {
					joinMUC(c, muc, "sendxmpp")
				}
			}

			if !stdIn {
				send(c, receiver, msgText)
				wg.Done()
				return
			}

			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				send(c, receiver, scanner.Text())
			}

			if err := scanner.Err(); err != nil {
				log.Errorf("error on reading stdin: %s", err)
			}
			wg.Done()
		})
		go func() {
			err := cm.Run()
			log.Panic("closed connection:", err)
		}()

		wg.Wait()

		leaveMUCs(client)
	},
}

func init() {
	cmd.Flags().StringVarP(&jid, "jid", "", "", "using jid (required)")
	viper.BindPFlag("jid", cmd.Flags().Lookup("jid"))
	// cmd.MarkFlagRequired("jid")

	cmd.Flags().StringVarP(&password, "password", "", "", "using password for your jid (required)")
	viper.BindPFlag("password", cmd.Flags().Lookup("password"))
	// cmd.MarkFlagRequired("password")

	cmd.Flags().BoolVarP(&stdIn, "stdin", "i", false, "read from stdin instatt of 2. argument")
	cmd.Flags().BoolVarP(&receiverMUC, "muc", "m", false, "reciever is a muc (join it before sending messages)")
}
