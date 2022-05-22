//go:build linux && !windows
// +build linux,!windows

package main

type ShellcodeInjector struct {
	shellcode []byte
	pid       int
}

func (si ShellcodeInjector) inject() error {
	//fmt.Println(si.shellcode)
	return nil
}

type RemoteShellcodeInjector struct {
	shellcode []byte
	pid       int
}

type RemotePipedShellcodeInjector struct {
	shellcode []byte
	args      string
}

func (si RemoteShellcodeInjector) inject() error {
	return nil
}

func (si RemotePipedShellcodeInjector) inject() string {
	return ""
}

const mimikatzShellcode = ""
