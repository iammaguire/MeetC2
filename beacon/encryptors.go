package main

import b64 "encoding/base64"

func (enc Base64Encoder) scramble() []byte {
	return []byte(b64.StdEncoding.EncodeToString(enc.data))
}