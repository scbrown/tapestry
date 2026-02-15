package cli

import (
	"fmt"

	"github.com/spf13/cobra"
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
	)

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: load and display config
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "# ~/.config/tapestry/config.toml")
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "# (not yet configured)")
			return err
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: resolve actual config path
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "~/.config/tapestry/config.toml")
			return err
		},
	}
}
