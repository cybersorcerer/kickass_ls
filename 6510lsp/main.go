package main

import (
	"fmt"
	"os"
	log "github.com/c64-lsp/6510lsp/internal/log"
	lsp "github.com/c64-lsp/6510lsp/internal/lsp"
)

func main() {
	if err := log.InitLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	log.Logger.Println("6510 Language Server started.")

	lsp.Start()

	fmt.Println("6510 Language Server is running. Check log file for details.")
}
