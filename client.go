package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
  "strconv"
)

// Client --
func Client(localPort int) {
  remote := os.Args[2]
  msgBuf := make([]byte, 1024)

  // Resolve the passed address as UDP4
	toAddr, err := net.ResolveUDPAddr("udp4", remote + ":" + strconv.Itoa(localPort))
	if err != nil {
		fmt.Printf("Could not resolve %s:%d\n", remote, localPort)
		return
	}

	fmt.Printf("Trying to punch a hole to %s:%d\n", remote, localPort)

	// Initiate the transaction (force IPv4 to demo firewall punch)
	conn, err := net.DialUDP("udp4", nil, toAddr)
	defer conn.Close()

	if err != nil {
		fmt.Printf("Unable to connect to %s:%d\n", remote, localPort)
		return
	}

	// Initiate the transaction, creating the hole
	msg := "trying..."
	fmt.Fprintf(conn, msg)
	fmt.Printf("Sent a UDP packet to %s:%d\n\tSent: %s\n", remote, localPort, msg)

	// Await a response through our firewall hole
	msgLen, fromAddr, err := conn.ReadFromUDP(msgBuf)
	if err != nil {
		fmt.Printf("Error reading UDP response!\n")
		return
	}

	fmt.Printf("Received a UDP packet back from %s:%d\n\tResponse: %s\n",
		fromAddr.IP, fromAddr.Port, msgBuf[:msgLen])

	fmt.Println("Success: NAT traversed! ^-^")
	//register()
}

func register() {
	signalAddress := os.Args[2]

	localAddress := ":9595" // default port
	if len(os.Args) > 3 {
		localAddress = os.Args[3]
	}

	remote, _ := net.ResolveUDPAddr("udp", signalAddress)
	local, _ := net.ResolveUDPAddr("udp", localAddress)
	conn, _ := net.ListenUDP("udp", local)
	go func() {
		bytesWritten, err := conn.WriteTo([]byte("register"), remote)
		if err != nil {
			panic(err)
		}

		fmt.Println(bytesWritten, " bytes written")
	}()

	listen(conn, local.String())
}

func listen(conn *net.UDPConn, local string) {
	for {
		fmt.Println("listening")
		buffer := make([]byte, 1024)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("[ERROR]", err)
			continue
		}

		fmt.Println("[INCOMING]", string(buffer[0:bytesRead]))
		if string(buffer[0:bytesRead]) == "Hello!" {
			continue
		}

		for _, a := range strings.Split(string(buffer[0:bytesRead]), ",") {
			if a != local {
				go chatter(conn, a)
			}
		}
	}
}

func chatter(conn *net.UDPConn, remote string) {
	addr, _ := net.ResolveUDPAddr("udp", remote)
	for {
		conn.WriteTo([]byte("Hello!"), addr)
		fmt.Println("sent Hello! to ", remote)
		time.Sleep(5 * time.Second)
	}
}
