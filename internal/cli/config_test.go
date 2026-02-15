package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfigCmd_Show(t *testing.T) {
	root := newRootCmd("test")
	root.AddCommand(newConfigCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "show"})

	if err := root.Execute(); err != nil {
		t.Fatalf("config show failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "config.toml") {
		t.Errorf("config show output = %q, want to mention config.toml", got)
	}
}

func TestConfigCmd_Path(t *testing.T) {
	root := newRootCmd("test")
	root.AddCommand(newConfigCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "path"})

	if err := root.Execute(); err != nil {
		t.Fatalf("config path failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "config.toml") {
		t.Errorf("config path output = %q, want to contain config file path", got)
	}
}
