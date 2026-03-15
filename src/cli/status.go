package cli

import (
	"fmt"
	"log/slog"

	"github.com/0xMeechie/Aranea/src/node"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "returns the status of the current node",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("checking node status")
			cfg, err := node.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			fmt.Printf("NODE INFO\n  config_dir: %s\n  version:    %d\n  log_level:  %s\n",
				cfg.ConfigDir, cfg.Version, cfg.LogLevel)
			return nil
		},
	}
}
