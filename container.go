package main

import (
	"fmt"
)

// ServiceContainer manages service instantiation and dependencies
type ServiceContainer struct {
	config         Config
	gcService      GoCardlessServicer
	ynabService    YNABServicer
	monitorService MonitoringServicer
	syncService    SynchronizationServicer
}

// NewServiceContainer creates a new service container with the given configuration
func NewServiceContainer(config Config) (*ServiceContainer, error) {
	container := &ServiceContainer{
		config: config,
	}

	// Initialize services
	if err := container.initializeServices(); err != nil {
		return nil, err
	}

	return container, nil
}

// initializeServices initializes all services in the container
func (c *ServiceContainer) initializeServices() error {
	// Initialize monitoring service first
	monitorService, err := c.createMonitoringService()
	if err != nil {
		return fmt.Errorf("failed to initialize monitoring service: %w", err)
	}
	c.monitorService = monitorService

	// Initialize GoCardless service
	c.gcService = c.createGoCardlessService()

	// Initialize YNAB service
	c.ynabService = c.createYNABService()

	// Initialize synchronization service
	c.syncService = c.createSyncService()

	return nil
}

// createMonitoringService creates a new monitoring service
func (c *ServiceContainer) createMonitoringService() (MonitoringServicer, error) {
	return NewMonitoringService(c.config.NewRelicAppName, c.config.NewRelicLicenseKey)
}

// createGoCardlessService creates a new GoCardless service
func (c *ServiceContainer) createGoCardlessService() GoCardlessServicer {
	return NewGoCardlessService(c.config.GCSecretID, c.config.GCSecretKey)
}

// createYNABService creates a new YNAB service
func (c *ServiceContainer) createYNABService() YNABServicer {
	return NewYNABService(c.config.YNABToken)
}

// createSyncService creates a new synchronization service
func (c *ServiceContainer) createSyncService() SynchronizationServicer {
	return NewSyncService(c.gcService, c.ynabService, c.monitorService, c.config.Jobs)
}

// Service getters

// GCService returns the GoCardless service
func (c *ServiceContainer) GCService() GoCardlessServicer {
	return c.gcService
}

// YNABServicer returns the YNAB service
func (c *ServiceContainer) YNABService() YNABServicer {
	return c.ynabService
}

// MonitorService returns the monitoring service
func (c *ServiceContainer) MonitorService() MonitoringServicer {
	return c.monitorService
}

// SyncService returns the synchronization service
func (c *ServiceContainer) SyncService() SynchronizationServicer {
	return c.syncService
}
