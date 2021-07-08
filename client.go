package main

import (
	"fmt"
	"net"
	"strconv"
	"os"
	"strings"
)

func GetOutboundIP() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        fmt.Printf("Error when getting local IP")
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP
}

func requestServer(msgBuf []byte, conn *net.UDPConn, relay string, relayPort int, serverPort int) (string, error) {
	ownIP := GetOutboundIP().String()

  // Resolve the passed (relay) address as UDP4
	toAddr, err := net.ResolveUDPAddr("udp4", relay + ":" + strconv.Itoa(relayPort))
	if err != nil {
		fmt.Printf("Could not resolve %s:%d\n", relay, relayPort)
		return "", err
	}

	// Resolve the server address as UDP4
  fromAddr, err := net.ResolveUDPAddr("udp4", ownIP + ":" + strconv.Itoa(serverPort))
	if err != nil {
		fmt.Printf("Could not resolve %s\n", ownIP + ":" + strconv.Itoa(serverPort))
		return "", err
	}

	fmt.Printf("Trying to punch a hole to %s:%d\n", relay, relayPort)

	// Create connection to relay
	tmpConn, err := net.DialUDP("udp4", fromAddr, toAddr)
	*conn = *tmpConn
	if err != nil {
		fmt.Printf("Unable to connect to %s:%d\n", relay, relayPort)
		return "", err
	}

	// Register client to server "test"
	msg := "CLIENT test"
	fmt.Fprintf(conn, msg)
	fmt.Printf("Sent a UDP packet to %s:%d from %s\n\tSent: %s\n", relay, relayPort, fromAddr, msg)

	// Await relay registation ack
	rcvLen, addr, err := conn.ReadFrom(msgBuf)
	if err != nil {
		fmt.Println("Transaction was initiated but encountered an error!")
	}

	fmt.Printf("Received a packet from: %s\n\tResult: %s\n",
		addr.String(), msgBuf[:rcvLen])

	return string(msgBuf[:rcvLen]), nil
}

// Client --
func Client(clientPort int, relayPort int) {
	msgBuf := make([]byte, 1024)
	var conn net.UDPConn
	relay := os.Args[2]

	// Get server ip and port through relay and request hole punching
	server, err := requestServer(msgBuf, &conn, relay, relayPort, clientPort)
	defer conn.Close()
	if err != nil {
		fmt.Println("Error getting server")
	}

	// Await server hole punching
	rcvLen, addr, err := conn.ReadFrom(msgBuf)
	if err != nil {
		fmt.Println("Transaction was initiated but encountered an error!")
	}

	fmt.Printf("Received a packet from: %s\n\tResult: %s\n",
		addr.String(), msgBuf[:rcvLen])

	if !strings.HasPrefix(string(msgBuf[:rcvLen]), "PUNCHED ") { return }
	conn.Close()

	// Get own address
	ownIP := GetOutboundIP().String()
  fromAddr, err := net.ResolveUDPAddr("udp4", ownIP + ":" + strconv.Itoa(clientPort))
	if err != nil {
		fmt.Printf("Could not resolve %s\n", ownIP + ":" + strconv.Itoa(clientPort))
		return
	}

	// Resolve the received (server) address as UDP4
	toAddr, err := net.ResolveUDPAddr("udp4", server)
	if err != nil {
		fmt.Printf("Could not resolve %s\n", server)
		return
	}

	fmt.Printf("Trying to punch a hole to %s\n", server)

	// Initiate handshake
	ln, err := net.DialUDP("udp4", fromAddr, toAddr)
	defer ln.Close()

	if err != nil {
		fmt.Printf("Unable to connect to %s\n", toAddr)
		return
	}

	// Send message creating the hole
	msg := "Hello mr. Server"
	fmt.Fprintf(ln, msg)
	fmt.Printf("Sent a UDP packet to %s from %s\n\tSent: %s\n", toAddr, fromAddr, msg)

	connected := false
	for(!connected) {
		// Await a response through our NAT hole
		msgLen, originAddr, err := ln.ReadFromUDP(msgBuf)
		if err != nil {
			fmt.Printf("Error reading UDP response!\n")
			return
		}

		fmt.Printf("Received a UDP packet back from %s:%d\n\tResponse: %s\n",
			originAddr.IP, originAddr.Port, msgBuf[:msgLen])
		if (string(msgBuf[:msgLen]) == "Hello client!") {
			fmt.Println("Success: NAT traversed! ^-^")
			connected = true
		}
	}
}
