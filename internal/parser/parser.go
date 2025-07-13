package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Migration is an in-memory representation of one *.sql file.
type Migration struct {
	Version int64
	Name    string
	UpSQL   string
	DownSQL string
}

// ParseDir walks `dir` and returns all recognised migrations, sorted by Version.
func ParseDir(dir string) ([]Migration, error) {
	list, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, err
	}

	m := make([]Migration, 0, len(list))
	for _, f := range list {
		mig, err := parseFile(f)
		if err != nil {
			return nil, fmt.Errorf("file %s: %w", f, err)
		}
		m = append(m, mig)
	}

	sort.Slice(m, func(i, j int) bool { return m[i].Version < m[j].Version })
	return m, nil
}

func parseFile(path string) (Migration, error) {
	fn := filepath.Base(path) // 20250713190900_init.sql
	parts := strings.SplitN(fn, "_", 2)
	if len(parts) != 2 {
		return Migration{}, fmt.Errorf("filename must be <version>_<name>.sql")
	}
	ver, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Migration{}, fmt.Errorf("invalid version prefix: %w", err)
	}
	name := strings.TrimSuffix(parts[1], ".sql")

	f, err := os.Open(path)
	if err != nil {
		return Migration{}, err
	}
	defer f.Close()

	var (
		cur  *strings.Builder
		up   strings.Builder
		down strings.Builder
	)

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		switch strings.TrimSpace(line) {
		case "-- +gomigrator Up":
			cur = &up
			continue
		case "-- +gomigrator Down":
			cur = &down
			continue
		}
		if cur != nil {
			cur.WriteString(line)
			cur.WriteByte('\n')
		}
	}
	if err := sc.Err(); err != nil {
		return Migration{}, err
	}
	return Migration{
		Version: ver,
		Name:    name,
		UpSQL:   strings.TrimSpace(up.String()),
		DownSQL: strings.TrimSpace(down.String()),
	}, nil
}
