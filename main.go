package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	log "c64.nvim/internal/log"
	lsp "c64.nvim/internal/lsp"
)

func main() {
	if f, err := os.OpenFile("6510lsp.log", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		f.Close()
	}

	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(os.Stderr)
	debug := flag.Bool("debug", false, "Enable debug logging")
	warnUnused := flag.Bool("warn-unused-labels", false, "Enable warnings for unused labels")

	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		log.Warn("Invalid command line argument: %v. Valid flags are: --debug, --warn-unused-labels", err)
	}

	if *debug {
		log.SetLevel(log.DEBUG)
	} else {
		log.SetLevel(log.INFO)
	}

	lsp.SetWarnUnusedLabels(*warnUnused)

	if err := log.InitLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(11)
	}

	log.Info("6510 Language Server started.")

	exePath, err := os.Executable()
	if err != nil {
		log.Error("Failed to get executable path: %v", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exePath)

	mnemonicPath := filepath.Join(exeDir, "mnemonic.json")
	kickassDir := exeDir

	lsp.Start(mnemonicPath, kickassDir)
}
