package cli

import (
	"github.com/spf13/cobra"
)

// Execute runs the root command with the given version string.
func Execute(version string) error {
	root := newRootCmd(version)
	root.AddCommand(
		newServeCmd(),
		newConfigCmd(),
		newWorkspaceCmd(),
	)
	return root.Execute()
}

func newRootCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:     "tapestry",
		Short:   "The archivist for your agent fleet",
		Long:    "Tapestry is an archivist dashboard for Gas Town agent fleets.\nIt reads from Dolt-backed beads databases and Gas Town event logs\nto produce drillable retrospective views.",
		Version: version,
	}
}
