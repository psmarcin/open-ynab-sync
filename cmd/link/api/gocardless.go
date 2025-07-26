package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// GoCardlessClient defines the interface for interacting with the GoCardless API
type GoCardlessClient interface {
	LogIn(ctx context.Context) error
	CreateAgreement(ctx context.Context, institutionID string) (string, error)
	CreateRequisition(ctx context.Context, institutionID, agreementID, redirectURL string) (string, string, error)
	GetRequisitionStatus(ctx context.Context, requisitionID string) (string, []string, error)
	ListRequisitions(ctx context.Context) ([]Requisition, error)
}

// GoCardless implements the GoCardlessClient interface
type GoCardless struct {
	SecretID     string
	SecretKey    string
	accessToken  string
	refreshToken string
	httpClient   *http.Client
	logger       *slog.Logger
}

// NewGoCardless creates a new GoCardless client
func NewGoCardless(secretID, secretKey string, timeout time.Duration, logger *slog.Logger) *GoCardless {
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &GoCardless{
		SecretID:   secretID,
		SecretKey:  secretKey,
		httpClient: httpClient,
		logger:     logger,
	}
}

// Requisition represents a GoCardless requisition
type Requisition struct {
	ID            string   `json:"id"`
	Status        string   `json:"status"`
	InstitutionID string   `json:"institution_id"`
	Agreement     string   `json:"agreement"`
	Link          string   `json:"link"`
	Accounts      []string `json:"accounts"`
}

type loginRequest struct {
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
}

type loginResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

// LogIn authenticates with the GoCardless API
func (gc *GoCardless) LogIn(ctx context.Context) error {
	requestBody := loginRequest{SecretID: gc.SecretID, SecretKey: gc.SecretKey}
	resp, err := gc.makeRequest(ctx, "POST", "https://bankaccountdata.gocardless.com/api/v2/token/new/", requestBody, nil)
	if err != nil {
		return errors.Wrap(err, "failed to login")
	}

	var parsedResponse loginResponse
	if err := resp.parseJSON(&parsedResponse); err != nil {
		return errors.Wrap(err, "failed to parse login response")
	}

	gc.accessToken = parsedResponse.Access
	gc.refreshToken = parsedResponse.Refresh
	gc.logger.InfoContext(ctx, "successfully logged in to GoCardless API")
	return nil
}

type agreementRequest struct {
	InstitutionID      string   `json:"institution_id"`
	MaxHistoricalDays  string   `json:"max_historical_days"`
	AccessValidForDays string   `json:"access_valid_for_days"`
	AccessScope        []string `json:"access_scope"`
}

type agreementResponse struct {
	ID string `json:"id"`
}

// CreateAgreement creates a new agreement with the institution
func (gc *GoCardless) CreateAgreement(ctx context.Context, institutionID string) (string, error) {
	requestBody := agreementRequest{
		InstitutionID:      institutionID,
		MaxHistoricalDays:  "90",
		AccessValidForDays: "179",
		AccessScope:        []string{"balances", "details", "transactions"},
	}

	resp, err := gc.makeRequest(ctx, "POST", "https://bankaccountdata.gocardless.com/api/v2/agreements/enduser/", requestBody, map[string]string{
		"Authorization": "Bearer " + gc.accessToken,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to create agreement")
	}

	var parsedResponse agreementResponse
	if err := resp.parseJSON(&parsedResponse); err != nil {
		return "", errors.Wrap(err, "failed to parse agreement response")
	}

	gc.logger.InfoContext(ctx, "created agreement", "id", parsedResponse.ID)
	return parsedResponse.ID, nil
}

type requisitionRequest struct {
	Redirect      string `json:"redirect"`
	InstitutionID string `json:"institution_id"`
	Reference     string `json:"reference"`
	Agreement     string `json:"agreement"`
	UserLanguage  string `json:"user_language"`
}

type requisitionResponse struct {
	ID    string `json:"id"`
	Link  string `json:"link"`
	State string `json:"status"`
}

// CreateRequisition creates a new requisition with the institution
func (gc *GoCardless) CreateRequisition(ctx context.Context, institutionID, agreementID, redirectURL string) (string, string, error) {
	requestBody := requisitionRequest{
		Redirect:      redirectURL,
		InstitutionID: institutionID,
		Reference:     generateReference(),
		Agreement:     agreementID,
		UserLanguage:  "EN",
	}

	resp, err := gc.makeRequest(ctx, "POST", "https://bankaccountdata.gocardless.com/api/v2/requisitions/", requestBody, map[string]string{
		"Authorization": "Bearer " + gc.accessToken,
	})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create requisition")
	}

	var parsedResponse requisitionResponse
	if err := resp.parseJSON(&parsedResponse); err != nil {
		return "", "", errors.Wrap(err, "failed to parse requisition response")
	}

	gc.logger.InfoContext(ctx, "created requisition", "id", parsedResponse.ID)
	return parsedResponse.ID, parsedResponse.Link, nil
}

type requisitionStatusResponse struct {
	ID       string   `json:"id"`
	Status   string   `json:"status"`
	Accounts []string `json:"accounts"`
}

// GetRequisitionStatus gets the status of a requisition
func (gc *GoCardless) GetRequisitionStatus(ctx context.Context, requisitionID string) (string, []string, error) {
	resp, err := gc.makeRequest(ctx, "GET", "https://bankaccountdata.gocardless.com/api/v2/requisitions/"+requisitionID+"/", nil, map[string]string{
		"Authorization": "Bearer " + gc.accessToken,
	})
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to get requisition status")
	}

	var parsedResponse requisitionStatusResponse
	if err := resp.parseJSON(&parsedResponse); err != nil {
		return "", nil, errors.Wrap(err, "failed to parse requisition status response")
	}

	gc.logger.InfoContext(ctx, "requisition status", "status", parsedResponse.Status, "accounts", len(parsedResponse.Accounts))
	return parsedResponse.Status, parsedResponse.Accounts, nil
}

type requisitionsResponse struct {
	Results []Requisition `json:"results"`
}

// ListRequisitions lists all requisitions
func (gc *GoCardless) ListRequisitions(ctx context.Context) ([]Requisition, error) {
	resp, err := gc.makeRequest(ctx, "GET", "https://bankaccountdata.gocardless.com/api/v2/requisitions/", nil, map[string]string{
		"Authorization": "Bearer " + gc.accessToken,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list requisitions")
	}

	var parsedResponse requisitionsResponse
	if err := resp.parseJSON(&parsedResponse); err != nil {
		return nil, errors.Wrap(err, "failed to parse requisitions response")
	}

	gc.logger.InfoContext(ctx, "listed requisitions", "count", len(parsedResponse.Results))
	return parsedResponse.Results, nil
}

// Helper types and methods for HTTP requests
type apiResponse struct {
	statusCode int
	body       []byte
}

func (r *apiResponse) parseJSON(v interface{}) error {
	return json.Unmarshal(r.body, v)
}

// generateReference creates a random reference string
func generateReference() string {
	return fmt.Sprintf("REF_%d", time.Now().UnixNano())
}

func (gc *GoCardless) makeRequest(ctx context.Context, method, url string, body interface{}, headers map[string]string) (*apiResponse, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal request body")
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Add("Accept", "application/json")
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return &apiResponse{
		statusCode: resp.StatusCode,
		body:       respBody,
	}, nil
}
