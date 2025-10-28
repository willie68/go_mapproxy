package configs

import (
	_ "embed"
	"strings"
)

//go:embed config.yaml
var ConfigFile string

//go:embed prefetch_blacklist.lst
var prefetchBlacklist string

func PrefetchBlacklist() []string {
	// Entfernt alle \r und konvertiert dann den String in ein Slice
	cleanStr := strings.ReplaceAll(prefetchBlacklist, "\r", "")
	lines := strings.Split(strings.TrimSpace(cleanStr), "\n")
	return lines
}
