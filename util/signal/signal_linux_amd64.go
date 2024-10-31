//go:build amd64 && linux

package signal

import "syscall"

const (
	SIGUSR1 = syscall.SIGUSR1
)
