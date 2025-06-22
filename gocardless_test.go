package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockTransport is a mock implementation of http.RoundTripper
type MockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestGoCardlessLogin(t *testing.T) {
	// Setup
	mockTransport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Check request is correct
			assert.Equal(t, "https://bankaccountdata.gocardless.com/api/v2/token/new/", req.URL.String())
			assert.Equal(t, http.MethodPost, req.Method)

			// Check request body
			body, _ := io.ReadAll(req.Body)
			var loginReq loginRequest
			err := json.Unmarshal(body, &loginReq)
			assert.NoError(t, err)
			assert.Equal(t, "test-id", loginReq.SecretID)
			assert.Equal(t, "test-key", loginReq.SecretKey)

			// Return mock response
			resp := &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"access": "test-access-token",
					"refresh": "test-refresh-token"
				}`)),
			}
			return resp, nil
		},
	}

	mockClient := &http.Client{Transport: mockTransport}

	gc := GoCardless{
		SecretID:   "test-id",
		SecretKey:  "test-key",
		httpClient: mockClient,
	}

	// Test
	err := gc.LogIn(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "test-access-token", gc.accessToken)
	assert.Equal(t, "test-refresh-token", gc.refreshToken)
}

func TestGoCardlessRefreshToken(t *testing.T) {
	// Setup
	mockTransport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Check request is correct
			assert.Equal(t, "https://bankaccountdata.gocardless.com/api/v2/token/refresh/", req.URL.String())
			assert.Equal(t, http.MethodPost, req.Method)

			// Check request body
			body, _ := io.ReadAll(req.Body)
			var refreshReq refreshTokenRequest
			err := json.Unmarshal(body, &refreshReq)
			assert.NoError(t, err)
			assert.Equal(t, "old-refresh-token", refreshReq.RefreshToken)

			// Return mock response
			resp := &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"access": "new-access-token"
				}`)),
			}
			return resp, nil
		},
	}

	mockClient := &http.Client{Transport: mockTransport}

	gc := GoCardless{
		refreshToken: "old-refresh-token",
		httpClient:   mockClient,
	}

	// Test
	err := gc.RefreshToken(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "new-access-token", gc.accessToken)
}

func TestListTransactions(t *testing.T) {
	// Setup
	mockTransport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Check request is correct
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, "Bearer test-access-token", req.Header.Get("Authorization"))

			// Return mock response
			resp := &http.Response{
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
			}
			return resp, nil
		},
	}

	mockClient := &http.Client{Transport: mockTransport}

	gc := GoCardless{
		accessToken: "test-access-token",
		httpClient:  mockClient,
	}

	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)

	// Test
	transactions, err := gc.ListTransactions(context.Background(), "account123", from, to)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, transactions, 1)
	assert.Equal(t, "tx1", transactions[0].ID)
	assert.Equal(t, int64(100500), transactions[0].AmountMili)
	assert.Equal(t, "Payment for services", transactions[0].Memo)
	assert.Equal(t, "John Doe", transactions[0].Name)
}

func TestToTransactions(t *testing.T) {
	// Setup
	response := goCardlessListTransactionResponse{
		Transactions: struct {
			Booked  []goCardlessListTransactionResponseTransaction `json:"booked"`
			Pending []goCardlessListTransactionResponseTransaction `json:"pending"`
		}{
			Booked: []goCardlessListTransactionResponseTransaction{
				{
					TransactionId: "tx1",
					ValueDate:     "2023-01-01",
					TransactionAmount: struct {
						Amount   string `json:"amount"`
						Currency string `json:"currency"`
					}{
						Amount:   "100.50",
						Currency: "EUR",
					},
					RemittanceInformationUnstructured: "Payment for services",
					DebtorName:                        "John Doe",
				},
			},
		},
	}

	// Test
	transactions := toTransactions(response)

	// Assert
	assert.Len(t, transactions, 1)
	assert.Equal(t, "tx1", transactions[0].ID)
	assert.Equal(t, int64(100500), transactions[0].AmountMili)
	assert.Equal(t, "Payment for services", transactions[0].Memo)
	assert.Equal(t, "John Doe", transactions[0].Name)
}

func TestListTransactionsWithResetIn(t *testing.T) {
	// Setup
	mockClient := &http.Client{}

	gc := GoCardless{
		accessToken: "test-access-token",
		httpClient:  mockClient,
		resetIn:     time.Minute * 5, // Set resetIn to 5 minutes
	}

	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)

	// Test
	transactions, err := gc.ListTransactions(context.Background(), "account123", from, to)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, transactions)
}

func TestListTransactionsTooManyRequests(t *testing.T) {
	// Setup
	mockTransport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Return a 429 Too Many Requests response
			resp := &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header: http.Header{
					"Http_X_Ratelimit_Account_Success_Reset": []string{"300"}, // 5 minutes in seconds
				},
				Body: io.NopCloser(bytes.NewBufferString(`{"error": "too many requests"}`)),
			}
			return resp, nil
		},
	}

	mockClient := &http.Client{Transport: mockTransport}

	gc := GoCardless{
		accessToken: "test-access-token",
		httpClient:  mockClient,
	}

	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)

	// Test
	transactions, err := gc.ListTransactions(context.Background(), "account123", from, to)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, transactions)
	// Check that the error message contains "too many requests"
	assert.Contains(t, err.Error(), "too many requests")
}

func TestListTransactionsTooManyRequestsInvalidHeader(t *testing.T) {
	// Setup
	mockTransport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Return a 429 Too Many Requests response with invalid header format
			resp := &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header: http.Header{
					http.CanonicalHeaderKey("http_x_ratelimit_account_success_reset"): []string{"invalid-duration"}, // Invalid duration format
				},
				Body: io.NopCloser(bytes.NewBufferString(`{"error": "too many requests"}`)),
			}
			return resp, nil
		},
	}

	mockClient := &http.Client{Transport: mockTransport}

	gc := GoCardless{
		accessToken: "test-access-token",
		httpClient:  mockClient,
	}

	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)

	// Test
	transactions, err := gc.ListTransactions(context.Background(), "account123", from, to)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, transactions)
	// Check that the error message contains "failed to parse rate limit reset header"
	assert.Contains(t, err.Error(), "failed to parse rate limit reset header")
}
