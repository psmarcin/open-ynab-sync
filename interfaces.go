package main

import (
	"context"
	"time"

	"github.com/brunomvsouza/ynab.go/api/transaction"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// GoCardlessServicer defines the interface for interacting with the GoCardless API
type GoCardlessServicer interface {
	LogIn(ctx context.Context) error
	RefreshToken(ctx context.Context) error
	ListTransactions(ctx context.Context, accountID string, from time.Time, to time.Time) ([]Transaction, error)
}

// YNABServicer defines the interface for interacting with the YNAB API
type YNABServicer interface {
	CreateTransactions(budgetID string, p []transaction.PayloadTransaction) (*transaction.OperationSummary, error)
}

// SynchronizationServicer defines the interface for synchronizing transactions between GoCardless and YNAB
type SynchronizationServicer interface {
	SynchronizeTransactions(ctx context.Context) error
	SynchronizeTransaction(ctx context.Context, j job) error
}

// MonitoringServicer defines the interface for monitoring and instrumentation
type MonitoringServicer interface {
	StartTransaction(name string) *newrelic.Transaction
	RecordError(txn *newrelic.Transaction, err error)
	AddAttribute(txn *newrelic.Transaction, key string, value interface{})
	NewContext(ctx context.Context, txn *newrelic.Transaction) context.Context
	FromContext(ctx context.Context) *newrelic.Transaction
	StartSegment(txn *newrelic.Transaction, name string) *newrelic.Segment
	Shutdown(timeout time.Duration)
}
