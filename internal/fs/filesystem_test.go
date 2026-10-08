package fs

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
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

func TestMapRenameErrorCrossDevice(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows reports a different error code, see filesystem_windows_test.go")
	}

	renameErr := &os.LinkError{Op: "rename", Old: "/mnt/c/a", New: "/home/user/a", Err: syscall.EXDEV}

	err := mapRenameError(renameErr, "/mnt/c/a")

	require.ErrorIs(t, err, ErrCrossDevice)
}

func TestMapRenameErrorKeepsOtherErrors(t *testing.T) {
	other := errors.New("other failure")

	require.NoError(t, mapRenameError(nil, "a"))
	require.Equal(t, other, mapRenameError(other, "a"))
	require.Equal(t, other, mapSymlinkError(other))
}
