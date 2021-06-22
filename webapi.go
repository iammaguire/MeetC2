package main

import (
	//"io/ioutil"
	"net/http"
	"encoding/json"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func (server WebInterface) wsHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		info(err.Error())
		return
	}
	
	wsWriter = w
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 1024)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func (server WebInterface) updateHandler(w http.ResponseWriter, h *http.Request) {
	json.NewEncoder(w).Encode(webInterfaceUpdates)
	webInterfaceUpdates = webInterfaceUpdates[:0]
}

func (server WebInterface) beaconsHandler(w http.ResponseWriter, h *http.Request) {
	json.NewEncoder(w).Encode(beacons)
}

func (server WebInterface) listenersHandler(w http.ResponseWriter, h *http.Request) {
	json.NewEncoder(w).Encode(listeners)
}