package creator

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreate_GeneratesFileWithSanitizedName(t *testing.T) {
	dir := t.TempDir()

	path, err := Create(dir, "add users-table")
	require.NoError(t, err)
	require.FileExists(t, path)

	// content equals embedded template
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, tmpl, string(data))

	// filename sanity check
	base := filepath.Base(path)
	require.True(t, strings.HasSuffix(base, "_add_users_table.sql"))

	prefix := strings.TrimSuffix(base, "_add_users_table.sql")
	require.Len(t, prefix, 14) // YYYYMMDDHHMMSS
	require.True(t, regexp.MustCompile(`^\d{14}$`).MatchString(prefix))
}

func TestCreate_EmptyNameReturnsError(t *testing.T) {
	_, err := Create(t.TempDir(), "   ")
	require.Error(t, err)
}
