package main

type BeaconHttp struct {
	method string
	data []byte
}

type Base64Encoder struct {
	data []byte
}

type CommandResponse struct {
	Exec []string
	Download []string
}

type CommandUpdate struct {
	Ip string
	Id string
	Type string
	Data []byte
}
