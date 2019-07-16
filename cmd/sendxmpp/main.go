package main

import (
	"github.com/bdlm/log"
)

func main() {
	log.AddHook(&hook{})
	cmd.Execute()
}
