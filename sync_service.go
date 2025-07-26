package main

import (
	"context"
	"log/slog"
	"time"
)

// SyncService implements the SynchronizationServicer interface
type SyncService struct {
	gcService      GoCardlessServicer
	ynabService    YNABServicer
	monitorService MonitoringServicer
	jobs           []job
}

// NewSyncService creates a new SynchronizationServicer
func NewSyncService(gcService GoCardlessServicer, ynabService YNABServicer, monitorService MonitoringServicer, jobs []job) SynchronizationServicer {
	return &SyncService{
		gcService:      gcService,
		ynabService:    ynabService,
		monitorService: monitorService,
		jobs:           jobs,
	}
}

// SynchronizeTransactions synchronizes all transactions for all jobs
func (s *SyncService) SynchronizeTransactions(ctx context.Context) error {
	for _, j := range s.jobs {
		if err := s.SynchronizeTransaction(ctx, j); err != nil {
			return err
		}
	}
	return nil
}

// SynchronizeTransaction synchronizes transactions for a single job
func (s *SyncService) SynchronizeTransaction(ctx context.Context, j job) error {
	txn := s.monitorService.StartTransaction("synchronization")
	defer txn.End()

	funcStartedAt := time.Now()
	l := slog.Default().With("gocardless_account_id", j.GCAccountID, "ynab_account_id", j.YNABAccountID, "ynab_budget_id", j.YNABBudgetID)
	to := time.Now().UTC().Truncate(time.Hour)
	from := to.AddDate(0, 0, -20).Truncate(24 * time.Hour)
	s.monitorService.AddAttribute(txn, "from", from.Format("2006-01-02"))
	s.monitorService.AddAttribute(txn, "to", to.Format("2006-01-02"))
	s.monitorService.AddAttribute(txn, "gocardlessAccountId", j.GCAccountID)
	s.monitorService.AddAttribute(txn, "ynabAccountId", j.YNABAccountID)
	s.monitorService.AddAttribute(txn, "ynabBudgetId", j.YNABBudgetID)

	ctx = s.monitorService.NewContext(ctx, txn)
	if err := s.gcService.LogIn(ctx); err != nil {
		s.monitorService.RecordError(txn, err)
		l.ErrorContext(ctx, "failed to log in", "error", err)
		return err
	}

	transactions, err := s.gcService.ListTransactions(ctx, j.GCAccountID, from, to)
	if err != nil {
		s.monitorService.RecordError(txn, err)
		l.ErrorContext(ctx, "failed to list transactions", "error", err)
		return err
	}

	s.monitorService.AddAttribute(txn, "transactionsCount", len(transactions))
	if err := uploadToYNAB(ctx, s.ynabService, j.YNABAccountID, j.YNABBudgetID, transactions); err != nil {
		s.monitorService.RecordError(txn, err)
		l.ErrorContext(ctx, "failed to upload transactions", "error", err)
		return err
	}

	l.InfoContext(ctx, "finished", "duration", time.Since(funcStartedAt))
	return nil
}
