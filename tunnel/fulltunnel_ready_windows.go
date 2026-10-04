//go:build windows

package tunnel

func (e *Engine) IsFullTunnelReady() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running && e.tcpStack != nil
}
