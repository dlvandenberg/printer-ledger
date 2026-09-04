package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/store"
	"github.com/dlvandenberg/printer-ledger/internal/tui"
)

const dbEnvVar = "PRINTER_LEDGER_DB"

func main() {
	dbFlag := flag.String("db", "", "path to the ledger database file (default: XDG data directory)")
	flag.Parse()

	dbPath, err := resolveDBPath(*dbFlag)
	if err == nil {
		err = run(dbPath)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "printer-ledger:", err)
		os.Exit(1)
	}
}

func resolveDBPath(dbFlag string) (string, error) {
	if dbFlag != "" {
		return dbFlag, nil
	}
	if env := os.Getenv(dbEnvVar); env != "" {
		return env, nil
	}
	return store.DefaultPath()
}

func run(dbPath string) error {
	created := !exists(dbPath)

	st, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	//nolint:errcheck
	defer st.Close()

	if created {
		fmt.Fprintln(os.Stderr, "printer-ledger: created", dbPath)
	}

	return tui.Run(app.New(st))
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, fs.ErrNotExist)
}
