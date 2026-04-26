package cli

import (
	"fmt"

	"github.com/0xMeechie/Aranea/internal/node"
	"github.com/spf13/cobra"
)

var path string

func newIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id",
		Short: "return the id of the node",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := node.InitConfig{Path: path}
			_, id, err := node.Load(cfg)
			if err != nil {
				return err
			}

			fmt.Printf("Node ID: %s \n", id.NodeName)
			return nil
		},
	}
	cmd.Flags().StringVarP(&path, "path", "p", "", "the path to the aranea config directory. ./.config is used by default")
	return cmd
}
