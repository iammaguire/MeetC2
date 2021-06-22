package main

import (
	//"os"
	//"fmt"
	"time"
	"strings"
)

var terminalPipe chan string = make(chan string)

type Hub struct {
	clients map[*Client]bool
	broadcast chan []byte
	register chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	go func() {
		for {
			for client := range h.clients {
				client.send <- []byte{}
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()
	
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			if redirectStdIn {
				terminalPipe <- strings.Join(strings.Split(string(message), ":")[2:], ":")
			} else {
				msgSplit := strings.Split(string(message), ":")
				//fmt.Println(strings.Join(msgSplit[2:], ":"))
				if msgSplit[0] == "beacon" { 
					processInput("use " + string(msgSplit[1]))
					processInput(strings.Join(msgSplit[2:], ":"))					
				} else if msgSplit[0] == "main" {
					processInput(strings.Join(msgSplit[2:], ":"))
				}
			}
		}
	}
}