package main

import (
	"os"
	"fmt"
	"bufio"
	"os/exec"
	"io/ioutil"
)

type Module interface {
	compile() error
}

type CSharpModule struct {
	source string
}

func (mod CSharpModule) compile() error {
	filename := "/tmp/" + genRandID()
	outfile := "/tmp/" + genRandID()
    err := ioutil.WriteFile(filename, []byte(mod.source), 0644)
	cmdHandle := exec.Command("/bin/sh", "-c", "mcs -out:" + outfile + " " + filename)
	
	stderr, err := cmdHandle.StderrPipe()
	stdout, err := cmdHandle.StdoutPipe()
	
	output := ""

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

	os.Remove(filename)
	os.Remove(outfile)

	if len(output) == 0 {
		return nil
	} else {
		return fmt.Errorf("Compilation error: %s\n", output)
	}
} 