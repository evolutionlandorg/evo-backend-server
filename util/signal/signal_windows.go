//go:build amd64 && windows

package signal

import "syscall"

const (
	SIGUSR1 = syscall.Signal(0xa)
)
