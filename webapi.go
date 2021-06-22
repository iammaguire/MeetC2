package main

import (
	//"io/ioutil"
	//"fmt"
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

func (server WebInterface) newBeaconHandler(w http.ResponseWriter, h *http.Request) {
	platform := h.URL.Query()["platform"][0]
	arch := h.URL.Query()["arch"][0]

	processInput("create 0 " + platform + " " + arch)
}