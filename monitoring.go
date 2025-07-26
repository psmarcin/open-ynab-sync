package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// NewRelicMonitoring implements the MonitoringServicer interface using New Relic
type NewRelicMonitoring struct {
	app *newrelic.Application
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(appName, licenseKey string) (MonitoringServicer, error) {
	// If license key is empty, return a no-op implementation
	if licenseKey == "" {
		return &NoOpMonitoring{}, nil
	}

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(licenseKey),
		newrelic.ConfigAppLogMetricsEnabled(true),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		return nil, err
	}

	// Wait for connection to New Relic
	if err := app.WaitForConnection(time.Second * 30); err != nil {
		slog.Error("failed to connect to New Relic", "error", err)
		return nil, err
	}

	return &NewRelicMonitoring{app: app}, nil
}

// StartTransaction starts a new transaction
func (m *NewRelicMonitoring) StartTransaction(name string) *newrelic.Transaction {
	return m.app.StartTransaction(name, newrelic.WithFunctionLocation())
}

// RecordError records an error in the transaction
func (m *NewRelicMonitoring) RecordError(txn *newrelic.Transaction, err error) {
	txn.NoticeError(err)
}

// AddAttribute adds an attribute to the transaction
func (m *NewRelicMonitoring) AddAttribute(txn *newrelic.Transaction, key string, value interface{}) {
	txn.AddAttribute(key, value)
}

// NewContext creates a new context with the transaction
func (m *NewRelicMonitoring) NewContext(ctx context.Context, txn *newrelic.Transaction) context.Context {
	return newrelic.NewContext(ctx, txn)
}

// FromContext gets the transaction from the context
func (m *NewRelicMonitoring) FromContext(ctx context.Context) *newrelic.Transaction {
	return newrelic.FromContext(ctx)
}

// StartSegment starts a new segment in the transaction
func (m *NewRelicMonitoring) StartSegment(txn *newrelic.Transaction, name string) *newrelic.Segment {
	return txn.StartSegment(name)
}

// Shutdown shuts down the New Relic application
func (m *NewRelicMonitoring) Shutdown(timeout time.Duration) {
	m.app.Shutdown(timeout)
}

// NoOpMonitoring is a no-op implementation of the MonitoringServicer interface
type NoOpMonitoring struct{}

// StartTransaction starts a new transaction (no-op)
func (m *NoOpMonitoring) StartTransaction(name string) *newrelic.Transaction {
	return nil
}

// RecordError records an error in the transaction (no-op)
func (m *NoOpMonitoring) RecordError(txn *newrelic.Transaction, err error) {
	// No-op
}

// AddAttribute adds an attribute to the transaction (no-op)
func (m *NoOpMonitoring) AddAttribute(txn *newrelic.Transaction, key string, value interface{}) {
	// No-op
}

// NewContext creates a new context with the transaction (no-op)
func (m *NoOpMonitoring) NewContext(ctx context.Context, txn *newrelic.Transaction) context.Context {
	return ctx
}

// FromContext gets the transaction from the context (no-op)
func (m *NoOpMonitoring) FromContext(ctx context.Context) *newrelic.Transaction {
	return nil
}

// StartSegment starts a new segment in the transaction (no-op)
func (m *NoOpMonitoring) StartSegment(txn *newrelic.Transaction, name string) *newrelic.Segment {
	return nil
}

// Shutdown shuts down the New Relic application (no-op)
func (m *NoOpMonitoring) Shutdown(timeout time.Duration) {
	// No-op
}
