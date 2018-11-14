package main

import (
	"log"
	"os"
)

var (
	Version  string
	Revision string
)

func main() {
	err := run()
	if err != nil {
		log.Printf("alert: %s", err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

