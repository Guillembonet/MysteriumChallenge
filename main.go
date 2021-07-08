package main

import (
	"os"
)

func main() {
  serverPort := 12001
  clientPort := 11011
	cmd := os.Args[1]
	switch cmd {
	case "c":
		Client(serverPort, clientPort)
	case "s":
		Server(serverPort)
	}
}
