package main

import (
	"github.com/brunomvsouza/ynab.go"
	"github.com/brunomvsouza/ynab.go/api/transaction"
)

// YNABService implements the YNABServicer interface
type YNABService struct {
	client ynaber
}

// NewYNABService creates a new YNABServicer
func NewYNABService(token string) YNABServicer {
	client := ynab.NewClient(token)
	return &YNABService{
		client: client.Transaction(),
	}
}

// CreateTransactions creates transactions in YNAB
func (s *YNABService) CreateTransactions(budgetID string, p []transaction.PayloadTransaction) (*transaction.OperationSummary, error) {
	return s.client.CreateTransactions(budgetID, p)
}
