//go:build windows

package tunnel

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

func mmapReadOnly(f *os.File, size int) ([]byte, error) {
	if size <= 0 {
		return []byte{}, nil
	}
	mapping, err := windows.CreateFileMapping(windows.Handle(f.Fd()), nil, windows.PAGE_READONLY, 0, 0, nil)
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

func munmap(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return windows.UnmapViewOfFile(uintptr(unsafe.Pointer(&data[0])))
}

func dupFD(fd int) (int, error) {
	current := windows.CurrentProcess()
	var dup windows.Handle
	if err := windows.DuplicateHandle(current, windows.Handle(uintptr(fd)), current, &dup, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
		return 0, err
	}
	return int(dup), nil
}
