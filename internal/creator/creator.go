package creator

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed templates/migration.sql.tmpl
var tmpl string

// Create generates a timestamp-prefixed SQL migration file and
// returns its absolute path.
func Create(dir, rawName string) (string, error) {
	name := strings.TrimSpace(rawName)
	if name == "" {
		return "", fmt.Errorf("migration name must not be empty")
	}
	// Safe file name: replace spaces with underscores, keep alnum & _ only
	name = strings.Map(func(r rune) rune {
		if r == ' ' || r == '-' {
			return '_'
		}
		if r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' ||
			r >= 'a' && r <= 'z' || r == '_' {
			return r
		}
		return -1
	}, name)

	ts := time.Now().UTC().Format("20060102150405") // yyyymmddHHMMSS
	file := fmt.Sprintf("%s_%s.sql", ts, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	full := filepath.Join(dir, file)
	//nolint:gosec
	if err := os.WriteFile(full, []byte(tmpl), 0o644); err != nil {
		return "", err
	}
	return full, nil
}
