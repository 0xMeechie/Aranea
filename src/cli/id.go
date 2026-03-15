package cli

import (
	"fmt"
	"log/slog"

	"github.com/0xMeechie/Aranea/src/node"
	"github.com/spf13/cobra"
)

func newIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "id",
		Short: "return the id of the node",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := node.LoadIdentity()
			if err != nil {
				return fmt.Errorf("failed to load identity: %w", err)
			}
			slog.Info("node identity",
				"node_name", id.NodeName,
				"public_key", id.PublicKey,
			)
			return nil
		},
	}
}
