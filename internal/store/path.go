package store

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	ledgerDirName  = "printer-ledger"
	ledgerFileName = "ledger.db"
)

// A relative XDG_DATA_HOME is invalid per the XDG spec and must be ignored; honouring
// that is what stops the ledger resolving against the working directory (ADR-0009).
func DefaultPath() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); filepath.IsAbs(dir) {
		return filepath.Join(dir, ledgerDirName, ledgerFileName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", ledgerDirName, ledgerFileName), nil
}
