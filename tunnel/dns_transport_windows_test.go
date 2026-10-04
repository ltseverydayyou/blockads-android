//go:build windows

package tunnel

import (
	"os"
	"testing"
)

func testTunPair(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	t.Skip("Android TUN socketpair transport is not available on Windows")
	return nil, nil
}
