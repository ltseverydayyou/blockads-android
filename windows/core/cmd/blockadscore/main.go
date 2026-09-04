package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	core "github.com/ltseverydayyou/blockads-windows-core"
)

type server struct{ m *core.Manager }
type apiError struct {
	Error string `json:"error"`
}
type enabledPayload struct {
	Enabled bool `json:"enabled"`
}
type customFilterPayload struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type rulePayload struct {
	Rule string `json:"rule"`
}
type startPayload struct {
	SystemDNS bool `json:"systemDns"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decode(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func fail(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true, "version": s.m.Status().Version})
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.m.Status()) })
	mux.HandleFunc("GET /settings", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.m.Settings()) })
	mux.HandleFunc("PUT /settings", func(w http.ResponseWriter, r *http.Request) {
		var p core.Settings
		if err := decode(r, &p); err != nil {
			fail(w, err)
			return
		}
		if err := s.m.ApplySettings(p); err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, 200, s.m.Settings())
	})
	mux.HandleFunc("POST /protection/start", func(w http.ResponseWriter, r *http.Request) {
		p := startPayload{SystemDNS: true}
		if r.ContentLength > 0 {
			if err := decode(r, &p); err != nil {
				fail(w, err)
				return
			}
		}
		if err := s.m.Start(p.SystemDNS); err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, 200, s.m.Status())
	})
	mux.HandleFunc("POST /protection/stop", func(w http.ResponseWriter, r *http.Request) {
		if err := s.m.Stop(true); err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, 200, s.m.Status())
	})
	mux.HandleFunc("GET /filters", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.m.Filters()) })
	mux.HandleFunc("POST /filters/update", func(w http.ResponseWriter, r *http.Request) {
		n, err := s.m.UpdateFilters()
		if err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"rules": n, "filters": s.m.Filters()})
	})
	mux.HandleFunc("POST /filters/custom", func(w http.ResponseWriter, r *http.Request) {
		var p customFilterPayload
		if err := decode(r, &p); err != nil {
			fail(w, err)
			return
		}
		f, err := s.m.AddCustomFilter(p.Name, p.URL)
		if err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, f)
	})
	mux.HandleFunc("PUT /filters/{id}", func(w http.ResponseWriter, r *http.Request) {
		var p enabledPayload
		if err := decode(r, &p); err != nil {
			fail(w, err)
			return
		}
		if err := s.m.SetFilterEnabled(r.PathValue("id"), p.Enabled); err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, 200, s.m.Filters())
	})
	mux.HandleFunc("DELETE /filters/{id}", func(w http.ResponseWriter, r *http.Request) {
		if err := s.m.RemoveCustomFilter(r.PathValue("id")); err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, 200, s.m.Filters())
	})
	mux.HandleFunc("GET /profiles", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.m.Profiles()) })
	mux.HandleFunc("POST /profiles/{id}/activate", func(w http.ResponseWriter, r *http.Request) {
		if err := s.m.ActivateProfile(r.PathValue("id")); err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"settings": s.m.Settings(), "filters": s.m.Filters()})
	})
	mux.HandleFunc("GET /rules", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.m.Rules()) })
	mux.HandleFunc("POST /rules", func(w http.ResponseWriter, r *http.Request) {
		var p rulePayload
		if err := decode(r, &p); err != nil {
			fail(w, err)
			return
		}
		x, err := s.m.AddRule(p.Rule)
		if err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, x)
	})
	mux.HandleFunc("PUT /rules/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			fail(w, err)
			return
		}
		var p enabledPayload
		if err := decode(r, &p); err != nil {
			fail(w, err)
			return
		}
		if err := s.m.SetRuleEnabled(id, p.Enabled); err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, 200, s.m.Rules())
	})
	mux.HandleFunc("DELETE /rules/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			fail(w, err)
			return
		}
		if err := s.m.DeleteRule(id); err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, 200, s.m.Rules())
	})
	mux.HandleFunc("GET /logs", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.m.Logs()) })
	mux.HandleFunc("DELETE /logs", func(w http.ResponseWriter, r *http.Request) {
		if err := s.m.ClearLogs(); err != nil {
			fail(w, err)
			return
		}
		writeJSON(w, 200, []core.LogEntry{})
	})
	mux.HandleFunc("GET /data-dir", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]string{"path": s.m.DataDir()})
	})
	mux.HandleFunc("POST /shutdown", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]bool{"ok": true})
		go func() {
			time.Sleep(100 * time.Millisecond)
			_ = s.m.Shutdown(true)
			os.Exit(0)
		}()
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1:") && !strings.HasPrefix(r.RemoteAddr, "[::1]:") {
			http.Error(w, "loopback only", 403)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func main() {
	if !ensureElevated() {
		return
	}
	startDNSCleanupWatchdog()
	m, err := core.NewManager()
	if err != nil {
		log.Fatal(err)
	}
	if m.Settings().ProtectionEnabled {
		if err := m.Start(true); err != nil {
			log.Printf("restore protection state: %v", err)
		}
	}
	s := &server{m: m}
	srv := &http.Server{Addr: "127.0.0.1:8754", Handler: s.routes(), ReadHeaderTimeout: 5 * time.Second}
	fmt.Println("BlockAds Windows core listening on http://127.0.0.1:8754")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
