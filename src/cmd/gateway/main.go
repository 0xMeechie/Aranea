package main

import (
	"log/slog"
	"os"

	"github.com/0xMeechie/Aranea/src/gateway"
)

func main() {
	cfg := gateway.DefaultConfig()

	if err := gateway.Serve(cfg); err != nil {
		slog.Error("gateway error", "err", err)
		os.Exit(1)
	}
}
