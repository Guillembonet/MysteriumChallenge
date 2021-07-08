package main

import (
	"os"
)

func main() {
  serverPort := 12001
  clientPort := 11012
	relayPort := 11000
	cmd := os.Args[1]
	switch cmd {
		case "c":
			Client(serverPort, clientPort)
		case "s":
			Server(serverPort, clientPort)
		case "r":
			Relay(relayPort)
	}
}
