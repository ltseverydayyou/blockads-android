package blockadswin

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tunnel "github.com/nqmgaming/blockads-tunnel"
)

func (m *Manager) syncFilters() error {
	var remote []remoteFilter
	resp, err := m.client.Get(filterMetadataURL)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("filter metadata HTTP %d", resp.StatusCode)
		} else {
			err = json.NewDecoder(resp.Body).Decode(&remote)
		}
	}
	if err != nil {
		log.Printf("remote filter catalog unavailable, using bundled catalog: %v", err)
		remote = nil
	}
	remote = mergeCatalog(remote)
	if len(remote) == 0 {
		return errors.New("filter catalog is empty")
	}
	return m.applyCatalog(remote)
}

func (m *Manager) download(url, path string, force bool) error {
	if strings.HasPrefix(url, "local://") {
		if _, err := os.Stat(path); err != nil {
			return err
		}
		return nil
	}
	if !force {
		if st, err := os.Stat(path); err == nil && st.Size() > 0 {
			return nil
		}
	}
	resp, err := m.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(tmp, path)
}

const dnsFilterCompilerVersion = "2"

func (m *Manager) ensureSafeLocalDNSFilter(f *FilterList, force bool) error {
	sourceURL := strings.TrimSpace(f.OriginalURL)
	if sourceURL == "" {
		sourceURL = strings.TrimSpace(f.URL)
	}
	if sourceURL == "" {
		return errors.New("filter has no original source URL")
	}

	triePath := m.filterPath(*f, ".trie")
	bloomPath := m.filterPath(*f, ".bloom")
	markerPath := filepath.Join(m.filtersDir, f.ID+".dns-compiler")

	if !force {
		marker, markerErr := os.ReadFile(markerPath)
		trieInfo, trieErr := os.Stat(triePath)
		bloomInfo, bloomErr := os.Stat(bloomPath)
		if markerErr == nil && strings.TrimSpace(string(marker)) == dnsFilterCompilerVersion &&
			trieErr == nil && trieInfo.Size() > 0 && bloomErr == nil && bloomInfo.Size() > 0 {
			f.TrieURL = "local://" + filepath.Base(triePath)
			f.BloomURL = "local://" + filepath.Base(bloomPath)
			return nil
		}
	}

	tmp := filepath.Join(m.dataDir, "compile_builtin_"+f.ID+".txt")
	defer os.Remove(tmp)
	if err := m.download(sourceURL, tmp, true); err != nil {
		return err
	}
	revision := strconv.FormatInt(time.Now().UnixNano(), 36)
	if force || filepath.Base(triePath) != f.ID+".trie" {
		triePath = filepath.Join(m.filtersDir, f.ID+".rev-"+revision+".trie")
		bloomPath = filepath.Join(m.filtersDir, f.ID+".rev-"+revision+".bloom")
	}
	count, err := tunnel.CompileFilterList(tmp, triePath, bloomPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(markerPath, []byte(dnsFilterCompilerVersion+"\n"), 0644); err != nil {
		return err
	}
	f.TrieURL = "local://" + filepath.Base(triePath)
	f.BloomURL = "local://" + filepath.Base(bloomPath)
	f.RuleCount = count
	return nil
}

func (m *Manager) loadEnabledFilters(force bool) (int, error) {
	m.filterMu.Lock()
	defer m.filterMu.Unlock()
	m.mu.RLock()
	filters := append([]FilterList(nil), m.filters...)
	eng := m.engine
	m.mu.RUnlock()
	var adTries, secTries, adBlooms, secBlooms []string
	var cssParts, scriptParts []string
	total := 0
	enabledFilters := 0
	loadedFilters := 0
	for i := range filters {
		f := &filters[i]
		if !f.Enabled {
			continue
		}
		enabledFilters++
		if err := m.ensureSafeLocalDNSFilter(f, force); err != nil {
			log.Printf("filter %s safe local compile: %v", f.Name, err)
			if cachedErr := m.ensureSafeLocalDNSFilter(f, false); cachedErr != nil {
				continue
			}
		}
		if f.TrieURL == "" || f.BloomURL == "" {
			log.Printf("filter %s has no compiled trie/bloom metadata", f.Name)
			continue
		}
		triePath := m.filterPath(*f, ".trie")
		bloomPath := m.filterPath(*f, ".bloom")
		if err := m.download(f.TrieURL, triePath, force); err != nil {
			log.Printf("filter %s trie: %v", f.Name, err)
			continue
		}
		if err := m.download(f.BloomURL, bloomPath, force); err != nil {
			log.Printf("filter %s bloom: %v", f.Name, err)
			continue
		}
		if f.Category == "SECURITY" {
			secTries = append(secTries, triePath)
			secBlooms = append(secBlooms, bloomPath)
		} else {
			adTries = append(adTries, triePath)
			adBlooms = append(adBlooms, bloomPath)
		}
		if f.CSSURL != "" {
			p := filepath.Join(m.filtersDir, f.ID+".css")
			if m.download(f.CSSURL, p, force) == nil {
				if b, e := os.ReadFile(p); e == nil {
					cssParts = append(cssParts, string(b))
				}
			}
		}
		if f.ScriptletsURL != "" {
			p := filepath.Join(m.filtersDir, f.ID+".scriptlets")
			if m.download(f.ScriptletsURL, p, force) == nil {
				if b, e := os.ReadFile(p); e == nil {
					scriptParts = append(scriptParts, string(b))
				}
			}
		}
		f.LastUpdated = time.Now().UnixMilli()
		loadedFilters++
		total += f.RuleCount
	}
	if enabledFilters > 0 && loadedFilters == 0 {
		return total, fmt.Errorf("no enabled filter lists could be loaded; existing protection was preserved")
	}
	m.mu.Lock()
	m.filters = filters
	m.mu.Unlock()
	_ = m.saveFilters()
	if eng != nil {
		eng.SetTries(strings.Join(adTries, ","), strings.Join(secTries, ","), strings.Join(adBlooms, ","), strings.Join(secBlooms, ","))
		eng.SetCosmeticCSS(strings.Join(cssParts, "\n"))
		eng.SetScriptletRules(strings.Join(scriptParts, "\n"))
		m.cleanupRetiredFilters(filters)
	}
	return total, nil
}

func (m *Manager) activateProfile(id string, reload bool) error {
	var p *Profile
	for _, x := range presetProfiles() {
		if strings.EqualFold(x.ID, id) || strings.EqualFold(x.Type, id) || strings.EqualFold(x.Name, id) {
			v := x
			p = &v
			break
		}
	}
	if p == nil {
		return errors.New("profile not found")
	}
	set := map[string]bool{}
	for _, u := range p.EnabledFilterURLs {
		set[u] = true
	}
	m.mu.Lock()
	for i := range m.filters {
		u := m.filters[i].OriginalURL
		if u == "" {
			u = m.filters[i].URL
		}
		m.filters[i].Enabled = set[u]
	}
	m.settings.SafeSearchEnabled = p.SafeSearchEnabled
	m.settings.YouTubeRestrictedMode = p.YouTubeRestrictedMode
	m.settings.ActiveProfile = p.ID
	e := m.engine
	m.mu.Unlock()
	_ = m.saveFilters()
	_ = m.saveSettings()
	if e != nil {
		e.SetSafeSearch(p.SafeSearchEnabled)
		e.SetYouTubeRestricted(p.YouTubeRestrictedMode)
		if reload {
			_, _ = m.loadEnabledFilters(false)
		}
	}
	return nil
}

func (m *Manager) addCustomFilter(name, url string) (FilterList, error) {
	name = strings.TrimSpace(name)
	url = strings.TrimSpace(url)
	if name == "" || url == "" {
		return FilterList{}, errors.New("name and url are required")
	}
	id := "custom_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	f := FilterList{ID: id, Name: name, URL: url, OriginalURL: url, Description: "Custom filter", Enabled: true, Category: "AD", LastUpdated: time.Now().UnixMilli()}
	if err := m.compileViaBackend(&f); err != nil {
		log.Printf("custom compiler API assets unavailable: %v", err)
	}
	// Always rebuild trie/bloom locally with the DNS-safe compiler. The backend
	// may still provide cosmetic/scriptlet assets, but its DNS trie must not
	// globalize contextual EasyList rules.
	if err := m.compileLocally(&f); err != nil {
		return FilterList{}, err
	}
	m.mu.Lock()
	m.filters = append(m.filters, f)
	m.settings.ActiveProfile = "CUSTOM"
	m.mu.Unlock()
	_ = m.saveFilters()
	_ = m.saveSettings()
	if m.engine != nil {
		_, _ = m.loadEnabledFilters(false)
	}
	return f, nil
}

func (m *Manager) compileViaBackend(f *FilterList) error {
	body, _ := json.Marshal(map[string]string{"url": f.URL})
	req, _ := http.NewRequest(http.MethodPost, compilerURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("compiler HTTP %d", resp.StatusCode)
	}
	var cr compilerResponse
	if err = json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return err
	}
	if cr.Status != "success" || cr.DownloadURL == "" {
		return errors.New("compiler did not return success")
	}
	zresp, err := m.client.Get(cr.DownloadURL)
	if err != nil {
		return err
	}
	defer zresp.Body.Close()
	zb, err := io.ReadAll(io.LimitReader(zresp.Body, 256<<20))
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(bytes.NewReader(zb), int64(len(zb)))
	if err != nil {
		return err
	}
	foundTrie, foundBloom := false, false
	for _, zf := range zr.File {
		ext := strings.ToLower(filepath.Ext(zf.Name))
		if ext != ".trie" && ext != ".bloom" && ext != ".css" && ext != ".scriptlets" {
			continue
		}
		rc, e := zf.Open()
		if e != nil {
			return e
		}
		out := filepath.Join(m.filtersDir, f.ID+ext)
		wf, e := os.Create(out)
		if e != nil {
			rc.Close()
			return e
		}
		_, e = io.Copy(wf, rc)
		wf.Close()
		rc.Close()
		if e != nil {
			return e
		}
		switch ext {
		case ".trie":
			foundTrie = true
			f.TrieURL = "local://" + f.ID + ext
		case ".bloom":
			foundBloom = true
			f.BloomURL = "local://" + f.ID + ext
		case ".css":
			f.CSSURL = "local://" + f.ID + ext
		case ".scriptlets":
			f.ScriptletsURL = "local://" + f.ID + ext
		}
	}
	if !foundTrie || !foundBloom {
		return errors.New("compiler archive missing trie/bloom")
	}
	f.RuleCount = cr.RuleCount
	return nil
}

func (m *Manager) compileLocally(f *FilterList) error {
	tmp := filepath.Join(m.dataDir, "compile_"+f.ID+".txt")
	defer os.Remove(tmp)
	if err := m.download(f.URL, tmp, true); err != nil {
		return err
	}
	trie := filepath.Join(m.filtersDir, f.ID+".trie")
	bloom := filepath.Join(m.filtersDir, f.ID+".bloom")
	count, err := tunnel.CompileFilterList(tmp, trie, bloom)
	if err != nil {
		return err
	}
	f.TrieURL = "local://" + filepath.Base(trie)
	f.BloomURL = "local://" + filepath.Base(bloom)
	f.RuleCount = count
	return nil
}
