package main

import (
	"fmt"
	"net"
	"os"
  "strconv"
)

// Client --
func Client(serverPort int, clientPort int) {
  remote := os.Args[2]
  msgBuf := make([]byte, 1024)

  // Resolve the passed address as UDP4
	toAddr, err := net.ResolveUDPAddr("udp4", remote + ":" + strconv.Itoa(serverPort))
	if err != nil {
		fmt.Printf("Could not resolve %s:%d\n", remote, serverPort)
		return
	}

  fromAddr, err := net.ResolveUDPAddr("udp4", "192.168.42.116:11001")
	if err != nil {
		fmt.Printf("Could not resolve %s\n", "192.168.42.116:11001")
		return
	}

	fmt.Printf("Trying to punch a hole to %s:%d\n", remote, serverPort)

	// Initiate the transaction (force IPv4 to demo firewall punch)
	conn, err := net.DialUDP("udp4", fromAddr, toAddr)
  fmt.Printf("%s\n", conn.LocalAddr())
	defer conn.Close()

	if err != nil {
		fmt.Printf("Unable to connect to %s:%d\n", remote, serverPort)
		return
	}

	// Initiate the transaction, creating the hole
	msg := "trying..."
	fmt.Fprintf(conn, msg)
	fmt.Printf("Sent a UDP packet to %s:%d from %s\n\tSent: %s\n", remote, serverPort, fromAddr, msg)

	// Await a response through our firewall hole
	msgLen, fromAddr, err := conn.ReadFromUDP(msgBuf)
	if err != nil {
		fmt.Printf("Error reading UDP response!\n")
		return
	}

	fmt.Printf("Received a UDP packet back from %s:%d\n\tResponse: %s\n",
		fromAddr.IP, fromAddr.Port, msgBuf[:msgLen])

	fmt.Println("Success: NAT traversed! ^-^")
}
