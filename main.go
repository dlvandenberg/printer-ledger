package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dlvandenberg/printer-ledger/internal/app"
	"github.com/dlvandenberg/printer-ledger/internal/store"
	"github.com/dlvandenberg/printer-ledger/internal/tui"
)

func main() {
	dbPath := flag.String("db", store.DefaultPath, "path to the ledger database file")
	flag.Parse()

	if err := run(*dbPath); err != nil {
		fmt.Fprintln(os.Stderr, "printer-ledger:", err)
		os.Exit(1)
	}
}

func run(dbPath string) error {
	st, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	return tui.Run(app.New(st))
}
