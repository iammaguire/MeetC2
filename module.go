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
	setSource(string)
	getName() string
	getLanguage() string
	getSourceFromDisk() string
	getShellcode() []byte
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

func (mod CSharpModule) getShellcode() []byte {
	fileName, err := mod.compile(false)
	shellcodeFileName := "modules/" + mod.Name + ".bin"

	if err != nil {
		info("Failed to compile module: " + err.Error())
		return []byte{}
	}

	output := ""
	cmdHandle := exec.Command("/bin/sh", "-c", "./includes/donut " + fileName + " -o " + shellcodeFileName)
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

	info(output)

	data, err := ioutil.ReadFile(shellcodeFileName)

	//os.Remove(fileName)
	//os.Remove(shellcodeFileName)

	if err != nil {
		return []byte{}
	} else {
		return data
	}
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

func (mod CSharpModule) compile(delete bool) (string, error) {
	filename := "/tmp/" + genRandID()
	outfile := "modules/" + mod.Name + ".exe"
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

	if delete {
		os.Remove(outfile)
	}
	
	if len(output) == 0 {
		return outfile, nil
	} else {
		return outfile, fmt.Errorf("%s", output)
	}
} 