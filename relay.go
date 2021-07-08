package main

import (
	"fmt"
	"net"
	"strconv"
  "strings"
)

type server struct {
  ip string
  port  int
}

// Relay --
func Relay(relayPort int) {
	msgBuf := make([]byte, 1024)
  reply := ""

  //list of servers by name
  servers := make(map[string]server)

	//Create connection
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: relayPort})
  if err != nil {
		fmt.Printf("Unable to listen on :%d\n", relayPort)
		return
	}
  defer conn.Close()

  //await messages
  for {
  	msgLen, originAddr, err := conn.ReadFromUDP(msgBuf)
  	if err != nil {
  		fmt.Printf("Error reading UDP response!\n")
  		return
  	}

  	fmt.Printf("Received a UDP packet back from %s:%d\n\tResponse: %s\n",
  		originAddr.IP, originAddr.Port, msgBuf[:msgLen])

    if (strings.HasPrefix(string(msgBuf[:msgLen]), "CLIENT ") && msgLen > 7) {
      fmt.Printf("Received a UDP packet back from %s:%d\n\tResponse: %s\n",
    		originAddr.IP, originAddr.Port, msgBuf[:msgLen])

      //find server with this name
      if val, ok := servers[string(msgBuf[7:msgLen])]; ok {
          reply = val.ip + ":" + strconv.Itoa(val.port)
      } else {
        reply = "none"
      }

      //reply using hole
      copy(msgBuf, []byte(reply))
  		_, err = conn.WriteTo(msgBuf[:len(reply)], originAddr)

  		if err != nil {
  			fmt.Println("Socket closed unexpectedly!")
  			continue
  		}

  		fmt.Printf("Sent reply to %s\n\tReply: %s\n",
  			originAddr.String(), msgBuf[:len(reply)])
    } else if (strings.HasPrefix(string(msgBuf[:msgLen]), "SERVER ") && msgLen > 7) {
      fmt.Printf("Received a UDP packet back from %s:%d\n\tResponse: %s\n",
    		originAddr.IP, originAddr.Port, msgBuf[:msgLen])

      //set server with this name
      servers[string(msgBuf[7:msgLen])] = server{ip: originAddr.IP.String(), port: originAddr.Port}
      reply = string(msgBuf[7:msgLen]) + " = " + originAddr.IP.String() + ":" + strconv.Itoa(originAddr.Port)


      //reply using hole
      copy(msgBuf, []byte(reply))
  		_, err = conn.WriteTo(msgBuf[:len(reply)], originAddr)

  		if err != nil {
  			fmt.Println("Socket closed unexpectedly!")
  			continue
  		}

  		fmt.Printf("Sent reply to %s\n\tReply: %s\n",
  			originAddr.String(), msgBuf[:len(reply)])
    }
  }
}
