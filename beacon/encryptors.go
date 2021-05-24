package main

import b64 "encoding/base64"

type Base64Encoder struct {
	data []byte
}

type IPIDEncoder struct {
	data []byte
}

func (enc Base64Encoder) scramble() []byte {
	return []byte(b64.StdEncoding.EncodeToString(enc.data))
}

func (enc IPIDEncoder) scramble() []byte {
	return enc.data
}