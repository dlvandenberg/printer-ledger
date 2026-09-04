package store_test

import (
	"path/filepath"
	"testing"

	"github.com/dlvandenberg/printer-ledger/internal/store"
)

// The one seam is internal/app (CLAUDE.md), but path resolution has no reachable
// behaviour there: it reads the environment and must never touch the real ledger.
func TestDefaultPath(t *testing.T) {
	home := t.TempDir()
	data := t.TempDir()

	tests := []struct {
		name        string
		xdgDataHome string
		want        string
	}{
		{
			name:        "absolute XDG_DATA_HOME",
			xdgDataHome: data,
			want:        filepath.Join(data, "printer-ledger", "ledger.db"),
		},
		{
			name:        "relative XDG_DATA_HOME is ignored",
			xdgDataHome: "tmp/data",
			want:        filepath.Join(home, ".local", "share", "printer-ledger", "ledger.db"),
		},
		{
			name:        "unset XDG_DATA_HOME",
			xdgDataHome: "",
			want:        filepath.Join(home, ".local", "share", "printer-ledger", "ledger.db"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HOME", home)
			t.Setenv("XDG_DATA_HOME", tc.xdgDataHome)

			got, err := store.DefaultPath()
			if err != nil {
				t.Fatalf("DefaultPath() error = %v", err)
			}
			if got != tc.want {
				t.Errorf("DefaultPath() = %q, want %q", got, tc.want)
			}
		})
	}
}
