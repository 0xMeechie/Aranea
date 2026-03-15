package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "serve",
		Aliases: []string{"s"},
		Short:   "start the agent",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("starting agent...")
			// TODO: load config, identity, start runtime loop
		},
	}
}
