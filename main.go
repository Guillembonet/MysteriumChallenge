package main

import (
	"os"
)

func main() {
  localPort := 9595
	cmd := os.Args[1]
	switch cmd {
	case "c":
		Client(localPort)
	case "s":
		Server(localPort)
	}
}
