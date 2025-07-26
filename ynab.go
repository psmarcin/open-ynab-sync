package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/brunomvsouza/ynab.go/api"
	"github.com/brunomvsouza/ynab.go/api/transaction"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/pkg/errors"
)

type ynaber interface {
	CreateTransactions(budgetID string, p []transaction.PayloadTransaction) (*transaction.OperationSummary, error)
}

func uploadToYNAB(ctx context.Context, ynabc ynaber, ynabAccountID, ynabBudgetID string, transactions []Transaction) error {
	txn := newrelic.FromContext(ctx)
	seg := txn.StartSegment("uploadToYNAB")
	defer seg.End()
	l := slog.Default()
	payloadTransactions := toYNABTransaction(ynabAccountID, transactions)
	seg.AddAttribute("payloadTransactionsCount", len(payloadTransactions))
	for _, payloadTransaction := range payloadTransactions {
		l.InfoContext(ctx, "uploading transaction", "date", payloadTransaction.Date, "payee", *payloadTransaction.PayeeName, "memo", *payloadTransaction.Memo, "amount", payloadTransaction.Amount)
	}

	result, err := ynabc.CreateTransactions(ynabBudgetID, payloadTransactions)
	if err != nil {
		return errors.Wrapf(err, "failed to upload transactions")
	}

	seg.AddAttribute("resultTransactionsCount", len(result.Transactions))

	l.InfoContext(ctx, "successfully uploaded transactions", "count", len(result.Transactions))
	return nil
}

func toYNABTransaction(ynabAccountID string, gcTransactions []Transaction) []transaction.PayloadTransaction {
	var ynabTransactions []transaction.PayloadTransaction
	uniqueImportIDs := make(map[string]int64)
	for _, gcTransaction := range gcTransactions {
		d, err := api.DateFromString(gcTransaction.Date.Format("2006-01-02"))
		if err != nil {
			panic(err)
		}

		occurrence := uniqueImportIDs[toImportID(gcTransaction)]
		occurrence = +1
		uniqueImportIDs[toImportID(gcTransaction)] = occurrence

		importID := toImportIDWithOccurrence(gcTransaction, int(occurrence))

		ynabTransactions = append(ynabTransactions, transaction.PayloadTransaction{
			ID:         gcTransaction.ID,
			AccountID:  ynabAccountID,
			Date:       d,
			Amount:     gcTransaction.AmountMili,
			Cleared:    transaction.ClearingStatusCleared, //
			Approved:   false,
			PayeeID:    nil,
			PayeeName:  &gcTransaction.Name,
			CategoryID: nil,
			Memo:       &gcTransaction.Memo,
			ImportID:   &importID,
		})
	}

	return ynabTransactions
}

func toImportIDWithOccurrence(transaction Transaction, occurrence int) string {
	return fmt.Sprintf("%s:%d", toImportID(transaction), occurrence)
}

func toImportID(transaction Transaction) string {
	return fmt.Sprintf("YNAB:%d:%s", transaction.AmountMili, transaction.Date.Format("2006-01-02"))
}
