package main

type BeaconHttp struct {
	method string
	data []byte
}

type BeaconIPID struct {
	data []byte
}

type CommandResponse struct {
	Exec []string
	Download []string
	Upload []string
}

type CommandUpdate struct {
	Ip string
	Id string
	Type string
	Data []byte
}
