package main

import (
	"github.com/bdlm/log"
	"github.com/spf13/cobra"
)

// cmdRoot represents the base command when called without any subcommands
var cmdRoot = &cobra.Command{
	Use:   "fluxxmpp",
	Short: "fluxxIO's xmpp comandline tool",
}

func main() {
	log.AddHook(&hook{})
	if err := cmdRoot.Execute(); err != nil {
		log.Fatal(err)
	}
}
