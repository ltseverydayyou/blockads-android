package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type ruleChecker struct{ m *Manager }
type dnsLogCallback struct{ m *Manager }

func normalizeDomain(d string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(d)), ".")
}
func matchesDomainSet(domain string, set map[string]struct{}) bool {
	d := normalizeDomain(domain)
	if _, ok := set[d]; ok {
		return true
	}
	for strings.Contains(d, ".") {
		d = strings.SplitN(d, ".", 2)[1]
		if _, ok := set[d]; ok {
			return true
		}
		if _, ok := set["*."+d]; ok {
			return true
		}
	}
	return false
}
func (r *ruleChecker) snapshots() (map[string]struct{}, map[string]struct{}) {
	allow, block := map[string]struct{}{}, map[string]struct{}{}
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	for _, x := range r.m.rules {
		if !x.Enabled {
			continue
		}
		switch x.RuleType {
		case "ALLOW":
			allow[normalizeDomain(x.Domain)] = struct{}{}
		case "BLOCK":
			block[normalizeDomain(x.Domain)] = struct{}{}
		}
	}
	return allow, block
}
func (r *ruleChecker) HasCustomRule(domain string) int {
	a, b := r.snapshots()
	if matchesDomainSet(domain, a) {
		return 0
	}
	if matchesDomainSet(domain, b) {
		return 1
	}
	return -1
}
func (r *ruleChecker) IsBlocked(domain string) bool { return r.HasCustomRule(domain) == 1 }
func (r *ruleChecker) GetBlockReason(domain string) string {
	if r.HasCustomRule(domain) == 1 {
		return "CUSTOM_RULE"
	}
	return ""
}

func (c *dnsLogCallback) OnDNSQuery(domain string, blocked bool, queryType int, responseTimeMS int64, appName, resolvedIP, blockedBy string) {
	c.m.mu.Lock()
	if !c.m.settings.RecordDNSLogs {
		c.m.mu.Unlock()
		return
	}
	id := c.m.ids.log.Add(1)
	c.m.logs = append(c.m.logs, LogEntry{ID: id, Timestamp: time.Now().UnixMilli(), Domain: domain, Blocked: blocked, QueryType: queryType, ResponseTimeMS: responseTimeMS, AppName: appName, ResolvedIP: resolvedIP, BlockedBy: blockedBy})
	if len(c.m.logs) > 10000 {
		c.m.logs = append([]LogEntry(nil), c.m.logs[len(c.m.logs)-10000:]...)
	}
	c.m.mu.Unlock()
	if id%20 == 0 {
		go c.m.saveLogs()
	}
}

func parseRuleText(text string, id int64) (Rule, error) {
	t := strings.TrimSpace(text)
	if t == "" {
		return Rule{}, errors.New("empty rule")
	}
	now := time.Now().UnixMilli()
	if strings.HasPrefix(t, "!") {
		return Rule{ID: id, Rule: t, RuleType: "COMMENT", Enabled: true, AddedTimestamp: now}, nil
	}
	typ := "BLOCK"
	raw := t
	if strings.HasPrefix(raw, "@@") {
		typ = "ALLOW"
		raw = strings.TrimPrefix(raw, "@@")
	}
	raw = strings.TrimPrefix(raw, "||")
	raw = strings.TrimSuffix(raw, "^")
	raw = normalizeDomain(raw)
	if raw == "" || strings.HasPrefix(raw, ".") || strings.HasSuffix(raw, ".") || strings.Contains(raw, "..") {
		return Rule{}, errors.New("invalid domain")
	}
	for _, lab := range strings.Split(raw, ".") {
		if lab == "*" {
			continue
		}
		if lab == "" {
			return Rule{}, errors.New("invalid domain")
		}
		for i, ch := range lab {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-') {
				return Rule{}, fmt.Errorf("invalid domain character at %d", i)
			}
		}
	}
	return Rule{ID: id, Rule: t, RuleType: typ, Domain: strings.ToLower(raw), Enabled: true, AddedTimestamp: now}, nil
}
