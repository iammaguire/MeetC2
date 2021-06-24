package main

import (
	"os"
	"fmt"
	"bufio"
	"os/exec"
	"io/ioutil"
)

type IModule interface {
	compile() error
	setSource(string)
	getName() string
	getLanguage() string
	getSourceFromDisk() string
	getShellcode() []byte
	writeToDisk()
}

type Module struct {
	Name string  
	Source string
	Language string
	extension string
}

func newModule(name string, source string, language string) *Module {
	var extension string

	if language == "Go" {
		extension = ".go"
	} else if language == "C#" {
		extension = ".cs"
	}

	return &Module { name, source, language, extension }
}

func (mod Module) getShellcode() []byte {
	fileName, err := mod.compile(false)
	shellcodeFileName := "modules/" + mod.Name + ".bin"

	if err != nil {
		info("Failed to compile module: " + err.Error())
		return []byte{}
	}

	output := ""
	cmdHandle := exec.Command("/bin/sh", "-c", "./includes/donut -x 1 -c TestModule2 -m Main " + fileName + " -o " + shellcodeFileName)
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

func (mod Module) writeToDisk() {
	ioutil.WriteFile("modules/" + mod.Name + mod.extension, []byte(mod.Source), 0644)
}

func (mod Module) setSource(source string) {
	mod.Source = source
}

func (mod Module) getSourceFromDisk() string {
	source, err := ioutil.ReadFile("modules/" + mod.Name + mod.extension)
	
	if err != nil {
		return ""
	} else {
		mod.Source = string(source)
		return mod.Source
	}
}

func (mod Module) getLanguage() string {
	return "C#"
}

func (mod Module) getName() string {
	return mod.Name
}

func (mod Module) compile(delete bool) (string, error) {
	var cmdHandle *exec.Cmd
	filename := "/tmp/" + genRandID()
	outfile := "modules/" + mod.Name + ".exe"
    err := ioutil.WriteFile(filename, []byte(mod.Source), 0644)

	fmt.Println(mod.Language)

	if mod.Language == "C#" {
		cmdHandle = exec.Command("/bin/sh", "-c", "mcs -out:" + outfile + " " + filename)
	} else if mod.Language == "Go" {
		cmdHandle = exec.Command("/bin/sh", "-c", "env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '-s -w' -o " + outfile + " modules/" + mod.Name + ".go")
	}

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