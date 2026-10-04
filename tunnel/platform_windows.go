//go:build windows

package tunnel

import (
	"golang.org/x/sys/windows"
)

func dupFD(fd int) (int, error) {
	current := windows.CurrentProcess()
	var dup windows.Handle
	if err := windows.DuplicateHandle(current, windows.Handle(uintptr(fd)), current, &dup, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
		return 0, err
	}
	return int(dup), nil
}

func setNonblock(fd int) error { return nil }
