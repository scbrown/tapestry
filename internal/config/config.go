// Package config manages tapestry configuration via TOML.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the top-level tapestry configuration.
type Config struct {
	Server    ServerConfig      `toml:"server"`
	Dolt      DoltConfig        `toml:"dolt"`
	Reactor   ReactorConfig     `toml:"reactor"`
	Workspace []WorkspaceConfig `toml:"workspace"`
	Repos     map[string]string `toml:"repos,omitempty"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

// DoltConfig holds Dolt connection settings.
type DoltConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password,omitempty"`
}

// ReactorConfig holds reactor SSE connection settings.
type ReactorConfig struct {
	URL string `toml:"url"` // e.g. "http://dolt.lan:8075/events/stream"
}

// WorkspaceConfig describes a Gas Town workspace to monitor.
type WorkspaceConfig struct {
	Name      string   `toml:"name"`
	Path      string   `toml:"path"`
	Databases []string `toml:"databases"`
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8070,
		},
		Dolt: DoltConfig{
			Host: "127.0.0.1",
			Port: 3306,
			User: "root",
		},
	}
}

// Path returns the config file path (~/.config/tapestry/config.toml).
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, ".config", "tapestry", "config.toml"), nil
}

// Load reads config from disk, falling back to defaults for missing values.
func Load() (Config, error) {
	cfg := DefaultConfig()

	path, err := Path()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// Save writes config to disk.
func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create config: %w", err)
	}

	enc := toml.NewEncoder(f)
	if err := enc.Encode(cfg); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
