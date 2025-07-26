package main

import (
	"context"
	"testing"
	"time"

	"github.com/brunomvsouza/ynab.go/api"
	"github.com/brunomvsouza/ynab.go/api/transaction"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSynchronizeTransactions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Create mock services
		goCardlessMock := newMockgoCardlesser(t)
		ynabMock := newMockynaber(t)
		monitorMock := NewMockMonitoringServicer(t)

		// Set up transaction time
		nowTS := time.Now().UTC().Truncate(time.Hour)
		from := nowTS.AddDate(0, 0, -20).Truncate(24 * time.Hour)

		// Create test transaction
		trans1 := Transaction{
			ID:         "123",
			Date:       nowTS.AddDate(0, 0, -1),
			AmountMili: 98765,
			Memo:       "memo",
			Name:       "John Doe",
		}

		// Set up mock expectations
		goCardlessMock.EXPECT().LogIn(mock.Anything).Return(nil)
		goCardlessMock.EXPECT().ListTransactions(mock.Anything, "aaa", from, nowTS).Return([]Transaction{trans1}, nil)

		importID := toImportIDWithOccurrence(trans1, 1)
		ynabMock.EXPECT().CreateTransactions("ccc", []transaction.PayloadTransaction{
			{
				ID:        "123",
				AccountID: "bbb",
				Date:      api.Date{Time: nowTS.AddDate(0, 0, -1).Truncate(24 * time.Hour)},
				Amount:    98765,
				Cleared:   transaction.ClearingStatusCleared,
				Approved:  false,
				PayeeName: &trans1.Name,
				Memo:      &trans1.Memo,
				ImportID:  &importID,
			},
		}).Return(&transaction.OperationSummary{Transactions: []*transaction.Transaction{{}}}, nil)

		// Mock monitoring service
		mockTxn := &newrelic.Transaction{}
		monitorMock.On("StartTransaction", "synchronization").Return(mockTxn)
		monitorMock.On("AddAttribute", mockTxn, mock.Anything, mock.Anything).Return()
		monitorMock.On("NewContext", mock.Anything, mockTxn).Return(context.Background())

		// Create test job
		testJob := job{
			GCAccountID:   "aaa",
			YNABAccountID: "bbb",
			YNABBudgetID:  "ccc",
		}

		// Create sync service with mocks
		syncService := NewSyncService(goCardlessMock, ynabMock, monitorMock, []job{testJob})

		// Test synchronization
		err := syncService.SynchronizeTransactions(context.Background())
		assert.NoError(t, err)
	})
}
