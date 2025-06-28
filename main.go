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
	GCSecretID    string
	GCSecretKey   string
	GCAccountID   string
	YNABToken     string
	YNABBudgetID  string
	YNABAccountID string
	CronSchedule  string

	newRelic *newrelic.Application
	gc       *GoCardless
	ynabc    ynab.ClientServicer
}

func main() {
	l := slog.Default()
	secretID := os.Getenv("GC_SECRET_ID")
	secretKey := os.Getenv("GC_SECRET_KEY")
	gcAccountID := os.Getenv("GC_ACCOUNT_ID")     // GoCardless account ID
	ynabAccountID := os.Getenv("YNAB_ACCOUNT_ID") // YNAB account ID
	ynabBudgetID := os.Getenv("YNAB_BUDGET_ID")
	ynabToken := os.Getenv("YNAB_TOKEN")       // YNAB personal access token
	cronSchedule := os.Getenv("CRON_SCHEDULE") // Cron schedule for synchronization
	newRelicLicenceKey := os.Getenv("NEW_RELIC_LICENCE_KEY")
	newRelicAppName := os.Getenv("NEW_RELIC_APP_NAME")
	l.Info("newRelicLicenceKey", newRelicLicenceKey, `os.Getenv("NEW_RELIC_LICENCE_KEY")`, os.Getenv("NEW_RELIC_LICENCE_KEY"))
	l.Info("newRelicAppName", newRelicAppName, `os.Getenv("NEW_RELIC_APP_NAME")`, os.Getenv("NEW_RELIC_APP_NAME"))

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
	if newRelicApp.WaitForConnection(time.Second*10) != nil {
		l.Error("failed to connect to New Relic")
		os.Exit(1)
	}
	defer newRelicApp.Shutdown(time.Second)

	txn := newRelicApp.StartTransaction("startup", newrelic.WithFunctionLocation())
	defer txn.End()

	txn1 := newRelicApp.StartTransaction("startup", newrelic.WithFunctionLocation())
	time.Sleep(time.Second)
	txn1.End()

	oys := openYNABSync{
		GCSecretID:    secretID,
		GCSecretKey:   secretKey,
		GCAccountID:   gcAccountID,
		YNABToken:     ynabToken,
		YNABBudgetID:  ynabBudgetID,
		YNABAccountID: ynabAccountID,
		CronSchedule:  cronSchedule,
		newRelic:      newRelicApp,
		gc:            nil,
		ynabc:         nil,
	}

	// Default cron schedule: run every minute
	if cronSchedule == "" {
		cronSchedule = "0 9/12 * * *"
	}

	if secretID == "" || secretKey == "" || gcAccountID == "" || ynabAccountID == "" || ynabToken == "" {
		fmt.Println("Error: GC_SECRET_ID, GC_SECRET_KEY, GC_ACCOUNT_ID, YNAB_ACCOUNT_ID and YNAB_TOKEN environment variables are required")
		os.Exit(1)
	}

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
	ctx := context.Background()
	txn := oys.newRelic.StartTransaction("synchronization", newrelic.WithFunctionLocation())
	defer txn.End()

	funcStartedAt := time.Now()
	l := slog.Default().With("accountID", oys.GCAccountID)
	to := time.Now()
	from := to.AddDate(0, 0, -14)
	txn.AddAttribute("from", from.Format("2006-01-02"))
	txn.AddAttribute("to", to.Format("2006-01-02"))

	transactions, err := oys.gc.ListTransactions(ctx, oys.GCAccountID, from, to)
	if err != nil {
		txn.NoticeError(err)
		l.ErrorContext(ctx, "failed to list transactions", "error", err)
		return
	}

	txn.AddAttribute("transactionsCount", len(transactions))

	ctx = newrelic.NewContext(ctx, txn)
	if err := uploadToYNAB(ctx, oys.ynabc, oys.YNABAccountID, oys.YNABBudgetID, transactions); err != nil {
		txn.NoticeError(err)
		l.ErrorContext(ctx, "failed to upload transactions", "error", err)
		return
	}

	l.InfoContext(ctx, "finished", "duration", time.Since(funcStartedAt))
}
