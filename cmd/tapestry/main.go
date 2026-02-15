package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Println("tapestry — the archivist for your agent fleet")
	fmt.Println("TODO: implement CLI (serve, config, workspace)")
	return nil
}
