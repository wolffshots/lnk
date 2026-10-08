package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandWildcards(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt", "c[1].txt", "notes.md"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0644))
	}
	in := func(name string) string { return filepath.Join(dir, name) }

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "no wildcard", args: []string{in("a.txt")}, want: []string{in("a.txt")}},
		{name: "star", args: []string{in("*.txt")}, want: []string{in("a.txt"), in("b.txt"), in("c[1].txt")}},
		{name: "question mark", args: []string{in("?.txt")}, want: []string{in("a.txt"), in("b.txt")}},
		{name: "bracket is literal", args: []string{in("c[1].*")}, want: []string{in("c[1].txt")}},
		{name: "no match keeps the argument", args: []string{in("*.json")}, want: []string{in("*.json")}},
		{name: "mixed arguments keep their order", args: []string{in("notes.md"), in("a*")}, want: []string{in("notes.md"), in("a.txt")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, expandWildcards(tt.args))
		})
	}
}
