package cli

import (
	"log/slog"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "returns the status of the current node",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("checking node status")
			return nil
		},
	}
}
