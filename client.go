package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type message struct {
	content string
	addr    *net.UDPAddr
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Printf("Error when getting local IP")
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func requestServer(msgBuf []byte, conn *net.UDPConn, relayAddr net.Addr) (string, error) {

	fmt.Printf("Trying to punch a hole to %s\n", relayAddr.String())

	// Register client to server "test"
	sendMessage(msgBuf, conn, "CLIENT test", relayAddr)

	// Await relay registation ack
	rcvLen, addr, err := conn.ReadFrom(msgBuf)
	if err != nil {
		fmt.Println("Transaction was initiated but encountered an error!")
		return "", err
	}

	fmt.Printf("Received a packet from: %s\n\tResult: %s\n",
		addr.String(), msgBuf[:rcvLen])

	return string(msgBuf[:rcvLen]), nil
}

func keepAlive(msgBuf []byte, ln *net.UDPConn, c chan bool) {
	for {
		select {
		case res := <-c:
			if res {
				//fmt.Println("DELAYING KEEP ALIVE")
			}
		case <-time.After(20 * time.Second):
			//fmt.Println("SENDING KEEP ALIVE")
			fmt.Fprintf(ln, "KEEP ALIVE")
		}
	}
}

func processMessages(msgBuf []byte, ln *net.UDPConn, c chan message) {
	for {
		msgLen, originAddr, err := ln.ReadFromUDP(msgBuf)
		if err != nil {
			fmt.Printf("Error reading UDP response!\n")
			return
		}

		c <- message{content: string(msgBuf[:msgLen]), addr: originAddr}
	}
}

func userInputHandler(msgBuf []byte, ln *net.UDPConn) {
	reader := bufio.NewReader(os.Stdin)

	for {
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		text = strings.Replace(text, "\r", "", -1)

		fmt.Fprintf(ln, text)
	}
}

// Client --
func Client(clientPort int, relayPort int) {
	msgBuf := make([]byte, 1024)
	relay := os.Args[2]

	// Create connection
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: clientPort})
	if err != nil {
		fmt.Printf("Unable to listen on :%d\n", clientPort)
		return
	}
	defer conn.Close()

	// Resolve the passed (relay) address as UDP4
	relayAddr, err := net.ResolveUDPAddr("udp4", relay+":"+strconv.Itoa(relayPort))
	if err != nil {
		fmt.Printf("Could not resolve %s:%d\n", relay, relayPort)
		return
	}

	// Get server ip and port through relay and request hole punching
	server, err := requestServer(msgBuf, conn, relayAddr)
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

	if !strings.HasPrefix(string(msgBuf[:rcvLen]), "PUNCHED ") {
		return
	}

	// Resolve the received (server) address as UDP4
	serverAddr, err := net.ResolveUDPAddr("udp4", server)
	if err != nil {
		fmt.Printf("Could not resolve %s\n", server)
		return
	}

	fmt.Printf("Trying to punch a hole to %s\n", server)

	// Send message creating the hole and starting the handshake
	sendMessage(msgBuf, conn, "Hello mr. Server", serverAddr)

	connected := false
	for !connected {
		// Await a response through our NAT hole
		msgLen, originAddr, err := conn.ReadFromUDP(msgBuf)
		if err != nil {
			fmt.Printf("Error reading UDP response!\n")
			return
		}

		fmt.Printf("Received a UDP packet back from %s:%d\n\tResponse: %s\n",
			originAddr.IP, originAddr.Port, msgBuf[:msgLen])

		// Completed handshake
		if string(msgBuf[:msgLen]) == "Hello client!" {
			fmt.Println("Success: NAT traversed! ^-^")
			connected = true
		}
	}

	//TODO: review from here
	alive := true
	c1 := make(chan bool)
	c2 := make(chan message)

	//spawn process for keep alive sending and for message processing
	go keepAlive(msgBuf, conn, c1)
	go processMessages(msgBuf, conn, c2)
	go userInputHandler(msgBuf, conn)

	for alive {
		select {
		case msg := <-c2:
			c1 <- true
			if msg.content == "KEEP ALIVE" {
				//fmt.Println("Still alive! ^-^")
			} else if !strings.HasPrefix(msg.content, "WALK ") {
				fmt.Println("\r" + msg.content + "\r")
			}
		//if no message in 25 seconds connection is lost
		case <-time.After(25 * time.Second):
			fmt.Println("LOST CONNECTION")
			alive = false
		}
	}

}
