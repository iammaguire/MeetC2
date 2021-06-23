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
	getName() string
	getLanguage() string
	getSourceFromDisk() string
	setSource(string)
	writeToDisk()
}

type CSharpModule struct {
	Name string  
	Source string
	Language string
}

func newCSharpModule(name string, source string) *CSharpModule {
	return &CSharpModule { name, source, "C#" }
}

func (mod CSharpModule) writeToDisk() {
	ioutil.WriteFile("modules/" + mod.Name + ".cs", []byte(mod.Source), 0644)
}

func (mod CSharpModule) setSource(source string) {
	mod.Source = source
}

func (mod CSharpModule) getSourceFromDisk() string {
	source, err := ioutil.ReadFile("modules/" + mod.Name + ".cs")
	
	if err != nil {
		return ""
	} else {
		mod.Source = string(source)
		return mod.Source
	}
}

func (mod CSharpModule) getLanguage() string {
	return "C#"
}

func (mod CSharpModule) getName() string {
	return mod.Name
}

func (mod CSharpModule) compile() error {
	filename := "/tmp/" + genRandID()
	outfile := "/tmp/" + genRandID()
    err := ioutil.WriteFile(filename, []byte(mod.Source), 0644)
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
		return fmt.Errorf("%s", output)
	}
} 