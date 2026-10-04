package blockadswin

import (
	"fmt"
	"os"
	"path/filepath"
)

func prepareDataDirectory(base string) (string, error) {
	dir := filepath.Join(base, "BlockAdsData")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	marker := filepath.Join(dir, "legacy-migrated")
	if _, err := os.Stat(marker); err == nil {
		return dir, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	// Installer upgrades remove the app directory, including untracked files.
	legacy := filepath.Join(base, "BlockAds")
	for _, name := range []string{"settings.json", "filters.json", "rules.json", "dns_logs.json", "dns_backup.json", "remote_filters"} {
		if err := copyLegacyData(filepath.Join(legacy, name), filepath.Join(dir, name)); err != nil {
			return "", fmt.Errorf("migrate %s: %w", name, err)
		}
	}
	if err := os.WriteFile(marker, []byte("1"), 0600); err != nil {
		return "", err
	}
	return dir, nil
}

func copyLegacyData(source, destination string) error {
	info, err := os.Stat(source)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err := os.MkdirAll(destination, 0700); err != nil {
			return err
		}
		entries, err := os.ReadDir(source)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := copyLegacyData(filepath.Join(source, entry.Name()), filepath.Join(destination, entry.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	if !info.Mode().IsRegular() {
		return nil
	}
	if _, err := os.Stat(destination); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, content, 0600)
}
