package main

import (
	"os"
	"testing"

	"craftops/internal/cli"
)

func TestExecute_VersionAndHelp(t *testing.T) {
	t.Run("version flag", func(t *testing.T) {
		orig := os.Args
		defer func() { os.Args = orig }()
		os.Args = []string{"craftops", "--version"}
		if err := cli.Execute(); err != nil {
			t.Fatalf("cli.Execute with --version returned error: %v", err)
		}
	})

	t.Run("help default", func(t *testing.T) {
		orig := os.Args
		defer func() { os.Args = orig }()
		os.Args = []string{"craftops"}
		if err := cli.Execute(); err != nil {
			t.Fatalf("cli.Execute default help returned error: %v", err)
		}
	})
}

func TestMainEntrypoint(t *testing.T) {
    t.Run("default help", func(t *testing.T) {
        orig := os.Args
        defer func() { os.Args = orig }()
        os.Args = []string{"craftops"}
        main()
    })
    t.Run("version flag", func(t *testing.T) {
        orig := os.Args
        defer func() { os.Args = orig }()
        os.Args = []string{"craftops", "--version"}
        main()
    })
}
