package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/scbrown/tapestry/internal/config"
)

func newWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"ws"},
		Short:   "Manage workspaces",
		Long:    "List and manage Gas Town workspaces.\nWorkspaces define which Dolt databases and event logs Tapestry monitors.",
	}

	cmd.AddCommand(
		newWorkspaceListCmd(),
		newWorkspaceAddCmd(),
		newWorkspaceRemoveCmd(),
	)

	return cmd
}

func newWorkspaceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			if len(cfg.Workspace) == 0 {
				_, _ = fmt.Fprintln(w, "No workspaces configured.")
				_, err = fmt.Fprintln(w, "Use 'tapestry workspace add' to add one.")
				return err
			}
			for _, ws := range cfg.Workspace {
				_, _ = fmt.Fprintf(w, "%s: %s\n", ws.Name, ws.Path)
				for _, db := range ws.Databases {
					_, _ = fmt.Fprintf(w, "  - %s\n", db)
				}
			}
			return nil
		},
	}
}

func newWorkspaceAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <path>",
		Short: "Add a workspace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, wsPath := args[0], args[1]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			for _, ws := range cfg.Workspace {
				if ws.Name == name {
					return fmt.Errorf("workspace %q already exists", name)
				}
			}

			cfg.Workspace = append(cfg.Workspace, config.WorkspaceConfig{
				Name: name,
				Path: wsPath,
			})

			if err := config.Save(cfg); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Added workspace %q at %s\n", name, wsPath)
			return err
		},
	}
}

func newWorkspaceRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			found := false
			filtered := cfg.Workspace[:0]
			for _, ws := range cfg.Workspace {
				if ws.Name == name {
					found = true
					continue
				}
				filtered = append(filtered, ws)
			}
			if !found {
				return fmt.Errorf("workspace %q not found", name)
			}

			cfg.Workspace = filtered
			if err := config.Save(cfg); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed workspace %q\n", name)
			return err
		},
	}
}
