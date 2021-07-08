package main

import (
	"fmt"
	"net"
	"os"
  "strconv"
	"strings"
)

func registerServer(msgBuf []byte, conn *net.UDPConn, relay string, relayPort int, serverPort int) {
	ownIP := GetOutboundIP().String()

  // Resolve the passed (relay) address as UDP4
	toAddr, err := net.ResolveUDPAddr("udp4", relay + ":" + strconv.Itoa(relayPort))
	if err != nil {
		fmt.Printf("Could not resolve %s:%d\n", relay, relayPort)
		return
	}

	// Resolve the server address as UDP4
  fromAddr, err := net.ResolveUDPAddr("udp4", ownIP + ":" + strconv.Itoa(serverPort))
	if err != nil {
		fmt.Printf("Could not resolve %s\n", ownIP + ":" + strconv.Itoa(serverPort))
		return
	}

	fmt.Printf("Trying to punch a hole to %s:%d\n", relay, relayPort)

	// Initiate the transaction (force IPv4 to demo firewall punch)
	tmpConn, err := net.DialUDP("udp4", fromAddr, toAddr)
	*conn = *tmpConn
	if err != nil {
		fmt.Printf("Unable to connect to %s:%d\n", relay, relayPort)
		return
	}

	// Initiate the transaction, creating the hole
	msg := "SERVER test"
	fmt.Fprintf(conn, msg)
	fmt.Printf("Sent a UDP packet to %s:%d from %s\n\tSent: %s\n", relay, relayPort, fromAddr, msg)

	//await server registation
	rcvLen, addr, err := conn.ReadFrom(msgBuf)
	if err != nil {
		fmt.Println("Transaction was initiated but encountered an error!")
	}

	fmt.Printf("Received a packet from: %s\n\tRegistered: %s\n",
		addr.String(), msgBuf[:rcvLen])
}

// Server --
func Server(serverPort int, relayPort int) {
	msgBuf := make([]byte, 1024)
	var conn net.UDPConn

	relay := os.Args[2]

	//1. Register
	registerServer(msgBuf, &conn, relay, relayPort, serverPort)
	conn.Close()

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

		if (strings.HasPrefix(string(msgBuf[:rcvLen]), "CLIENT ")) {
			//resolve client
			clientAddr, err := net.ResolveUDPAddr("udp4", string(msgBuf[7:rcvLen]))
			if err != nil {
				fmt.Printf("Could not resolve %s\n", string(msgBuf[7:rcvLen]))
				return
			}

			//5. Punch hole
			reply := "PUNCHED " + string(msgBuf[7:rcvLen])
			copy(msgBuf, []byte(reply))
			_, err = ln.WriteTo(msgBuf[:len(reply)], clientAddr)
			//ack to relay
			copy(msgBuf, []byte(reply))
			_, err = ln.WriteTo(msgBuf[:len(reply)], clientAddr)
			fmt.Printf("Sent punch to client %s\n",
				clientAddr.String())

		} else {
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
}
