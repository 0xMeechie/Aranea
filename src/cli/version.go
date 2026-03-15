package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Show CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("aranea v0.0.1")
		},
	}
}
