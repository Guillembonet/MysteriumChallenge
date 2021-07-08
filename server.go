package main

import (
	"fmt"
	"net"
	"os"
  "strconv"
)

// Server --
func Server(serverPort int, clientPort int) {
	//msgBuf := make([]byte, 1024)

	remote := os.Args[2]
	ownIP := GetOutboundIP().String()

  // Resolve the passed (client) address as UDP4
	toAddr, err := net.ResolveUDPAddr("udp4", remote + ":" + strconv.Itoa(clientPort))
	if err != nil {
		fmt.Printf("Could not resolve %s:%d\n", remote, clientPort)
		return
	}

	// Resolve the server address as UDP4
  fromAddr, err := net.ResolveUDPAddr("udp4", ownIP + ":" + strconv.Itoa(serverPort))
	if err != nil {
		fmt.Printf("Could not resolve %s\n", ownIP + ":" + strconv.Itoa(serverPort))
		return
	}

	fmt.Printf("Trying to punch a hole to %s:%d\n", remote, clientPort)

	// Initiate the transaction (force IPv4 to demo firewall punch)
	conn, err := net.DialUDP("udp4", fromAddr, toAddr)

	if err != nil {
		fmt.Printf("Unable to connect to %s:%d\n", remote, clientPort)
		return
	}

	// Initiate the transaction, creating the hole
	msg := "punch!"
	fmt.Fprintf(conn, msg)
	fmt.Printf("Sent a UDP packet to %s:%d from %s\n\tSent: %s\n", remote, clientPort, fromAddr, msg)
	defer conn.Close()

	// // Initiatlize a UDP listener
	// ln, err := net.ListenUDP("udp4", &net.UDPAddr{Port: serverPort})
	// if err != nil {
	// 	fmt.Printf("Unable to listen on :%d\n", serverPort)
	// 	return
	// }

	// fmt.Printf("Listening on :%d\n", serverPort)

	// for {
	// 	fmt.Println("---")
	// 	// Await incoming packets
	// 	rcvLen, addr, err := conn.ReadFrom(msgBuf)
	// 	if err != nil {
	// 		fmt.Println("Transaction was initiated but encountered an error!")
	// 		continue
	// 	}
	//
	// 	fmt.Printf("Received a packet from: %s\n\tSays: %s\n",
	// 		addr.String(), msgBuf[:rcvLen])
	//
	// 	// Let the client confirm a hole was punched through to us
	// 	reply := "hole punched!"
	// 	copy(msgBuf, []byte(reply))
	// 	_, err = conn.WriteTo(msgBuf[:len(reply)], addr)
	//
	// 	if err != nil {
	// 		fmt.Println("Socket closed unexpectedly!")
	// 		continue
	// 	}
	//
	// 	fmt.Printf("Sent reply to %s\n\tReply: %s\n",
	// 		addr.String(), msgBuf[:len(reply)])
	// }
}
