package cli

import (
	"fmt"
	"log/slog"

	"github.com/0xMeechie/Aranea/src/node"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"i"},
		Short:   "initialize the node",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("node initializing")
			if err := node.InitNode(force); err != nil {
				return fmt.Errorf("init failed: %w", err)
			}
			slog.Info("node initialized successfully")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false,
		"force creation of new keys, overwriting existing config")

	return cmd
}
