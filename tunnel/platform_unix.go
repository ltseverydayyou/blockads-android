//go:build !windows

package tunnel

import "syscall"

func dupFD(fd int) (int, error) { return syscall.Dup(fd) }
func setNonblock(fd int) error  { return syscall.SetNonblock(fd, true) }
