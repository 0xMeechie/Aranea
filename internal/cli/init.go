package cli

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/0xMeechie/Aranea/internal/node"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var force bool
	var path string

	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"i"},
		Short:   "initialize the node",
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("node initializing...")
			cfg := node.InitConfig{Forced: force, Path: path}
			if err := node.Init(cfg); err != nil {
				if errors.Is(err, node.ErrAlreadyInitialized) {
					slog.Info("node is already initialized")
					return nil
				}
				return fmt.Errorf("init failed: %w", err)
			}
			slog.Info("node initialized successfully")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false,
		"force creation of new keys, overwriting existing config")
	cmd.Flags().StringVarP(&path, "path", "p", "", "the path to the aranea config directory. ./.config is used by default")

	return cmd
}
