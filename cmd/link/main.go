package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"psmarcin.github.com/open-ynab-sync/cmd/link/api"
	"psmarcin.github.com/open-ynab-sync/cmd/link/auth"
	"psmarcin.github.com/open-ynab-sync/cmd/link/config"
	"psmarcin.github.com/open-ynab-sync/cmd/link/server"
)

func main() {
	// Initialize logger
	logger := slog.Default()

	// Set up context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		logger.Info("received shutdown signal")
		cancel()
	}()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create GoCardless client
	gcClient := api.NewGoCardless(cfg.GCSecretID, cfg.GCSecretKey, cfg.HTTPTimeout, logger)

	// Create callback server
	callbackServer := server.NewCallbackServer(cfg.Port, logger)

	// Create authorization flow
	authFlow := auth.NewAuthFlow(gcClient, cfg, logger, callbackServer)
	authFlow.AutoOpenBrowser = true

	// Execute authorization flow
	accounts, err := authFlow.Execute(ctx)
	if err != nil {
		logger.Error("authorization flow failed", "error", err)
		os.Exit(1)
	}

	// Display linked accounts
	logger.Info("Successfully linked accounts:")
	for i, accountID := range accounts {
		logger.Info("account", "id", accountID, "index", i+1)
	}
}
