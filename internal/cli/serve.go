package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var (
		host string
		port int
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the web dashboard",
		Long:  "Start the Tapestry web dashboard server.\nServes the HTMX-based dashboard for browsing beads, events, and agent activity.",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "tapestry server starting on %s:%d\n", host, port)
			if err != nil {
				return err
			}
			// TODO: start HTTP server
			return nil
		},
	}

	cmd.Flags().StringVar(&host, "host", "localhost", "host to listen on")
	cmd.Flags().IntVar(&port, "port", 8070, "port to listen on")

	return cmd
}
