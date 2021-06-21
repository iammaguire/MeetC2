// +build linux,!windows

package main

import (
	"fmt"
)

type ShellcodeInjector struct {
	shellcode []byte
	pid int
}

func (si ShellcodeInjector) inject() error {
	fmt.Println(si.shellcode)
	return nil
}