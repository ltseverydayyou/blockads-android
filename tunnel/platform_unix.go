//go:build !windows

package tunnel

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func mmapReadOnly(f *os.File, size int) ([]byte, error) {
	return unix.Mmap(int(f.Fd()), 0, size, unix.PROT_READ, unix.MAP_SHARED)
}

func munmap(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return unix.Munmap(data)
}

func dupFD(fd int) (int, error) {
	return syscall.Dup(fd)
}
