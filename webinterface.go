package main

import (
	"log"
	"time"
	"strconv"
	"net/http"
	"github.com/gorilla/mux"
)

type IWebInterface interface {
	startListener() (error)
	beaconsHandler(http.ResponseWriter, *http.Request)
}

type WebInterface struct {
	port int
}

func (server WebInterface) startListener() (error) {
	var router = mux.NewRouter()
	router.HandleFunc("/api/beacons", server.beaconsHandler).Methods("Get")
	staticFileDirectory := http.Dir("./www/")
	staticFileHandler := http.StripPrefix("/c2/", http.FileServer(staticFileDirectory))
	router.PathPrefix("/c2/").Handler(staticFileHandler).Methods("GET")

	srv := &http.Server{
        Handler:      router,
        Addr:         "127.0.0.1:" + strconv.Itoa(server.port),
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

	go func() {
    	log.Fatal(srv.ListenAndServe())	
	}()

	return nil
}