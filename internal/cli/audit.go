package cli

import (
	"log/slog"

	"github.com/spf13/cobra"
)

func newAuditCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "audit",
		Aliases: []string{"a"},
		Short:   "view recent audit events",
		Run: func(cmd *cobra.Command, args []string) {
			slog.Info("reading audit log")
			// TODO: open audit-*.ael files and print events
		},
	}
}
