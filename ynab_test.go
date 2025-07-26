package main

import (
	"context"
	"testing"
	"time"

	"github.com/brunomvsouza/ynab.go/api/transaction"
	"github.com/stretchr/testify/assert"
)

func TestToYNABTransaction(t *testing.T) {
	// Setup
	date := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	gcTransactions := []Transaction{
		{
			ID:         "tx1",
			Date:       date,
			AmountMili: 100500,
			Memo:       "Payment for services",
			Name:       "John Doe",
		},
	}
	ynabAccountID := "account123"

	// Test
	ynabTransactions := toYNABTransaction(ynabAccountID, gcTransactions)

	// Assert
	assert.Len(t, ynabTransactions, 1)

	tx := ynabTransactions[0]
	assert.Equal(t, ynabAccountID, tx.AccountID)
	assert.Equal(t, int64(100500), tx.Amount)
	assert.NotNil(t, tx.Memo)
	assert.Equal(t, "Payment for services", *tx.Memo)
	assert.NotNil(t, tx.PayeeName)
	assert.Equal(t, "John Doe", *tx.PayeeName)
	assert.Equal(t, transaction.ClearingStatusCleared, tx.Cleared)
	assert.False(t, tx.Approved)

	// Check import ID format
	expectedImportID := "YNAB:100500:2023-01-01:1"
	assert.NotNil(t, tx.ImportID)
	assert.Equal(t, expectedImportID, *tx.ImportID)
}

func TestToImportID(t *testing.T) {
	// Setup
	date := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	trans := Transaction{
		ID:         "tx1",
		Date:       date,
		AmountMili: 100500,
		Memo:       "Payment for services",
		Name:       "John Doe",
	}

	// Test
	importID := toImportID(trans)

	// Assert
	expectedImportID := "YNAB:100500:2023-01-01"
	assert.Equal(t, expectedImportID, importID)
}

func TestToImportIDWithOccurrence(t *testing.T) {
	// Setup
	date := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	trans := Transaction{
		ID:         "tx1",
		Date:       date,
		AmountMili: 100500,
		Memo:       "Payment for services",
		Name:       "John Doe",
	}
	occurrence := 2

	// Test
	importID := toImportIDWithOccurrence(trans, occurrence)

	// Assert
	expectedImportID := "YNAB:100500:2023-01-01:2"
	assert.Equal(t, expectedImportID, importID)
}

func TestUploadToYNAB(t *testing.T) {
	// Setup
	ctx := context.Background()
	ynabAccountID := "account123"
	ynabBudgetID := "budget123"
	date := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	transactions := []Transaction{
		{
			ID:         "tx1",
			Date:       date,
			AmountMili: 100500,
			Memo:       "Payment for services",
			Name:       "John Doe",
		},
	}

	ynaberMock := newMockynaber(t)
	ynabTransactions := toYNABTransaction(ynabAccountID, transactions)
	ynaberMock.EXPECT().CreateTransactions(ynabBudgetID, ynabTransactions).Return(&transaction.OperationSummary{Transactions: []*transaction.Transaction{{}}}, nil)

	// Test
	err := uploadToYNAB(ctx, ynaberMock, ynabAccountID, ynabBudgetID, transactions)

	// Assert
	assert.NoError(t, err)
}
