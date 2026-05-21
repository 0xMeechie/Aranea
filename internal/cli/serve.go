package cli

import (
	"fmt"

	"github.com/0xMeechie/Aranea/internal/node"
	"github.com/0xMeechie/Aranea/pkg/policy"
	"github.com/0xMeechie/Aranea/pkg/runtime"
	"github.com/0xMeechie/Aranea/pkg/server"
	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	var (
		path string
		addr string
	)
	cmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"s"},
		Short:   "start the agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeConfig, id, err := node.Load(node.InitConfig{Path: path})
			if err != nil {
				return fmt.Errorf("load node: %w", err)
			}

			kp, err := policy.LoadKeyPair(nodeConfig.KeyDir())
			if err != nil {
				return fmt.Errorf("load keys: %w", err)
			}

			rt, err := runtime.New(id.NodeName, *kp, nodeConfig.LogDir())
			if err != nil {
				return fmt.Errorf("init runtime: %w", err)
			}
			defer rt.Close()

			srv := server.New(rt, server.Config{Addr: addr})
			return srv.Start()
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "the path to the aranea config directory. ./.config is used by default")
	cmd.Flags().StringVar(&addr, "addr", ":8080", "host:port the HTTP server binds to")
	return cmd
}
