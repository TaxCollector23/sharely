//go:build darwin || linux

package main

import "syscall"

func daemonProcessAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
