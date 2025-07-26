package auth

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"time"

	"github.com/pkg/errors"

	"psmarcin.github.com/open-ynab-sync/cmd/link/api"
	"psmarcin.github.com/open-ynab-sync/cmd/link/config"
	"psmarcin.github.com/open-ynab-sync/cmd/link/server"
)

// AuthFlow manages the authorization flow
type AuthFlow struct {
	Client          api.GoCardlessClient
	Config          *config.Config
	Logger          *slog.Logger
	CallbackSrv     *server.CallbackServer
	AutoOpenBrowser bool
}

// NewAuthFlow creates a new authorization flow
func NewAuthFlow(client api.GoCardlessClient, cfg *config.Config, logger *slog.Logger, callbackSrv *server.CallbackServer) *AuthFlow {
	return &AuthFlow{
		Client:      client,
		Config:      cfg,
		Logger:      logger,
		CallbackSrv: callbackSrv,
	}
}

// Execute runs the complete authorization flow
func (a *AuthFlow) Execute(ctx context.Context) ([]string, error) {
	// Start the callback server
	if err := a.CallbackSrv.Start(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to start callback server")
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.CallbackSrv.Shutdown(shutdownCtx); err != nil {
			a.Logger.ErrorContext(ctx, "failed to shutdown callback server", "error", err)
		}
	}()

	// Log in to GoCardless
	if err := a.Client.LogIn(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to log in to GoCardless")
	}

	// Create agreement
	agreementID, err := a.Client.CreateAgreement(ctx, a.Config.InstitutionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create agreement")
	}

	// Create requisition
	redirectURL := a.CallbackSrv.GetCallbackURL()
	requisitionID, link, err := a.Client.CreateRequisition(ctx, a.Config.InstitutionID, agreementID, redirectURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create requisition")
	}

	if a.AutoOpenBrowser {
		// Open the link in the browser
		a.Logger.InfoContext(ctx, "opening authorization link in browser...", "link", link)
		if err := openBrowser(link); err != nil {
			a.Logger.WarnContext(ctx, "failed to open browser", "error", err)
		}
	} else {
		a.Logger.InfoContext(ctx, "authorization link", "link", link)
	}

	// Wait for callback
	if err := a.CallbackSrv.WaitForCallback(ctx, a.Config.AuthTimeout); err != nil {
		return nil, errors.Wrap(err, "authorization flow failed")
	}

	// Give the API some time to process the authorization
	time.Sleep(2 * time.Second)

	// Check requisition status
	status, accounts, err := a.Client.GetRequisitionStatus(ctx, requisitionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get requisition status")
	}

	if status != "LN" {
		return nil, errors.Errorf("requisition is not in linked state (LN), current state: %s", status)
	}

	if len(accounts) == 0 {
		return nil, errors.New("no accounts linked")
	}

	return accounts, nil
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
