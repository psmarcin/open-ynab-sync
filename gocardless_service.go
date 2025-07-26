package main

import (
	"context"
	"time"
)

// GoCardlessServicer implementation using the existing GoCardless struct
type GoCardlessService struct {
	gc GoCardless
}

// NewGoCardlessService creates a new GoCardlessServicer
func NewGoCardlessService(secretID, secretKey string) GoCardlessServicer {
	return &GoCardlessService{
		gc: NewGoCardless(secretID, secretKey),
	}
}

// LogIn logs in to the GoCardless API
func (s *GoCardlessService) LogIn(ctx context.Context) error {
	return s.gc.LogIn(ctx)
}

// RefreshToken refreshes the GoCardless API token
func (s *GoCardlessService) RefreshToken(ctx context.Context) error {
	return s.gc.RefreshToken(ctx)
}

// ListTransactions lists transactions from the GoCardless API
func (s *GoCardlessService) ListTransactions(ctx context.Context, accountID string, from, to time.Time) ([]Transaction, error) {
	return s.gc.ListTransactions(ctx, accountID, from, to)
}
