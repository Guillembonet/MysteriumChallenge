package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type order struct {
	client net.Addr
	value  string
}

type game struct {
	started bool
	clients []net.Addr
	channel chan order
	winner  net.Addr
	ended   bool
}

type coordinate struct {
	x int
	y int
}

func registerServer(msgBuf []byte, conn *net.UDPConn, relayAddr net.Addr, serverPort int) {

	fmt.Printf("Trying to punch a hole to %s\n", relayAddr.String())

	// Initiate the transaction, creating the hole
	sendMessage(msgBuf, conn, "SERVER test", relayAddr, true)

	// Await server registation
	rcvLen, _, err := conn.ReadFrom(msgBuf)
	if err != nil {
		fmt.Println("Transaction was initiated but encountered an error!")
	}

	if strings.HasPrefix(string(msgBuf[:rcvLen]), "test = ") {
		fmt.Printf("Registration successful\n\tRegistered: %s\n",
			msgBuf[:rcvLen])
	}
}

func handleNewClient(msgBuf []byte, conn *net.UDPConn, clientAddrString string, relayAddr net.Addr) {
	// Resolve client
	clientAddr, err := net.ResolveUDPAddr("udp4", clientAddrString)
	if err != nil {
		fmt.Printf("Could not resolve %s\n", clientAddrString)
		return
	}

	reply := "PUNCHED " + clientAddrString
	// Punch hole
	sendMessage(msgBuf, conn, reply, clientAddr, true)

	// Ack to relay
	sendMessage(msgBuf, conn, reply, relayAddr, true)

	fmt.Printf("Sent punch to client %s\n",
		clientAddr.String())
}

// Server node
func Server(serverPort int, relayPort int) {
	msgBuf := make([]byte, 1024)

	// Create connection
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: serverPort})
	if err != nil {
		fmt.Printf("Unable to listen on :%d\n", serverPort)
		return
	}
	defer conn.Close()

	relay := os.Args[2]

	// Resolve the passed (relay) address as UDP4
	relayAddr, err := net.ResolveUDPAddr("udp4", relay+":"+strconv.Itoa(relayPort))
	if err != nil {
		fmt.Printf("Could not resolve %s:%d\n", relay, relayPort)
		return
	}

	// Register Server
	registerServer(msgBuf, conn, relayAddr, serverPort)

	// map of games by name
	games := make(map[string]game)

	fmt.Printf("Listening on :%d\n", serverPort)

	for {
		fmt.Println("---")
		// Await incoming packets
		rcvLen, addr, err := conn.ReadFrom(msgBuf)
		if err != nil {
			fmt.Println("Transaction was initiated but encountered an error!")
			continue
		}

		fmt.Printf("Received a packet from: %s\n\tSays: %s\n",
			addr.String(), msgBuf[:rcvLen])

		// NEW CLIENT EVENT
		if strings.HasPrefix(string(msgBuf[:rcvLen]), "CLIENT ") {
			handleNewClient(msgBuf, conn, string(msgBuf[7:rcvLen]), relayAddr)

			//HANDLE KEEP ALIVES
		} else if string(msgBuf[:rcvLen]) == "KEEP ALIVE" {
			sendMessage(msgBuf, conn, "KEEP ALIVE", addr, true)

			// CREATE GAME
		} else if strings.HasPrefix(string(msgBuf[:rcvLen]), "CREATE ") {
			key := string(msgBuf[7:rcvLen])
			games[key] = game{started: false, clients: []net.Addr{addr}, channel: make(chan order), winner: nil, ended: false}

			sendMessage(msgBuf, conn, "CREATED "+string(msgBuf[7:rcvLen]), addr, true)

			// JOIN GAME
		} else if strings.HasPrefix(string(msgBuf[:rcvLen]), "JOIN ") {

			reply := "Game does not exist "
			key := string(msgBuf[5:rcvLen])
			if val, ok := games[key]; ok {
				if !Exists(val.clients, addr) {
					if !val.started {
						val.clients = append(val.clients, addr)
						reply = "JOINED"
					} else {
						reply = "Game already started"
					}
				} else {
					reply = "Already joined"
				}
			}
			sendMessage(msgBuf, conn, reply, addr, true)

			// START GAME
		} else if strings.HasPrefix(string(msgBuf[:rcvLen]), "START ") {

			key := string(msgBuf[6:rcvLen])
			if val, ok := games[key]; ok {
				if !val.started {
					games[key] = game{started: true, clients: val.clients, channel: val.channel, winner: nil, ended: false}
					gameElement := games[key]
					go startGame(msgBuf, conn, &gameElement)
				} else {
					reply := "Game already started"
					sendMessage(msgBuf, conn, reply, addr, true)
				}
			}

			// HANDLE SHOTS
		} else if strings.HasPrefix(string(msgBuf[:rcvLen]), "SHOOT ") {

			orderValue := string(msgBuf[:rcvLen])
			gameName := strings.Split(orderValue, " ")[1]
			// If game is running and client is a member send the order to the game channel
			if val, ok := games[gameName]; ok {
				if val.started && !val.ended && Exists(games[gameName].clients, addr) {
					games[gameName].channel <- order{value: orderValue, client: addr}
				} else {
					reply := "Game ended, not started or did not join"
					sendMessage(msgBuf, conn, reply, addr, true)
				}
			} else {
				reply := "Game does not exist"
				sendMessage(msgBuf, conn, reply, addr, true)
			}

			// HANDLE HANDSHAKE
		} else if string(msgBuf[:rcvLen]) == "Hello mr. Server" {
			// Let the client confirm a hole was punched through to us
			sendMessage(msgBuf, conn, "Hello client!", addr, true)
		}
	}
}
