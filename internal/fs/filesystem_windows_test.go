package fs

import (
	"testing"

	"github.com/stretchr/testify/require"

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
