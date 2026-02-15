package cli

import (
	"testing"
)

func TestServeCmd_Flags(t *testing.T) {
	cmd := newServeCmd()
	if cmd.Use != "serve" {
		t.Errorf("Use = %q, want serve", cmd.Use)
	}

	hostFlag := cmd.Flags().Lookup("host")
	if hostFlag == nil {
		t.Fatal("missing --host flag")
	}

	portFlag := cmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatal("missing --port flag")
	}
}
