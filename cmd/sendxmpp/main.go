package main

import (
	"github.com/bdlm/log"
)

func main() {
	log.AddHook(&hook{})
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
