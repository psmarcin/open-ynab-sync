package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/brunomvsouza/ynab.go"
	"github.com/go-co-op/gocron/v2"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type openYNABSync struct {
	GCSecretID   string
	GCSecretKey  string
	YNABToken    string
	CronSchedule string
	Jobs         []job

	newRelic *newrelic.Application
	gc       *GoCardless
	ynabc    ynab.ClientServicer
}

func main() {
	l := slog.Default()
	secretID := os.Getenv("GC_SECRET_ID")
	secretKey := os.Getenv("GC_SECRET_KEY")
	ynabToken := os.Getenv("YNAB_TOKEN")       // YNAB personal access token
	cronSchedule := os.Getenv("CRON_SCHEDULE") // Cron schedule for synchronization
	newRelicLicenceKey := os.Getenv("NEW_RELIC_LICENCE_KEY")
	newRelicAppName := os.Getenv("NEW_RELIC_APP_NAME")

	newRelicApp, err := newrelic.NewApplication(
		newrelic.ConfigAppName(newRelicAppName),
		newrelic.ConfigLicense(newRelicLicenceKey),
		newrelic.ConfigAppLogMetricsEnabled(true),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		l.Error("failed to initialize New Relic", "error", err)
		os.Exit(1)
	}
	if newRelicApp.WaitForConnection(time.Second*30) != nil {
		l.Error("failed to connect to New Relic")
		os.Exit(1)
	}
	defer newRelicApp.Shutdown(time.Second)

	txn := newRelicApp.StartTransaction("startup", newrelic.WithFunctionLocation())
	defer txn.End()

	jobs, err := envToJobs(os.Getenv("JOBS"))
	if err != nil {
		txn.NoticeError(err)
		l.Error("failed to parse jobs", "error", err)
		os.Exit(1)
	}

	oys := openYNABSync{
		GCSecretID:   secretID,
		GCSecretKey:  secretKey,
		YNABToken:    ynabToken,
		CronSchedule: cronSchedule,
		Jobs:         jobs,

		newRelic: newRelicApp,
		gc:       nil,
		ynabc:    nil,
	}

	// Default cron schedule: run every minute
	if cronSchedule == "" {
		cronSchedule = "0 6,18 * * *"
	}

	if secretID == "" || secretKey == "" || ynabToken == "" {
		fmt.Println("Error: GC_SECRET_ID, GC_SECRET_KEY, GC_ACCOUNT_ID, YNAB_ACCOUNT_ID and YNAB_TOKEN environment variables are required")
		os.Exit(1)
	}

	l.Info("configuration", "cron", cronSchedule, "jobs", jobs)

	ctx := newrelic.NewContext(context.Background(), txn)

	gc := NewGoCardless(secretID, secretKey)
	oys.gc = &gc
	if err := gc.LogIn(ctx); err != nil {
		txn.NoticeError(err)
		l.ErrorContext(ctx, "failed to log in", "error", err)
		os.Exit(1)
	}

	ynabc := ynab.NewClient(ynabToken)
	oys.ynabc = ynabc

	s, err := gocron.NewScheduler()
	if err != nil {
		txn.NoticeError(err)
		l.ErrorContext(ctx, "failed to create scheduler", "error", err)
		os.Exit(1)
	}
	defer func() { _ = s.Shutdown() }()

	_, err = s.NewJob(
		gocron.CronJob(cronSchedule, false),
		gocron.NewTask(oys.synchronizeTransactions),
	)
	if err != nil {
		txn.NoticeError(err)
		l.ErrorContext(ctx, "failed to create job", "error", err)
		os.Exit(1)
	}

	s.Start()
	// block until you are ready to shut down
	select {}
}

func (oys *openYNABSync) synchronizeTransactions() {
	for _, j := range oys.Jobs {
		oys.synchronizeTransaction(j)
	}
}

func (oys *openYNABSync) synchronizeTransaction(j job) {
	ctx := context.Background()
	txn := oys.newRelic.StartTransaction("synchronization", newrelic.WithFunctionLocation())
	defer txn.End()

	funcStartedAt := time.Now()
	l := slog.Default().With("gocardless_account_id", j.GCAccountID, "ynab_account_id", j.YNABAccountID, "ynab_budget_id", j.YNABBudgetID)
	to := time.Now()
	from := to.AddDate(0, 0, -20)
	txn.AddAttribute("from", from.Format("2006-01-02"))
	txn.AddAttribute("to", to.Format("2006-01-02"))
	txn.AddAttribute("gocardlessAccountId", j.GCAccountID)
	txn.AddAttribute("ynabAccountId", j.YNABAccountID)
	txn.AddAttribute("ynabBudgetId", j.YNABBudgetID)

	transactions, err := oys.gc.ListTransactions(ctx, j.GCAccountID, from, to)
	if err != nil {
		txn.NoticeError(err)
		l.ErrorContext(ctx, "failed to list transactions", "error", err)
		return
	}

	txn.AddAttribute("transactionsCount", len(transactions))

	ctx = newrelic.NewContext(ctx, txn)
	if err := uploadToYNAB(ctx, oys.ynabc, j.YNABAccountID, j.YNABBudgetID, transactions); err != nil {
		txn.NoticeError(err)
		l.ErrorContext(ctx, "failed to upload transactions", "error", err)
		return
	}

	l.InfoContext(ctx, "finished", "duration", time.Since(funcStartedAt))
}
