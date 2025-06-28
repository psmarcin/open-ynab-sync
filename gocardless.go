package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/pkg/errors"
)

type GoCardless struct {
	SecretID     string
	SecretKey    string
	accessToken  string
	refreshToken string
	httpClient   *http.Client
	resetIn      time.Duration
}

func NewGoCardless(secretID, secretKey string) GoCardless {
	httpClient := http.DefaultClient
	httpClient.Timeout = 20 * time.Second

	return GoCardless{
		SecretID:   secretID,
		SecretKey:  secretKey,
		httpClient: httpClient,
	}
}

type loginRequest struct {
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
}

type loginResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

func (gc *GoCardless) LogIn(ctx context.Context) error {
	txn := newrelic.FromContext(ctx)
	seg := txn.StartSegment("goCardlessLogIn")
	defer seg.End()

	l := slog.Default()
	requestBody := loginRequest{SecretID: gc.SecretID, SecretKey: gc.SecretKey}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return errors.Wrap(err, "failed to marshal request body")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://bankaccountdata.gocardless.com/api/v2/token/new/", bytes.NewReader(requestBodyJSON))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json")

	response, err := gc.httpClient.Do(request)
	if err != nil {
		return errors.Wrap(err, "failed to make request")
	}

	seg.AddAttribute("responseStatusCode", response.StatusCode)

	if response.StatusCode != 200 {
		return errors.Errorf("failed to login: %s", response.Status)
	}

	parsedResponse := loginResponse{}
	if err := json.NewDecoder(response.Body).Decode(&parsedResponse); err != nil {
		return errors.Wrap(err, "failed to parse response")
	}

	l.InfoContext(ctx, "logged in")

	gc.accessToken = parsedResponse.Access
	gc.refreshToken = parsedResponse.Refresh

	return nil
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh"`
}

type refreshTokenResponse struct {
	AccessToken string `json:"access"`
}

func (gc *GoCardless) RefreshToken(ctx context.Context) error {
	l := slog.Default()
	requestBody := refreshTokenRequest{RefreshToken: gc.refreshToken}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return errors.Wrap(err, "failed to marshal request body")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://bankaccountdata.gocardless.com/api/v2/token/refresh/", bytes.NewReader(requestBodyJSON))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json")

	response, err := gc.httpClient.Do(request)
	if err != nil {
		return errors.Wrap(err, "failed to make request")
	}

	if response.StatusCode != 200 {
		return errors.Errorf("failed to refresh token: %s", response.Status)
	}

	parsedResponse := refreshTokenResponse{}
	if err := json.NewDecoder(response.Body).Decode(&parsedResponse); err != nil {
		return errors.Wrap(err, "failed to parse response")
	}

	l.InfoContext(ctx, "got new access token", "accessToken", parsedResponse.AccessToken)

	gc.accessToken = parsedResponse.AccessToken
	return nil
}

type goCardlessListTransactionResponse struct {
	Transactions struct {
		Booked  []goCardlessListTransactionResponseTransaction `json:"booked"`
		Pending []goCardlessListTransactionResponseTransaction `json:"pending"`
	} `json:"transactions"`
	LastUpdated time.Time `json:"last_updated"`
}

type goCardlessListTransactionResponseTransaction struct {
	TransactionId     string `json:"transactionId"`
	BookingDate       string `json:"bookingDate"`
	ValueDate         string `json:"valueDate"`
	TransactionAmount struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"transactionAmount"`
	RemittanceInformationUnstructured string `json:"remittanceInformationUnstructured"`
	InternalTransactionId             string `json:"internalTransactionId"`
	DebtorName                        string `json:"debtorName"`
	CreditorName                      string `json:"creditorName"`
	AdditionalInformation             string `json:"additionalInformation"`
}

type Transaction struct {
	ID         string
	Date       time.Time
	AmountMili int64
	Memo       string
	Name       string
}

func (gc *GoCardless) ListTransactions(ctx context.Context, accountID string, from, to time.Time) ([]Transaction, error) {
	txn := newrelic.FromContext(ctx)
	seg := txn.StartSegment("listTransactions")
	defer seg.End()

	seg.AddAttribute("accountID", accountID)
	seg.AddAttribute("from", from)
	seg.AddAttribute("to", to)

	l := slog.Default().With("accountID", accountID, "from", from, "to", to)
	if gc.resetIn > 0 {
		l.Info("sleeping for reset in", "reset_in", gc.resetIn)
		return nil, nil
	}

	u := fmt.Sprintf("https://bankaccountdata.gocardless.com/api/v2/accounts/%s/transactions/", accountID)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create request: GET %s", u)
	}

	request.Header.Add("Authorization", "Bearer "+gc.accessToken)
	request.Header.Add("Accept", "application/json")

	queryParams := url.Values{}
	queryParams.Add("date_from", from.Format("2006-01-02"))
	queryParams.Add("date_to", to.Format("2006-01-02"))
	request.URL.RawQuery = queryParams.Encode()

	response, err := gc.httpClient.Do(request)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to make request: GET %s", u)
	}

	if response.StatusCode == http.StatusTooManyRequests {
		resetIn := time.Duration(0)
		accountReset := response.Header[http.CanonicalHeaderKey("http_x_ratelimit_account_success_reset")]
		if len(accountReset) > 0 && accountReset[0] != "0" {
			resetIn, err = time.ParseDuration(accountReset[0] + "s")
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse rate limit reset header: %s", response.Header[http.CanonicalHeaderKey("http_x_ratelimit_account_success_reset")][0])
			}
		}

		l.WarnContext(
			ctx,
			"too many requests",
			"status",
			response.Status,
			"http_x_ratelimit_limit",
			response.Header[http.CanonicalHeaderKey("http_x_ratelimit_limit")],
			http.CanonicalHeaderKey("http_x_ratelimit_remaining"),
			response.Header[http.CanonicalHeaderKey("http_x_ratelimit_remaining")],
			http.CanonicalHeaderKey("http_x_ratelimit_reset"),
			response.Header[http.CanonicalHeaderKey("http_x_ratelimit_reset")],
			http.CanonicalHeaderKey("http_x_ratelimit_account_success_limit"),
			response.Header[http.CanonicalHeaderKey("http_x_ratelimit_account_success_limit")],
			http.CanonicalHeaderKey("http_x_ratelimit_account_success_remaining"),
			response.Header[http.CanonicalHeaderKey("http_x_ratelimit_account_success_remaining")],
			http.CanonicalHeaderKey("http_x_ratelimit_account_success_reset"),
			response.Header[http.CanonicalHeaderKey("http_x_ratelimit_account_success_reset")],
			"reset_in",
			resetIn,
		)
		return nil, errors.Errorf("too many requests: %s", response.Status)
	}

	if response.StatusCode != 200 {
		l.WarnContext(ctx, "failed to list transactions", "status", response.Status, "headers", response.Header)
		return nil, errors.Errorf("failed to list transactions: %s", response.Status)
	}

	seg.AddAttribute("responseStatusCode", response.StatusCode)

	parsedResponse := goCardlessListTransactionResponse{}
	if err := json.NewDecoder(response.Body).Decode(&parsedResponse); err != nil {
		return nil, errors.Wrapf(err, "failed to parse response: GET %s", u)
	}

	transactions := toTransactions(parsedResponse)
	seg.AddAttribute("transactionsCount", len(transactions))
	l.InfoContext(ctx, "got transactions", "count", len(transactions))

	return transactions, nil
}

func toTransactions(response goCardlessListTransactionResponse) []Transaction {
	l := slog.Default()

	var transactions []Transaction
	for _, transaction := range response.Transactions.Booked {
		t, err := toTransaction(transaction)
		if err != nil {
			l.Warn("failed to parse transaction", "error", err)
			continue
		}

		transactions = append(transactions, t)
	}

	for _, transaction := range response.Transactions.Pending {
		t, err := toTransaction(transaction)
		if err != nil {
			l.Warn("failed to parse transaction", "error", err)
			continue
		}

		transactions = append(transactions, t)
	}

	return transactions
}

func toTransaction(goCardlessTransaction goCardlessListTransactionResponseTransaction) (Transaction, error) {
	l := slog.Default()
	valueDate, err := time.Parse("2006-01-02", goCardlessTransaction.ValueDate)
	if err != nil {
		l.Warn("failed to parse value date", "valueDate", goCardlessTransaction.ValueDate, "error", err)
		return Transaction{}, errors.Wrapf(err, "failed to parse value date: %s", goCardlessTransaction.ValueDate)
	}

	amount, err := strconv.ParseFloat(goCardlessTransaction.TransactionAmount.Amount, 64)
	if err != nil {
		l.Warn("failed to parse amount", "amount", goCardlessTransaction.TransactionAmount.Amount, "error", err)
		return Transaction{}, errors.Wrapf(err, "failed to parse amount: %s", goCardlessTransaction.TransactionAmount.Amount)
	}

	transaction := Transaction{
		ID:         toID(goCardlessTransaction),
		Date:       valueDate,
		AmountMili: int64(amount * 1000),
		Memo:       goCardlessTransaction.RemittanceInformationUnstructured,
		Name:       toName(goCardlessTransaction),
	}

	l.Info("gocardless transaction", "date", goCardlessTransaction.ValueDate, "amount", goCardlessTransaction.TransactionAmount.Amount, "memo", goCardlessTransaction.RemittanceInformationUnstructured, "name", toName(goCardlessTransaction), "debtor_name", goCardlessTransaction.DebtorName, "creditor_name", goCardlessTransaction.CreditorName, "additional_information", goCardlessTransaction.AdditionalInformation)

	return transaction, nil
}

func toID(transaction goCardlessListTransactionResponseTransaction) string {
	if transaction.TransactionId != "" {
		return transaction.TransactionId
	}
	return transaction.InternalTransactionId
}

func toName(transaction goCardlessListTransactionResponseTransaction) string {
	amountParsed, err := strconv.ParseFloat(transaction.TransactionAmount.Amount, 64)
	if err == nil {
		if amountParsed >= 0 && transaction.DebtorName != "" {
			return transaction.DebtorName
		}
		if amountParsed < 0 && transaction.CreditorName != "" {
			return transaction.CreditorName
		}
	}

	return transaction.RemittanceInformationUnstructured
}
