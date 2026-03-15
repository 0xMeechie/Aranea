package cli

import "github.com/spf13/cobra"

// NewRootCmd builds and returns the root cobra command with all subcommands attached.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "aranea",
		Short:   "aranea is a cli tool for interacting with nodes",
		Version: "0.0.1",
	}

	root.AddCommand(
		newVersionCmd(),
		newIDCmd(),
		newPingCmd(),
		newInitCmd(),
		newServeCmd(),
		newStatusCmd(),
		newAuditCmd(),
	)

	return root
}
