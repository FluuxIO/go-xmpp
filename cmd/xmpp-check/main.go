package main

import (
	"github.com/bdlm/log"
	"github.com/spf13/cobra"
	"gosrc.io/xmpp"
)

func main() {
	log.AddHook(&hook{})
	cmd.Execute()
}

var domain = ""
var cmd = &cobra.Command{
	Use:     "xmpp-check <host[:port]>",
	Example: "xmpp-check chat.sum7.eu:5222 --domain meckerspace.de",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runCheck(args[0], domain)
	},
}

func init() {
	cmd.Flags().StringVarP(&domain, "domain", "d", "", "domain if host handle multiple domains")
}

func runCheck(address, domain string) {
	logger := log.WithFields(map[string]interface{}{
		"address": address,
		"domain":  domain,
	})
	client, err := xmpp.NewChecker(address, domain)

	if err != nil {
		log.Fatal("Error: ", err)
	}

	if err = client.Check(); err != nil {
		logger.Fatal("Failed connection check: ", err)
	}

	logger.Println("All checks passed")
}
