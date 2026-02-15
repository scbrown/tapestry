package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestServeCmd_Defaults(t *testing.T) {
	root := newRootCmd("test")
	root.AddCommand(newServeCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"serve"})

	if err := root.Execute(); err != nil {
		t.Fatalf("serve command failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "localhost:8070") {
		t.Errorf("serve output = %q, want to contain default address", got)
	}
}

func TestServeCmd_CustomHostPort(t *testing.T) {
	root := newRootCmd("test")
	root.AddCommand(newServeCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"serve", "--host", "0.0.0.0", "--port", "9090"})

	if err := root.Execute(); err != nil {
		t.Fatalf("serve command failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "0.0.0.0:9090") {
		t.Errorf("serve output = %q, want to contain custom address", got)
	}
}
