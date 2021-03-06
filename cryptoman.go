package main

import (
    "io"
    "io/ioutil"
    "crypto/aes"
    "crypto/rand"
    "crypto/cipher"
)

type SecurityContext struct {
	key []byte
}

func newSecurityContext() *SecurityContext {
    context := &SecurityContext{ []byte{} }
    context.loadKey()
    return context  
}

func (sc *SecurityContext) loadKey() {
	key, err := ioutil.ReadFile("./includes/sharedkey.txt")

    if err != nil {
        panic("[!] No key in includes/sharedkey.txt. Exiting!")
    }

    sc.key = key
}

func (sc *SecurityContext) encrypt(msg []byte) []byte {
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
	return enc
}


func (sc *SecurityContext) decrypt(encrypted []byte) string {
	block, err := aes.NewCipher(sc.key)

	if err != nil {
		info(err.Error())
        return ""
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		info(err.Error())
        return ""
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		info(err.Error())
        return ""
	}

	return string(plaintext)
}