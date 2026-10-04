package blockadswin

import (
	"context"
	"time"
)

func (m *Manager) startAutoUpdater() {
	m.mu.Lock()
	if m.autoUpdateStop != nil {
		close(m.autoUpdateStop)
	}
	stop := make(chan struct{})
	m.autoUpdateStop = stop
	s := m.settings
	m.mu.Unlock()
	if !s.AutoUpdateEnabled || s.AutoUpdateFrequency == "manual" {
		return
	}
	dur := 24 * time.Hour
	switch s.AutoUpdateFrequency {
	case "6h":
		dur = 6 * time.Hour
	case "12h":
		dur = 12 * time.Hour
	case "48h":
		dur = 48 * time.Hour
	}
	go func() {
		t := time.NewTicker(dur)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				_ = m.syncFilters()
				_, _ = m.loadEnabledFilters(true)
			case <-stop:
				return
			}
		}
	}()
}

func (m *Manager) trustedWatcher(ctx context.Context) {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			m.mu.RLock()
			enabled := m.settings.PauseOnTrusted
			trusted := append([]string(nil), m.settings.TrustedSSIDs...)
			running := m.running
			paused := m.pausedTrusted
			delayOn := m.settings.NetworkSwitchDelayEnabled
			delay := m.settings.NetworkSwitchDelaySec
			m.mu.RUnlock()
			if !enabled {
				continue
			}
			ssid := currentSSID()
			hit := false
			for _, s := range trusted {
				if s == ssid && ssid != "" {
					hit = true
					break
				}
			}
			if hit && running && !paused {
				if m.stop(true) == nil {
					m.mu.Lock()
					m.pausedTrusted = true
					m.mu.Unlock()
				}
			}
			if !hit && paused {
				if delayOn {
					time.Sleep(time.Duration(delay) * time.Second)
				}
				if m.start(true) == nil {
					m.mu.Lock()
					m.pausedTrusted = false
					m.mu.Unlock()
				}
			}
		}
	}
}
