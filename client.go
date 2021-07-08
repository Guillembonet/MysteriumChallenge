package main

import (
	"fmt"
	"net"
	"os"
  "strconv"
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

// Client --
func Client(serverPort int, clientPort int) {
  remote := os.Args[2]
	ownIP := GetOutboundIP().String()
  msgBuf := make([]byte, 1024)

  // Resolve the passed address as UDP4
	toAddr, err := net.ResolveUDPAddr("udp4", remote + ":" + strconv.Itoa(serverPort))
	if err != nil {
		fmt.Printf("Could not resolve %s:%d\n", remote, serverPort)
		return
	}

  fromAddr, err := net.ResolveUDPAddr("udp4", ownIP + ":" + strconv.Itoa(clientPort))
	if err != nil {
		fmt.Printf("Could not resolve %s\n", ownIP + ":" + strconv.Itoa(clientPort))
		return
	}

	fmt.Printf("Trying to punch a hole to %s:%d\n", remote, serverPort)

	// Initiate the transaction (force IPv4 to demo firewall punch)
	conn, err := net.DialUDP("udp4", fromAddr, toAddr)
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
