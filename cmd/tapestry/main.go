package main

import (
	"fmt"
	"os"

	"github.com/scbrown/tapestry/internal/cli"
)

// version is set by ldflags at build time.
var version = "dev"

func main() {
	if err := cli.Execute(version); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
