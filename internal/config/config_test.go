package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.Host != "localhost" {
		t.Errorf("default host = %q, want localhost", cfg.Server.Host)
	}
	if cfg.Server.Port != 8070 {
		t.Errorf("default port = %d, want 8070", cfg.Server.Port)
	}
}

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	if err := os.MkdirAll(filepath.Join(dir, ".config", "tapestry"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	cfg.Workspace = []WorkspaceConfig{
		{Name: "test", Path: "/tmp/gt", Databases: []string{"beads_test"}},
	}

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Workspace) != 1 {
		t.Fatalf("workspaces = %d, want 1", len(loaded.Workspace))
	}
	if loaded.Workspace[0].Name != "test" {
		t.Errorf("workspace name = %q, want test", loaded.Workspace[0].Name)
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != 8070 {
		t.Errorf("port = %d, want 8070", cfg.Server.Port)
	}
}
