package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/brunomvsouza/ynab.go"
	"github.com/go-co-op/gocron/v2"
)

func main() {
	l := slog.Default()
	secretID := os.Getenv("GC_SECRET_ID")
	secretKey := os.Getenv("GC_SECRET_KEY")
	gcAccountID := os.Getenv("GC_ACCOUNT_ID")     // GoCardless account ID
	ynabAccountID := os.Getenv("YNAB_ACCOUNT_ID") // YNAB account ID
	ynabBudgetID := os.Getenv("YNAB_BUDGET_ID")
	ynabToken := os.Getenv("YNAB_TOKEN")       // YNAB personal access token
	cronSchedule := os.Getenv("CRON_SCHEDULE") // Cron schedule for synchronization

	// Default cron schedule: run every minute
	if cronSchedule == "" {
		cronSchedule = "0 9/12 * * *"
	}

	if secretID == "" || secretKey == "" || gcAccountID == "" || ynabAccountID == "" || ynabToken == "" {
		fmt.Println("Error: GC_SECRET_ID, GC_SECRET_KEY, GC_ACCOUNT_ID, YNAB_ACCOUNT_ID and YNAB_TOKEN environment variables are required")
		os.Exit(1)
	}

	ctx := context.Background()

	gc := NewGoCardless(secretID, secretKey)
	if err := gc.LogIn(ctx); err != nil {
		l.ErrorContext(ctx, "failed to log in", "error", err)
		os.Exit(1)
	}

	ynabc := ynab.NewClient(ynabToken)

	s, err := gocron.NewScheduler()
	if err != nil {
		l.ErrorContext(ctx, "failed to create scheduler", "error", err)
		os.Exit(1)
	}
	defer func() { _ = s.Shutdown() }()

	_, err = s.NewJob(
		gocron.CronJob(cronSchedule, false),
		gocron.NewTask(synchronizeTransactions(gc, ynabc, gcAccountID, ynabAccountID, ynabBudgetID)),
	)
	if err != nil {
		l.ErrorContext(ctx, "failed to create job", "error", err)
		os.Exit(1)
	}

	s.Start()
	// block until you are ready to shut down
	select {}
}

func synchronizeTransactions(gc GoCardless, ynabc ynab.ClientServicer, gcAccountID, ynabAccountID, ynabBudgetID string) func() {
	return func() {
		ctx := context.Background()
		funcStartedAt := time.Now()
		l := slog.Default().With("accountID", gcAccountID)
		to := time.Now()
		from := to.AddDate(0, 0, -14)

		transactions, err := gc.ListTransactions(ctx, gcAccountID, from, to)
		if err != nil {
			l.ErrorContext(ctx, "failed to list transactions", "error", err)
			return
		}

		if err := uploadToYNAB(ctx, ynabc, ynabAccountID, ynabBudgetID, transactions); err != nil {
			l.ErrorContext(ctx, "failed to upload transactions", "error", err)
			return
		}

		l.InfoContext(ctx, "finished", "duration", time.Since(funcStartedAt))
	}

}
