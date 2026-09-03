//go:build !windows

package blockadswin

import "errors"

func systemDNSServers() ([]string, error) {
	return nil, errors.New("system DNS discovery is only available on Windows")
}
