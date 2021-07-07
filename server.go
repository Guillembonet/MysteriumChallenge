package main

import (
	"fmt"
	"net"
)

// Server --
func Server(serverPort int) {
	msgBuf := make([]byte, 1024)

	// Initiatlize a UDP listener
	ln, err := net.ListenUDP("udp4", &net.UDPAddr{Port: serverPort})
	if err != nil {
		fmt.Printf("Unable to listen on :%d\n", serverPort)
		return
	}

	fmt.Printf("Listening on :%d\n", serverPort)

	for {
		fmt.Println("---")
		// Await incoming packets
		rcvLen, addr, err := ln.ReadFrom(msgBuf)
		if err != nil {
			fmt.Println("Transaction was initiated but encountered an error!")
			continue
		}

		fmt.Printf("Received a packet from: %s\n\tSays: %s\n",
			addr.String(), msgBuf[:rcvLen])

		// Let the client confirm a hole was punched through to us
		reply := "hole punched!"
		copy(msgBuf, []byte(reply))
		_, err = ln.WriteTo(msgBuf[:len(reply)], addr)

		if err != nil {
			fmt.Println("Socket closed unexpectedly!")
			continue
		}

		fmt.Printf("Sent reply to %s\n\tReply: %s\n",
			addr.String(), msgBuf[:len(reply)])
	}
}
