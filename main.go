package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"boundless-cli/cmd"
)

func main() {
	err := cmd.App.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
