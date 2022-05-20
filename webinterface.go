package main

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type IWebInterface interface {
	startListener() error
	beaconsHandler(http.ResponseWriter, *http.Request)
	listenersHandler(http.ResponseWriter, *http.Request)
	commandHandler(http.ResponseWriter, *http.Request)
	terminalHandler(http.ResponseWriter, *http.Request)
	newBeaconHandler(http.ResponseWriter, *http.Request)
	newHTTPListenerHandler(http.ResponseWriter, *http.Request)
	netInterfacesHandler(http.ResponseWriter, *http.Request)
	modulesHandler(http.ResponseWriter, *http.Request)
	newModuleHandler(http.ResponseWriter, *http.Request)
}

type WebInterface struct {
	ip   string
	port int
}

type WebUpdate struct {
	Title string
	Msg   string
}

var webInterfaceUpdates []*WebUpdate = make([]*WebUpdate, 0)
var stdOutBuffer string
var redirectStdIn bool = false
var wsWriter http.ResponseWriter
var hub *Hub

func (server WebInterface) startListener() error {
	mime.AddExtensionType(".js", "application/javascript")
	hub = newHub()
	go hub.run()

	var router = mux.NewRouter()
	router.HandleFunc("/api/beacons", server.beaconsHandler).Methods("Get")
	router.HandleFunc("/api/listeners", server.listenersHandler).Methods("Get")
	router.HandleFunc("/api/updates", server.updateHandler).Methods("Get")
	router.HandleFunc("/api/newbeacon", server.newBeaconHandler).Methods("Get")
	router.HandleFunc("/api/newhttplistener", server.newHTTPListenerHandler).Methods("Get")
	router.HandleFunc("/api/netifaces", server.netInterfacesHandler).Methods("Get")
	router.HandleFunc("/api/modules", server.modulesHandler).Methods("Get")
	router.HandleFunc("/api/compile", server.compileHandler).Methods("Get")
	router.HandleFunc("/api/updatemodule", server.updateModuleHandler).Methods("Get")
	router.HandleFunc("/api/ws", func(w http.ResponseWriter, r *http.Request) { server.wsHandler(hub, w, r) })

	staticFileDirectory := http.Dir("./www/")
	staticFileHandler := http.StripPrefix("/c2/", http.FileServer(staticFileDirectory))
	router.PathPrefix("/c2/").Handler(staticFileHandler).Methods("GET")

	srv := &http.Server{
		Handler:      router,
		Addr:         server.ip + ":" + strconv.Itoa(server.port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		log.Fatal(srv.ListenAndServe())
		fmt.Println("Web interface killed")
	}()

	fmt.Println("Web interface listening on " + server.ip + ":" + strconv.Itoa(server.port))

	return nil
}
