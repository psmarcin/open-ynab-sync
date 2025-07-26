package main

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"psmarcin.github.com/open-ynab-sync/cmd/link/api"
	"psmarcin.github.com/open-ynab-sync/cmd/link/auth"
	"psmarcin.github.com/open-ynab-sync/cmd/link/config"
	"psmarcin.github.com/open-ynab-sync/cmd/link/server"
)

// MockGoCardlessClient is a mock implementation of the GoCardlessClient interface
type MockGoCardlessClient struct {
	mock.Mock
}

func (m *MockGoCardlessClient) LogIn(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockGoCardlessClient) CreateAgreement(ctx context.Context, institutionID string) (string, error) {
	args := m.Called(ctx, institutionID)
	return args.String(0), args.Error(1)
}

func (m *MockGoCardlessClient) CreateRequisition(ctx context.Context, institutionID, agreementID, redirectURL string) (string, string, error) {
	args := m.Called(ctx, institutionID, agreementID, redirectURL)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockGoCardlessClient) GetRequisitionStatus(ctx context.Context, requisitionID string) (string, []string, error) {
	args := m.Called(ctx, requisitionID)
	return args.String(0), args.Get(1).([]string), args.Error(2)
}

func (m *MockGoCardlessClient) ListRequisitions(ctx context.Context) ([]api.Requisition, error) {
	args := m.Called(ctx)
	return args.Get(0).([]api.Requisition), args.Error(1)
}

// TestMainFlow tests the main flow of the application without actually running main()
func TestMainFlow(t *testing.T) {
	// This test verifies that the components are correctly orchestrated

	// Mock the GoCardless client
	mockClient := new(MockGoCardlessClient)

	// Setup expectations
	mockClient.On("LogIn", mock.Anything).Return(nil)
	mockClient.On("CreateAgreement", mock.Anything, "test-institution").Return("agreement-id", nil)
	mockClient.On("CreateRequisition", mock.Anything, "test-institution", "agreement-id", mock.Anything).
		Return("requisition-id", "http://auth-link", nil)
	mockClient.On("GetRequisitionStatus", mock.Anything, "requisition-id").
		Return("LN", []string{"account-id-1", "account-id-2"}, nil)

	// Create configuration
	cfg := &config.Config{
		GCSecretID:    "test-id",
		GCSecretKey:   "test-key",
		InstitutionID: "test-institution",
		Port:          8080,
		AuthTimeout:   100 * time.Millisecond,
		HTTPTimeout:   100 * time.Millisecond,
	}

	// Create logger
	logger := slog.Default()

	// Create callback server
	callbackServer := server.NewCallbackServer(cfg.Port, logger)

	// Create auth flow
	authFlow := auth.NewAuthFlow(mockClient, cfg, logger, callbackServer)

	// Start the server
	err := callbackServer.Start(context.Background())
	assert.NoError(t, err)

	// Simulate callback
	go func() {
		time.Sleep(50 * time.Millisecond)
		close(callbackServer.CallbackCh)
	}()

	// Execute the flow
	accounts, err := authFlow.Execute(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{"account-id-1", "account-id-2"}, accounts)

	// Shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = callbackServer.Shutdown(ctx)
	assert.NoError(t, err)

	// Verify all expectations were met
	mockClient.AssertExpectations(t)
}
