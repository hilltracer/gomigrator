package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDir_OK(t *testing.T) {
	tmp := t.TempDir()
	fname := filepath.Join(tmp, "20250713101010_init.sql")
	sql := `
-- +gomigrator Up
CREATE TABLE qwe(id INT);
-- +gomigrator Down
DROP TABLE qwe;`
	require.NoError(t, os.WriteFile(fname, []byte(sql), 0o644))

	got, err := ParseDir(tmp)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, int64(20250713101010), got[0].Version)
	require.Equal(t, "init", got[0].Name)
	require.Equal(t, "CREATE TABLE qwe(id INT);", got[0].UpSQL)
	require.Equal(t, "DROP TABLE qwe;", got[0].DownSQL)
}

func TestParseDir_BadFilename(t *testing.T) {
	tmp := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(tmp, "badname.sql"), []byte(""), 0o644))

	_, err := ParseDir(tmp)
	require.Error(t, err)
}
