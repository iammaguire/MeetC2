package main

import (
	//"io/ioutil"
	//"fmt"
	"net"
	"strings"
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

func (server WebInterface) updateModuleHandler(w http.ResponseWriter, h *http.Request) {
	name := h.URL.Query()["name"][0]
	language := h.URL.Query()["language"][0]
	source := h.URL.Query()["source"][0]

	updateModule(name, language, source)
}

func (server WebInterface) modulesHandler(w http.ResponseWriter, h *http.Request) {
	json.NewEncoder(w).Encode(csharpModules)
}

func (server WebInterface) compileHandler(w http.ResponseWriter, h *http.Request) {
	name := h.URL.Query()["name"][0]

	for _, module := range csharpModules {
		if module.Name == name {
			_, err := module.compile(true)

			if err != nil {
				w.Write([]byte(err.Error()))
			} else {
				w.Write([]byte("Good"))
			}

			return
		}
	}

	w.Write([]byte("Backend error... module not found."))
}

func (server WebInterface) newHTTPListenerHandler(w http.ResponseWriter, h *http.Request) {
	iface := h.URL.Query()["interface"][0]
	hostname := h.URL.Query()["hostname"][0]
	port := h.URL.Query()["port"][0]
	
	processInput("httplistener " + iface + " " + hostname + " " + port)
}

func (server WebInterface) netInterfacesHandler(w http.ResponseWriter, h *http.Request) {
	var retIfaces []byte
	ifaces, _ := net.Interfaces()
	
	for _, iface := range ifaces {
		iname := iface.Name
		addrs, _ := iface.Addrs()
		if len(addrs) > 0 && addrs[0] != nil {
			retIfaces = append(retIfaces, []byte(iname + " " + strings.Split(addrs[0].String(), "/")[0] + "\n")...)
		}
	}
	
	w.Write(retIfaces)
}