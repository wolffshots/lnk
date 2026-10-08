package fs

import (
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yarlson/lnk/internal/lnkerror"
	"github.com/yarlson/lnk/internal/testenv"
)

func TestGetRelativePathRejectsOutsideHome(t *testing.T) {
	testenv.SetHome(t, `C:\Users\someone`)

	tests := []struct {
		name string
		path string
	}{
		{name: "same volume", path: `C:\ProgramData\app\config.json`},
		{name: "sibling of home", path: `C:\Users\other\.bashrc`},
		{name: "different volume", path: `D:\dotfiles\.bashrc`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRelativePath(tt.path)

			require.ErrorIs(t, err, ErrOutsideHome)
			require.Empty(t, got)
		})
	}
}

func TestMapRenameError(t *testing.T) {
	tests := []struct {
		name  string
		errno syscall.Errno
		want  error
	}{
		{name: "different volume", errno: winErrNotSameDevice, want: ErrCrossDevice},
		{name: "access denied", errno: winErrAccessDenied, want: ErrFileInUse},
		{name: "sharing violation", errno: winErrSharingViolation, want: ErrFileInUse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renameErr := &os.LinkError{Op: "rename", Old: `C:\a`, New: `D:\a`, Err: tt.errno}

			err := mapRenameError(renameErr, `C:\a`)

			require.ErrorIs(t, err, tt.want)
			var lnkErr *lnkerror.Error
			require.ErrorAs(t, err, &lnkErr)
			require.Equal(t, `C:\a`, lnkErr.Path)
			require.NotEmpty(t, lnkErr.Suggestion)
		})
	}
}

func TestMapSymlinkErrorPrivilegeNotHeld(t *testing.T) {
	symlinkErr := &os.LinkError{Op: "symlink", Old: `..\a`, New: `C:\a`, Err: winErrPrivilegeNotHeld}

	err := mapSymlinkError(symlinkErr)

	require.ErrorIs(t, err, ErrSymlinkDenied)
	var lnkErr *lnkerror.Error
	require.ErrorAs(t, err, &lnkErr)
	require.Contains(t, lnkErr.Suggestion, "Developer Mode")
}

func TestCreateSymlinkRejectsDifferentVolume(t *testing.T) {
	err := New().CreateSymlink(`D:\repo\.bashrc`, `C:\Users\someone\.bashrc`)

	require.ErrorIs(t, err, ErrCrossDevice)
}
