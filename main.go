package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/go-co-op/gocron/v2"
)

func main() {
	l := slog.Default()

	// Load configuration from environment
	config, err := LoadConfigFromEnv()
	if err != nil {
		l.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create a service container
	container, err := NewServiceContainer(config)
	if err != nil {
		l.Error("failed to initialize services", "error", err)
		os.Exit(1)
	}

	// Get monitoring service for startup transaction
	monitorService := container.MonitorService()
	txn := monitorService.StartTransaction("startup")
	defer txn.End()

	// Get synchronization service
	syncService := container.SyncService()

	// Create context with transaction
	ctx := monitorService.NewContext(context.Background(), txn)

	// Log configuration
	l.Info("configuration", "cron", config.CronSchedule)
	for _, job := range config.Jobs {
		l.Info("job", "gocardless_account_id", job.GCAccountID, "ynab_account_id", job.YNABAccountID, "ynab_budget_id", job.YNABBudgetID)
	}

	// Set up scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		monitorService.RecordError(txn, err)
		l.ErrorContext(ctx, "failed to create scheduler", "error", err)
		os.Exit(1)
	}
	defer func() { _ = s.Shutdown() }()

	// Create a synchronization job
	_, err = s.NewJob(
		gocron.CronJob(config.CronSchedule, false),
		gocron.NewTask(func() {
			if err := syncService.SynchronizeTransactions(context.Background()); err != nil {
				l.Error("synchronization failed", "error", err)
			}
		}),
	)
	if err != nil {
		monitorService.RecordError(txn, err)
		l.ErrorContext(ctx, "failed to create job", "error", err)
		os.Exit(1)
	}

	// Start scheduler
	s.Start()

	// Block until shutdown
	select {}
}
