package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/scbrown/tapestry/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "View and manage Tapestry configuration.\nConfiguration is stored in TOML format at ~/.config/tapestry/config.toml.",
	}

	cmd.AddCommand(
		newConfigShowCmd(),
		newConfigGetCmd(),
		newConfigSetCmd(),
		newConfigPathCmd(),
		newConfigInitCmd(),
	)

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			path, _ := config.Path()
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Config: %s\n\n", path)
			_, _ = fmt.Fprintf(w, "Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
			_, _ = fmt.Fprintf(w, "Dolt:   %s@%s:%d\n", cfg.Dolt.User, cfg.Dolt.Host, cfg.Dolt.Port)
			_, _ = fmt.Fprintf(w, "\nWorkspaces:\n")
			if len(cfg.Workspace) == 0 {
				_, _ = fmt.Fprintln(w, "  (none configured)")
			}
			for _, ws := range cfg.Workspace {
				_, _ = fmt.Fprintf(w, "  %s: %s\n", ws.Name, ws.Path)
				for _, db := range ws.Databases {
					_, _ = fmt.Fprintf(w, "    - %s\n", db)
				}
			}
			return nil
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), path)
			return err
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long:  "Get a specific configuration value by dotted key.\nSupported keys: server.host, server.port, dolt.host, dolt.port, dolt.user, dolt.password.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			val, err := configGet(cfg, args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), val)
			return err
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  "Set a specific configuration value by dotted key.\nSupported keys: server.host, server.port, dolt.host, dolt.port, dolt.user, dolt.password.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := configSet(&cfg, args[0], args[1]); err != nil {
				return err
			}

			if err := config.Save(cfg); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s\n", args[0], args[1])
			return err
		},
	}
}

func configGet(cfg config.Config, key string) (string, error) {
	switch key {
	case "server.host":
		return cfg.Server.Host, nil
	case "server.port":
		return fmt.Sprintf("%d", cfg.Server.Port), nil
	case "dolt.host":
		return cfg.Dolt.Host, nil
	case "dolt.port":
		return fmt.Sprintf("%d", cfg.Dolt.Port), nil
	case "dolt.user":
		return cfg.Dolt.User, nil
	case "dolt.password":
		return cfg.Dolt.Password, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

func configSet(cfg *config.Config, key, value string) error {
	switch key {
	case "server.host":
		cfg.Server.Host = value
	case "server.port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
		cfg.Server.Port = port
	case "dolt.host":
		cfg.Dolt.Host = value
	case "dolt.port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
		cfg.Dolt.Port = port
	case "dolt.user":
		cfg.Dolt.User = value
	case "dolt.password":
		cfg.Dolt.Password = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a default config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("config already exists at %s", path)
			}

			cfg := config.DefaultConfig()
			cfg.Workspace = []config.WorkspaceConfig{
				{
					Name:      "homelab",
					Path:      os.ExpandEnv("$HOME/gt"),
					Databases: []string{"beads_aegis"},
				},
			}

			if err := config.Save(cfg); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", path)
			return nil
		},
	}
}
