package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/joho/godotenv"
)

// GoCardless API client for linking accounts
type GoCardless struct {
	SecretID     string
	SecretKey    string
	accessToken  string
	refreshToken string
	httpClient   *http.Client
}

// NewGoCardless creates a new GoCardless client
func NewGoCardless(secretID, secretKey string) *GoCardless {
	httpClient := http.DefaultClient
	httpClient.Timeout = 20 * time.Second

	return &GoCardless{
		SecretID:   secretID,
		SecretKey:  secretKey,
		httpClient: httpClient,
	}
}

// LogIn authenticates with the GoCardless API
func (gc *GoCardless) LogIn(ctx context.Context) error {
	type loginRequest struct {
		SecretID  string `json:"secret_id"`
		SecretKey string `json:"secret_key"`
	}

	type loginResponse struct {
		Access  string `json:"access"`
		Refresh string `json:"refresh"`
	}

	requestBody := loginRequest{SecretID: gc.SecretID, SecretKey: gc.SecretKey}
	resp, err := gc.makeRequest(ctx, "POST", "https://bankaccountdata.gocardless.com/api/v2/token/new/", requestBody, nil)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	var parsedResponse loginResponse
	if err := resp.parseJSON(&parsedResponse); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	gc.accessToken = parsedResponse.Access
	gc.refreshToken = parsedResponse.Refresh
	return nil
}

// CreateAgreement creates a new agreement with the institution
func (gc *GoCardless) CreateAgreement(ctx context.Context, institutionID string) (string, error) {
	type agreementRequest struct {
		InstitutionID      string   `json:"institution_id"`
		MaxHistoricalDays  string   `json:"max_historical_days"`
		AccessValidForDays string   `json:"access_valid_for_days"`
		AccessScope        []string `json:"access_scope"`
	}

	type agreementResponse struct {
		ID string `json:"id"`
	}

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
		return "", fmt.Errorf("failed to create agreement: %w", err)
	}

	var parsedResponse agreementResponse
	if err := resp.parseJSON(&parsedResponse); err != nil {
		return "", fmt.Errorf("failed to parse agreement response: %w", err)
	}

	return parsedResponse.ID, nil
}

// CreateRequisition creates a new requisition with the institution
func (gc *GoCardless) CreateRequisition(ctx context.Context, institutionID, agreementID, redirectURL string) (string, string, error) {
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
		return "", "", fmt.Errorf("failed to create requisition: %w", err)
	}

	var parsedResponse requisitionResponse
	if err := resp.parseJSON(&parsedResponse); err != nil {
		return "", "", fmt.Errorf("failed to parse requisition response: %w", err)
	}

	return parsedResponse.ID, parsedResponse.Link, nil
}

// GetRequisitionStatus gets the status of a requisition
func (gc *GoCardless) GetRequisitionStatus(ctx context.Context, requisitionID string) (string, []string, error) {
	type requisitionStatusResponse struct {
		ID       string   `json:"id"`
		Status   string   `json:"status"`
		Accounts []string `json:"accounts"`
	}

	resp, err := gc.makeRequest(ctx, "GET", "https://bankaccountdata.gocardless.com/api/v2/requisitions/"+requisitionID+"/", nil, map[string]string{
		"Authorization": "Bearer " + gc.accessToken,
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to get requisition status: %w", err)
	}

	var parsedResponse requisitionStatusResponse
	if err := resp.parseJSON(&parsedResponse); err != nil {
		return "", nil, fmt.Errorf("failed to parse requisition status response: %w", err)
	}

	return parsedResponse.Status, parsedResponse.Accounts, nil
}

// ListRequisitions lists all requisitions
func (gc *GoCardless) ListRequisitions(ctx context.Context) ([]Requisition, error) {
	type requisitionsResponse struct {
		Results []Requisition `json:"results"`
	}

	resp, err := gc.makeRequest(ctx, "GET", "https://bankaccountdata.gocardless.com/api/v2/requisitions/", nil, map[string]string{
		"Authorization": "Bearer " + gc.accessToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list requisitions: %w", err)
	}

	var parsedResponse requisitionsResponse
	if err := resp.parseJSON(&parsedResponse); err != nil {
		return nil, fmt.Errorf("failed to parse requisitions response: %w", err)
	}

	return parsedResponse.Results, nil
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
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return &apiResponse{
		statusCode: resp.StatusCode,
		body:       respBody,
	}, nil
}

// openBrowser opens a URL in the default browser
func openBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}

func main() {
	// Initialize logger
	l := slog.Default()

	// Load .env file
	err := godotenv.Load("./../.env")
	if err != nil {
		l.Error("failed to load .env file", "error", err)
		os.Exit(1)
	}

	// Parse command line arguments
	institutionID := flag.String("institution", "", "Institution ID (required)")
	port := flag.Int("port", 8080, "Port to listen for callback")
	flag.Parse()

	if *institutionID == "" {
		l.Error("institution ID is required")
		flag.Usage()
		os.Exit(1)
	}

	// Get GoCardless credentials from environment variables
	secretID := os.Getenv("GC_SECRET_ID")
	secretKey := os.Getenv("GC_SECRET_KEY")

	if secretID == "" || secretKey == "" {
		l.Error("GC_SECRET_ID and GC_SECRET_KEY environment variables are required")
		os.Exit(1)
	}

	// Create GoCardless client and log in
	gc := NewGoCardless(secretID, secretKey)
	ctx := context.Background()

	if err := gc.LogIn(ctx); err != nil {
		l.Error("failed to log in", "error", err)
		os.Exit(1)
	}

	l.Info("successfully logged in to GoCardless API")

	// Create agreement
	agreementID, err := gc.CreateAgreement(ctx, *institutionID)
	if err != nil {
		l.Error("failed to create agreement", "error", err)
		os.Exit(1)
	}

	l.Info("created agreement", "id", agreementID)

	var requisitionID, link string

	// Create requisition with redirect to localhost
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", *port)
	requisitionID, link, err = gc.CreateRequisition(ctx, *institutionID, agreementID, redirectURL)
	if err != nil {
		l.Error("failed to create requisition", "error", err)
		os.Exit(1)
	}
	l.Info("created requisition", "id", requisitionID)

	// Always go through the authorization flow to allow linking new accounts
	l.Info("opening authorization link in browser...", "link", link)

	// Open the link in the browser
	if err := openBrowser(link); err != nil {
		l.Warn("failed to open browser", "error", err)
	}

	// Set up a channel to receive the callback result
	callbackCh := make(chan struct{})

	// Set up HTTP server to listen for callback
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		l.Info("authorization callback received", "code", r.URL.Query().Get("code"))
		close(callbackCh)
	})

	// Start HTTP server
	server := &http.Server{Addr: fmt.Sprintf(":%d", *port)}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("failed to start HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for callback or timeout
	l.Info("waiting for authorization callback...")
	select {
	case <-callbackCh:
		l.Info("authorization callback received")
	case <-time.After(5 * time.Minute):
		l.Error("timed out waiting for authorization")
		os.Exit(1)
	}

	// Give the API some time to process the authorization
	time.Sleep(2 * time.Second)

	// Shutdown the server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		l.Info("Failed to shutdown HTTP server", "error", err)
	}

	// Check requisition status
	status, accounts, err := gc.GetRequisitionStatus(ctx, requisitionID)
	if err != nil {
		l.Error("failed to get requisition status", "error", err)
		os.Exit(1)
	}

	l.Info("requisition status", "status", status)

	if status != "LN" {
		l.Error("requisition is not in linked state (LN)")
		os.Exit(1)
	}

	if len(accounts) == 0 {
		l.Error("No accounts linked")
		os.Exit(1)
	}

	slog.Info("Successfully linked accounts:")
	for i, accountID := range accounts {
		l.Info("account", "id", accountID, "index", i+1)
	}
}
