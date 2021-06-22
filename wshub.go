package main

import (
	"time"
	"strings"
)

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
			for client := range h.clients {
				select {
				case client.send <- message:
					if(strings.HasPrefix(string(message), "meet2c:")) {
						client.beaconId = strings.Split(string(message), ":")[1]
					}

					processInput(string(message))
					time.Sleep(time.Millisecond * 500)
					client.send <- stdOutBuffer
					stdOutBuffer = stdOutBuffer[:0]
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}