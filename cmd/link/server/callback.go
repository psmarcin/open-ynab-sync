package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// CallbackServer handles the OAuth callback
type CallbackServer struct {
	Port       int
	Logger     *slog.Logger
	CallbackCh chan struct{}
	server     *http.Server
}

// NewCallbackServer creates a new callback server
func NewCallbackServer(port int, logger *slog.Logger) *CallbackServer {
	return &CallbackServer{
		Port:       port,
		Logger:     logger,
		CallbackCh: make(chan struct{}),
	}
}

// Start starts the callback server
func (s *CallbackServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
	}

	s.Logger.InfoContext(ctx, "starting callback server", "port", s.Port)

	// Start the server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Logger.ErrorContext(ctx, "callback server error", "error", err)
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the server
func (s *CallbackServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.Logger.InfoContext(ctx, "shutting down callback server")
	return s.server.Shutdown(ctx)
}

// WaitForCallback waits for the callback or times out
func (s *CallbackServer) WaitForCallback(ctx context.Context, timeout time.Duration) error {
	s.Logger.InfoContext(ctx, "waiting for authorization callback...", "timeout", timeout)

	select {
	case <-s.CallbackCh:
		s.Logger.InfoContext(ctx, "authorization callback received")
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timed out waiting for authorization")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetCallbackURL returns the callback URL
func (s *CallbackServer) GetCallbackURL() string {
	return fmt.Sprintf("http://localhost:%d/callback", s.Port)
}

// handleCallback handles the callback request
func (s *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	s.Logger.Info("authorization callback received", "code", r.URL.Query().Get("code"))

	// Write a simple response
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<html><body><h1>Authorization Successful</h1><p>You can now close this window and return to the application.</p></body></html>"))

	// Signal that the callback has been received
	close(s.CallbackCh)
}
