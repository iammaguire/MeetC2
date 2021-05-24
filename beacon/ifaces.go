package main 	

const BYTES = 0;
const STRING = 1;

type EncType interface {
    scramble() []byte
}

type Request interface {
	queryServer()
    handleQueryResponse(CommandResponse)
    upload(string)
    download(string)
}

type CommType interface {
    EncType
	Request
}