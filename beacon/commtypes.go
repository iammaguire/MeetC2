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
	ProxyClients []string
}

type CommandUpdate struct {
	Ip string
	Id string
	Type string
	Data []byte
}
