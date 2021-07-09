package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type game struct {
	started bool
	clients []net.Addr
	channel chan string
}

type coordinate struct {
	x int
	y int
}

func registerServer(msgBuf []byte, conn *net.UDPConn, relay string, relayPort int, serverPort int) {
	ownIP := GetOutboundIP().String()

	// Resolve the passed (relay) address as UDP4
	toAddr, err := net.ResolveUDPAddr("udp4", relay+":"+strconv.Itoa(relayPort))
	if err != nil {
		fmt.Printf("Could not resolve %s:%d\n", relay, relayPort)
		return
	}

	// Resolve the server address as UDP4
	fromAddr, err := net.ResolveUDPAddr("udp4", ownIP+":"+strconv.Itoa(serverPort))
	if err != nil {
		fmt.Printf("Could not resolve %s\n", ownIP+":"+strconv.Itoa(serverPort))
		return
	}

	fmt.Printf("Trying to punch a hole to %s:%d\n", relay, relayPort)

	// Initiate the server registration
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

	// Await server registation
	rcvLen, addr, err := conn.ReadFrom(msgBuf)
	if err != nil {
		fmt.Println("Transaction was initiated but encountered an error!")
	}

	fmt.Printf("Received a packet from: %s\n\tRegistered: %s\n",
		addr.String(), msgBuf[:rcvLen])
}

func sendMessage(msgBuf []byte, ln *net.UDPConn, message string, address net.Addr) {

	copy(msgBuf, []byte(message))
	_, err := ln.WriteTo(msgBuf[:len(message)], address)

	if err != nil {
		fmt.Println("Socket closed unexpectedly!")
	}

	fmt.Printf("Sent reply to %s\n\tReply: %s\n",
		address.String(), msgBuf[:len(message)])
}

func zombieMover(msgBuf []byte, ln *net.UDPConn, game game, c chan net.Addr, position *coordinate) {

	finished := false
	for !finished {
		select {
		case res := <-c:
			sendMessage(msgBuf, ln, "YOU WON!", res)

			finished = true
		case <-time.After(2 * time.Second):
			rnd1 := rand.Intn(4)
			rnd2 := rand.Intn(6)
			if rnd1 != 2 {
				*position = coordinate{x: position.x + 1, y: position.y}
			}
			if rnd2 < 2 && position.y > 0 {
				*position = coordinate{x: position.x, y: position.y - 1}
			} else if rnd2 >= 4 && position.y < 10 {
				*position = coordinate{x: position.x, y: position.y + 1}
			}
			for _, element := range game.clients {
				sendMessage(msgBuf, ln, "WALK night-king "+strconv.Itoa(position.x)+" "+strconv.Itoa(position.y), element)
			}
			if position.x >= 30 {
				game.channel <- "END"
			}
		}
	}

}

func startGame(msgBuf []byte, ln *net.UDPConn, game game) {

	ch := make(chan net.Addr)
	position := coordinate{x: 0, y: 5}
	go zombieMover(msgBuf, ln, game, ch, &position)

	for {
		select {
		case order := <-game.channel:
			if ()
	}
}

// Server --
func Server(serverPort int, relayPort int) {
	msgBuf := make([]byte, 1024)
	var conn net.UDPConn

	relay := os.Args[2]

	// Register Server
	registerServer(msgBuf, &conn, relay, relayPort, serverPort)
	conn.Close()

	ln, err := net.ListenUDP("udp4", &net.UDPAddr{Port: serverPort})
	if err != nil {
		fmt.Printf("Unable to listen on :%d\n", serverPort)
		return
	}

	fmt.Printf("Listening on :%d\n", serverPort)

	// map of games by name
	games := make(map[string]game)

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

		// NEW CLIENT EVENT
		if strings.HasPrefix(string(msgBuf[:rcvLen]), "CLIENT ") {
			// Resolve client
			clientAddr, err := net.ResolveUDPAddr("udp4", string(msgBuf[7:rcvLen]))
			if err != nil {
				fmt.Printf("Could not resolve %s\n", string(msgBuf[7:rcvLen]))
				return
			}

			// Resolve relay
			relayAddr, err := net.ResolveUDPAddr("udp4", relay+":"+strconv.Itoa(relayPort))
			if err != nil {
				fmt.Printf("Could not resolve %s:%d\n", relay, relayPort)
				return
			}

			reply := "PUNCHED " + string(msgBuf[7:rcvLen])
			// Punch hole
			sendMessage(msgBuf, ln, reply, clientAddr)

			// Ack to relay
			sendMessage(msgBuf, ln, reply, relayAddr)

			fmt.Printf("Sent punch to client %s\n",
				clientAddr.String())

			//HANDLE KEEP ALIVES
		} else if string(msgBuf[:rcvLen]) == "KEEP ALIVE" {
			sendMessage(msgBuf, ln, "KEEP ALIVE", addr)

			// CREATE GAME
		} else if strings.HasPrefix(string(msgBuf[:rcvLen]), "CREATE ") {
			games[string(msgBuf[7:rcvLen])] = game{started: false, clients: []net.Addr{addr}, channel: make(chan string)}

			sendMessage(msgBuf, ln, "CREATED "+string(msgBuf[7:rcvLen]), addr)

			// JOIN GAME
		} else if strings.HasPrefix(string(msgBuf[:rcvLen]), "JOIN ") {

			reply := "Game does not exist "
			if val, ok := games[string(msgBuf[5:rcvLen])]; ok {
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
			sendMessage(msgBuf, ln, reply, addr)

			// START GAME
		} else if strings.HasPrefix(string(msgBuf[:rcvLen]), "START ") {

			key := string(msgBuf[6:rcvLen])
			if val, ok := games[key]; ok {
				if !val.started {
					games[key] = game{started: true, clients: val.clients, channel: val.channel}
					go startGame(msgBuf, ln, game)
				} else {
					reply := "Game already started"
					sendMessage(msgBuf, ln, reply, addr)
				}
			}

			// HANDLE MESSAGES
		} else {
			// Let the client confirm a hole was punched through to us
			reply := "Hello client!"
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
