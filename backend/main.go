package main

import (
	"log"
	"os"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatal("usage: go run . [server|client]")
	}

	switch os.Args[1] {
	case "server":
		if err := runServer(); err != nil {
			log.Fatal(err)
		}
	case "client":
		RunClient()
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}
