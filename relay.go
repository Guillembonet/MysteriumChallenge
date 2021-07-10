package main

import (
	"fmt"
	"net"
	"strings"
)

type server struct {
	addr net.Addr
	port int
}

func HandleClientRegistration(msgBuf []byte, conn *net.UDPConn, servers map[string]net.Addr, serverName string, clientAddr net.Addr) {
	// find server with this name in dictionary
	reply := "none"
	if serverAddr, ok := servers[serverName]; ok {
		reply = serverAddr.String()

		sendMessage(msgBuf, conn, "CLIENT "+serverAddr.String(), serverAddr)
	}

	// reply the server ip and port to the client
	sendMessage(msgBuf, conn, reply, clientAddr)
}

func HandleServerRegistration(msgBuf []byte, conn *net.UDPConn, servers map[string]net.Addr, serverName string, serverAddr net.Addr) {
	// set server id with received name
	servers[serverName] = serverAddr
	reply := serverName + " = " + serverAddr.String()

	// reply ack using hole
	sendMessage(msgBuf, conn, reply, serverAddr)
}

func handleHolePunch(msgBuf []byte, conn *net.UDPConn, servers map[string]net.Addr, clientAddrString string, serverAddr net.Addr) {
	// resolve client address
	clientAddr, err := net.ResolveUDPAddr("udp4", clientAddrString)
	if err != nil {
		fmt.Printf("Could not resolve %s\n", clientAddrString)
		return
	}

	// notify client that hole has been punched
	sendMessage(msgBuf, conn, "PUNCHED test", clientAddr)
}

// Relay --
func Relay(relayPort int) {
	msgBuf := make([]byte, 1024)

	// map of servers by name
	servers := make(map[string]net.Addr)

	// Create connection
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: relayPort})
	if err != nil {
		fmt.Printf("Unable to listen on :%d\n", relayPort)
		return
	}
	defer conn.Close()

	for {
		// await messages
		msgLen, originAddr, err := conn.ReadFromUDP(msgBuf)
		if err != nil {
			fmt.Printf("Error reading UDP response!\n")
			return
		}

		fmt.Printf("Received a UDP packet back from %s:%d\n\tResponse: %s\n",
			originAddr.IP, originAddr.Port, msgBuf[:msgLen])

		// CLIENT REGISTATION EVENT
		if strings.HasPrefix(string(msgBuf[:msgLen]), "CLIENT ") && msgLen > 7 {
			HandleClientRegistration(msgBuf, conn, servers, string(msgBuf[7:msgLen]), originAddr)

			// SERVER REGISTATION EVENT
		} else if strings.HasPrefix(string(msgBuf[:msgLen]), "SERVER ") && msgLen > 7 {
			HandleServerRegistration(msgBuf, conn, servers, string(msgBuf[7:msgLen]), originAddr)

			// SERVER CREATED HOLE FOR CLIENT EVENT
		} else if strings.HasPrefix(string(msgBuf[:msgLen]), "PUNCHED ") && msgLen > 8 {
			handleHolePunch(msgBuf, conn, servers, string(msgBuf[8:msgLen]), originAddr)
		}
	}
}
