package main

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "io"
)

type SecurityContext struct {
	key []byte
}

func (sc SecurityContext) getKey() {
	key := make([]byte, 32)
    _, err := rand.Read(key)

	if err != nil {
		info(err.Error())
	} else {
		sc.key = key
	}
}

func (sc SecurityContext) encrypt(msg []byte) []byte {
	c, err := aes.NewCipher(sc.key)
    
	if err != nil {
        info(err.Error())
		return []byte{}
    }

    gcm, err := cipher.NewGCM(c)

    if err != nil {
        info(err.Error())
    }

    nonce := make([]byte, gcm.NonceSize())
    
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        info(err.Error())
    }

    enc := gcm.Seal(nonce, nonce, msg, nil)
	info(string(enc))
	return enc
}