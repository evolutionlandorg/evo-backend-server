//go:build darwin

package signal

import "syscall"

const (
	SIGUSR1 = syscall.SIGUSR1
)
