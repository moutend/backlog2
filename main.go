package main

import (
	"log"
	"os"
)

func main() {
	if err := rootCommand.Execute(); err != nil {
		log.New(os.Stderr, "", 0).Fatal(err)
	}
}
