package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPingCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "ping",
		Aliases: []string{"p"},
		Short:   "ping a node",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pinging a node")
			// TODO: connect to node and ping
		},
	}
}
