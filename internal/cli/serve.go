package cli

import (
	"fmt"
	"net/http"

	"github.com/scbrown/tapestry/internal/dolt"
	"github.com/scbrown/tapestry/internal/web"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var (
		host     string
		port     int
		doltHost string
		doltPort int
		doltUser string
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the web dashboard",
		Long:  "Start the Tapestry web dashboard server.\nServes the HTMX-based dashboard for browsing beads, events, and agent activity.",
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := fmt.Sprintf("%s:%d", host, port)

			// Connect to Dolt (optional — server starts without it)
			var ds web.DataSource
			cfg := dolt.Config{Host: doltHost, Port: doltPort, User: doltUser}
			client, err := dolt.New(cfg)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: dolt config invalid: %v\n", err)
			} else {
				ds = client
				defer func() { _ = client.Close() }()
			}

			srv := web.New(ds)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "tapestry server listening on http://%s\n", addr)
			return http.ListenAndServe(addr, srv)
		},
	}

	cmd.Flags().StringVar(&host, "host", "localhost", "host to listen on")
	cmd.Flags().IntVar(&port, "port", 8070, "port to listen on")
	cmd.Flags().StringVar(&doltHost, "dolt-host", "127.0.0.1", "Dolt server host")
	cmd.Flags().IntVar(&doltPort, "dolt-port", 3306, "Dolt server port")
	cmd.Flags().StringVar(&doltUser, "dolt-user", "root", "Dolt server user")

	return cmd
}
