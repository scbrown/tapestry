package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/scbrown/tapestry/internal/config"
	"github.com/scbrown/tapestry/internal/web"
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
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			// CLI flags override config
			if host != "" {
				cfg.Server.Host = host
			}
			if port != 0 {
				cfg.Server.Port = port
			}

			srv, err := web.New(cfg)
			if err != nil {
				return err
			}
			defer srv.Close()
			return srv.ListenAndServe()
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "host to listen on (overrides config)")
	cmd.Flags().IntVar(&port, "port", 0, "port to listen on (overrides config)")

	return cmd
}
