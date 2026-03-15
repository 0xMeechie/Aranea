package main

import (
	"log/slog"
	"os"

	"github.com/0xMeechie/Aranea/src/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		slog.Error("cli error", "err", err)
		os.Exit(1)
	}
}
