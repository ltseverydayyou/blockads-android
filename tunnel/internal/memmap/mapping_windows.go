//go:build windows

package memmap

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"unsafe"
)

func mapHandle(fd windows.Handle, size int) ([]byte, error) {
	if size <= 0 {
		return nil, fmt.Errorf("invalid mapping size %d", size)
	}
	mapping, err := windows.CreateFileMapping(fd, nil, windows.PAGE_READONLY, 0, 0, nil)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(mapping)
	addr, err := windows.MapViewOfFile(mapping, windows.FILE_MAP_READ, 0, 0, uintptr(size))
	if err != nil {
		return nil, err
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(addr)), size), nil
}

func ReadOnly(f *os.File, size int) ([]byte, error) {
	return mapHandle(windows.Handle(f.Fd()), size)
}

func Unmap(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return windows.UnmapViewOfFile(uintptr(unsafe.Pointer(&data[0])))
}

func Region(fd int, offset, length int64) ([]byte, func(), error) {
	if length <= 0 {
		return nil, func() {}, nil
	}
	if offset < 0 || offset+length < offset {
		return nil, nil, fmt.Errorf("invalid mapping region")
	}
	data, err := mapHandle(windows.Handle(uintptr(fd)), int(offset+length))
	if err != nil {
		return nil, nil, err
	}
	return data[offset : offset+length], func() { _ = Unmap(data) }, nil
}
