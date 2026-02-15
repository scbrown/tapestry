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

func TestConfigCmd_GetDefault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := newRootCmd("test")
	root.AddCommand(newConfigCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "get", "server.port"})

	if err := root.Execute(); err != nil {
		t.Fatalf("config get failed: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "8070" {
		t.Errorf("config get server.port = %q, want 8070", got)
	}
}

func TestConfigCmd_GetUnknown(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := newRootCmd("test")
	root.AddCommand(newConfigCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "get", "bogus.key"})

	err := root.Execute()
	if err == nil {
		t.Fatal("config get unknown key should fail")
	}
}

func TestConfigCmd_SetAndGet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	root := newRootCmd("test")
	root.AddCommand(newConfigCmd())

	// Set
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "set", "server.host", "0.0.0.0"})

	if err := root.Execute(); err != nil {
		t.Fatalf("config set failed: %v", err)
	}
	if !strings.Contains(buf.String(), "0.0.0.0") {
		t.Errorf("config set output = %q, want confirmation", buf.String())
	}

	// Get it back
	root2 := newRootCmd("test")
	root2.AddCommand(newConfigCmd())
	buf2 := new(bytes.Buffer)
	root2.SetOut(buf2)
	root2.SetErr(buf2)
	root2.SetArgs([]string{"config", "get", "server.host"})

	if err := root2.Execute(); err != nil {
		t.Fatalf("config get after set failed: %v", err)
	}
	got := strings.TrimSpace(buf2.String())
	if got != "0.0.0.0" {
		t.Errorf("config get server.host = %q, want 0.0.0.0", got)
	}
}

func TestConfigCmd_SetPort(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	root := newRootCmd("test")
	root.AddCommand(newConfigCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "set", "server.port", "9090"})

	if err := root.Execute(); err != nil {
		t.Fatalf("config set port failed: %v", err)
	}

	// Verify
	root2 := newRootCmd("test")
	root2.AddCommand(newConfigCmd())
	buf2 := new(bytes.Buffer)
	root2.SetOut(buf2)
	root2.SetErr(buf2)
	root2.SetArgs([]string{"config", "get", "server.port"})

	if err := root2.Execute(); err != nil {
		t.Fatalf("config get port failed: %v", err)
	}
	got := strings.TrimSpace(buf2.String())
	if got != "9090" {
		t.Errorf("config get server.port = %q, want 9090", got)
	}
}

func TestConfigCmd_SetInvalidPort(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := newRootCmd("test")
	root.AddCommand(newConfigCmd())

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"config", "set", "server.port", "notanumber"})

	if err := root.Execute(); err == nil {
		t.Fatal("config set invalid port should fail")
	}
}
