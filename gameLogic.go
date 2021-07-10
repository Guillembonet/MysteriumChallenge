package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

func zombieMover(msgBuf []byte, ln *net.UDPConn, gameElement *game, c chan net.Addr, position *coordinate) {

	finished := false
	for !finished {
		select {
		// SOMEONE WON
		case res := <-c:
			sendMessage(msgBuf, ln, "BOOM "+res.String()+" 1 night-king", res, true)
			finished = true
			for _, element := range gameElement.clients {
				if gameElement.winner.String() == element.String() {
					sendMessage(msgBuf, ln, "YOU WIN", element, true)
				} else {
					sendMessage(msgBuf, ln, "YOU LOST", element, true)
				}
			}
		case <-time.After(2 * time.Second):
			// MOVE LOGIC
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
			// BROADCAST MOVE
			for _, element := range gameElement.clients {
				sendMessage(msgBuf, ln, "WALK night-king "+strconv.Itoa(position.x)+" "+strconv.Itoa(position.y), element, true)
			}
			// IF ZOMBIE REACHED THE WALL
			if position.x >= 30 {
				finished = true
				gameElement.channel <- order{client: nil, value: "END"}
			}
		}
	}

}

func startGame(msgBuf []byte, ln *net.UDPConn, gameElement *game) {

	ch := make(chan net.Addr)
	position := coordinate{x: 0, y: 5}
	go zombieMover(msgBuf, ln, gameElement, ch, &position)

	for {
		order := <-gameElement.channel
		if strings.HasPrefix(order.value, "SHOOT ") {
			// CHECK SHOT
			shotString := strings.Split(order.value, " ")
			x, _ := strconv.Atoi(shotString[2])
			y, _ := strconv.Atoi(shotString[3])
			fmt.Printf("SHOOT at %d,%d and zombie at %d,%d\n", x, y, position.x, position.y)
			if x == position.x && y == position.y {
				// CLIENT WON
				*gameElement = game{started: true, clients: gameElement.clients, channel: gameElement.channel, winner: order.client, ended: true}
				ch <- order.client
			} else {
				sendMessage(msgBuf, ln, "BOOM "+order.client.String()+" 0", order.client, true)
			}
		} else if order.value == "END" {
			*gameElement = game{started: true, clients: gameElement.clients, channel: gameElement.channel, winner: nil, ended: true}
			fmt.Println("Zombie won")
			for _, element := range gameElement.clients {
				sendMessage(msgBuf, ln, "YOU LOST", element, true)
			}
		}
	}
}
