package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	slog.Info("aranead starting")

	// TODO: load config
	// TODO: load identity
	// TODO: start agent runtime loop

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	slog.Info("aranead shutting down")
}
