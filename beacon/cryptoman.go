package main

import (
    "io"
    "fmt"
    "crypto/aes"
    "crypto/rand"
    "crypto/cipher"
)

type BeaconSecurityContext struct {
	key []byte
}

func newBeaconSecurityContext() *BeaconSecurityContext {
    context := &BeaconSecurityContext{ []byte{} }
    return context  
}

func (sc *BeaconSecurityContext) encrypt(msg []byte) []byte {
	c, err := aes.NewCipher([]byte(secret))
    
	if err != nil {
        debugFatal(err)
		return []byte{}
    }

    gcm, err := cipher.NewGCM(c)

    if err != nil {
        debugFatal(err)
    }

    nonce := make([]byte, gcm.NonceSize())
    
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        debugFatal(err)
    }

    enc := gcm.Seal(nonce, nonce, msg, nil)
	return enc
}


func (sc *BeaconSecurityContext) decrypt(encrypted []byte) string {
	block, err := aes.NewCipher([]byte(secret))

	if err != nil {
		debugFatal(err)
        fmt.Println(err)
        return ""
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		debugFatal(err)
        fmt.Println(err)
        return ""
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		debugFatal(err)
        fmt.Println(err)
        return ""
	}

	return string(plaintext)
}