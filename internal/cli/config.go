package cli

import (
	"fmt"
	"os"

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
