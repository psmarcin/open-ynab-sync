package main

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apimock "psmarcin.github.com/open-ynab-sync/cmd/link/api/apimock"
	"psmarcin.github.com/open-ynab-sync/cmd/link/auth"
	"psmarcin.github.com/open-ynab-sync/cmd/link/config"
	"psmarcin.github.com/open-ynab-sync/cmd/link/server"
)

// TestMainFlow tests the main flow of the application without actually running main()
func TestMainFlow(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Mock the GoCardless client
		mockClient := apimock.NewMockGoCardlessClient(t)

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

		// Create a callback server
		callbackServer := server.NewCallbackServer(cfg.Port, logger)

		// Create auth flow
		authFlow := auth.NewAuthFlow(mockClient, cfg, logger, callbackServer)
		authFlow.AutoOpenBrowser = false

		// Start the server
		err := callbackServer.Start(context.Background())
		assert.NoError(t, err)

		// Simulate callback
		go func() {
			time.Sleep(time.Millisecond)
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
	})
}
