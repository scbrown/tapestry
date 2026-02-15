package cli

import (
	"fmt"

	"github.com/spf13/cobra"
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
			// TODO: load workspaces from config
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No workspaces configured.")
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Use 'tapestry workspace add' to add one.")
			return err
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
			// TODO: persist workspace to config
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Added workspace %q at %s\n", name, wsPath)
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
			// TODO: remove workspace from config
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Removed workspace %q\n", name)
			return err
		},
	}
}
