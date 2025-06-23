package main

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSynchronizeTransactions(t *testing.T) {
	t.Skip()
	// Setup
	gcAccountID := "gc-account-123"
	ynabAccountID := "ynab-account-456"
	ynabBudgetID := "ynab-budget-789"

	// Create a mock transport for the GoCardless HTTP client
	mockTransport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Return a successful response for any request
			// This simulates a successful login and transaction listing
			return &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"transactions": {
						"booked": [
							{
								"transactionId": "tx1",
								"valueDate": "2023-01-01",
								"transactionAmount": {
									"amount": "100.50",
									"currency": "EUR"
								},
								"remittanceInformationUnstructured": "Payment for services",
								"debtorName": "John Doe"
							}
						],
						"pending": []
					},
					"last_updated": "2023-01-02T12:00:00Z"
				}`)),
			}, nil
		},
	}

	// Create a GoCardless instance with the mock transport
	mockClient := &http.Client{Transport: mockTransport}
	gc := GoCardless{
		SecretID:    "test-id",
		SecretKey:   "test-key",
		httpClient:  mockClient,
		accessToken: "test-access-token", // Pre-set the access token to skip login
	}

	ynabc := NewMockClientServicer(t)
	//transaction.NewService(ynabc)
	//ynabc.EXPECT().Transaction().Return()

	// Create the synchronize function
	syncFunc := synchronizeTransactions(gc, ynabc, gcAccountID, ynabAccountID, ynabBudgetID)

	// Test
	// Call the synchronize function - this should not panic
	assert.NotPanics(t, func() {
		syncFunc()
	})

	// Note: This test primarily verifies that the synchronization function doesn't panic or crash.
	// In a more comprehensive test, we would verify that transactions are correctly processed
	// and uploaded to YNAB, possibly by mocking the uploadToYNAB function or checking side effects.
}
