//go:build !windows

package memmap

import (
	"golang.org/x/sys/unix"
	"os"
)

func ReadOnly(f *os.File, size int) ([]byte, error) {
	return unix.Mmap(int(f.Fd()), 0, size, unix.PROT_READ, unix.MAP_SHARED)
}

func Unmap(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return unix.Munmap(data)
}

func Region(fd int, offset, length int64) ([]byte, func(), error) {
	if length <= 0 {
		return nil, func() {}, nil
	}
	pageSize := int64(os.Getpagesize())
	alignedOffset := offset / pageSize * pageSize
	diff := offset - alignedOffset
	data, err := unix.Mmap(fd, alignedOffset, int(length+diff), unix.PROT_READ, unix.MAP_SHARED)
	if err != nil {
		return nil, nil, err
	}
	return data[diff : diff+length], func() { _ = Unmap(data) }, nil
}
