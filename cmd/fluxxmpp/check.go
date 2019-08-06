package main

import (
	"github.com/bdlm/log"
	"github.com/spf13/cobra"
	"gosrc.io/xmpp"
)

var domain = ""
var cmdCheck = &cobra.Command{
	Use:     "check <host[:port]>",
	Short:   "is a command-line to check if you XMPP TLS certificate is valid and warn you before it expires",
	Example: "fluxxmpp check chat.sum7.eu:5222 --domain meckerspace.de",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runCheck(args[0], domain)
	},
}

func init() {
	cmdRoot.AddCommand(cmdCheck)
	cmdCheck.Flags().StringVarP(&domain, "domain", "d", "", "domain if host handle multiple domains")
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
