package main

import (
	"fmt"
	"net"
)

// Exists takes a slice and looks for an element in it. If found it will
// return true else it will return false.
func Exists(slice []net.Addr, val net.Addr) bool {
	for _, item := range slice {
		if item.String() == val.String() {
			return true
		}
	}
	return false
}

//Send a message containing *message* to *address*
func sendMessage(msgBuf []byte, ln *net.UDPConn, message string, address net.Addr) {

	copy(msgBuf, []byte(message))
	_, err := ln.WriteTo(msgBuf[:len(message)], address)

	if err != nil {
		fmt.Println("Socket closed unexpectedly!")
	}

	fmt.Printf("Sent reply to %s\n\tReply: %s\n",
		address.String(), msgBuf[:len(message)])
}
