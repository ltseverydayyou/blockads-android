package blockadswin

import (
	"context"
	"log"
	"time"
)

func (m *Manager) RunMonitor(ctx context.Context) {
	go m.trustedWatcher(ctx)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.mu.RLock()
			wanted, reconnect := m.settings.ProtectionEnabled, m.settings.AutoReconnect
			running, starting := m.running, m.engine != nil && !m.running
			m.mu.RUnlock()
			if !wanted || starting {
				continue
			}
			if running {
				if err := verifyDNS(53); err == nil {
					continue
				} else {
					log.Printf("DNS health check: %v", err)
				}
				_ = m.shutdown(true)
			}
			if reconnect {
				if err := m.start(true); err != nil {
					log.Printf("DNS reconnect: %v", err)
				}
			}
		}
	}
}
