//go:build !windows

package tunnel

import (
	"golang.org/x/sys/unix"
	"os"
	"testing"
)

func testTunPair(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_DGRAM, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, fd := range fds {
		unix.CloseOnExec(fd)
		if err := unix.SetNonblock(fd, true); err != nil {
			t.Fatal(err)
		}
	}
	tun := os.NewFile(uintptr(fds[0]), "tun")
	peer := os.NewFile(uintptr(fds[1]), "peer")

	return tun, peer
}
