package main

import (
	"testing"
	"time"

	"github.com/brunomvsouza/ynab.go/api"
	"github.com/brunomvsouza/ynab.go/api/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSynchronizeTransactions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		goCardlessMock := newMockgoCardlesser(t)
		goCardlessMock.EXPECT().LogIn(mock.Anything).Return(nil)

		nowTS := time.Now().UTC().Truncate(time.Hour)
		from := nowTS.AddDate(0, 0, -20).Truncate(24 * time.Hour)

		trans1 := Transaction{
			ID:         "123",
			Date:       nowTS.AddDate(0, 0, -1),
			AmountMili: 98765,
			Memo:       "memo",
			Name:       "John Doe",
		}
		goCardlessMock.EXPECT().ListTransactions(mock.Anything, "aaa", from, nowTS).Return([]Transaction{trans1}, nil)

		ynabc := newMockynaber(t)
		importID := toImportIDWithOccurrence(trans1, 1)
		ynabc.EXPECT().CreateTransactions("ccc", []transaction.PayloadTransaction{
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

		oys := openYNABSync{
			newRelic: nil,
			gc:       goCardlessMock,
			ynabc:    ynabc,
			Jobs: []job{
				{
					GCAccountID:   "aaa",
					YNABAccountID: "bbb",
					YNABBudgetID:  "ccc",
				},
			},
		}

		assert.NotPanics(t, func() {
			oys.synchronizeTransactions()
		})
	})
}
