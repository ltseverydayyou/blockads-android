//go:build !windows

package blockadswin

import (
	"errors"

	tunnel "github.com/nqmgaming/blockads-tunnel"
)

func startWindowsFullTunnel(e *tunnel.Engine) (func() error, error) {
	return nil, errors.New("Windows full tunnel is only available on Windows")
}
