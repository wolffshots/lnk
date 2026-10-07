package fs

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yarlson/lnk/internal/testenv"
)

func TestGetRelativePathUsesForwardSlashes(t *testing.T) {
	home := t.TempDir()
	testenv.SetHome(t, home)

	got, err := GetRelativePath(filepath.Join(home, ".config", "app", "config.json"))

	require.NoError(t, err)
	require.Equal(t, ".config/app/config.json", got)
}
