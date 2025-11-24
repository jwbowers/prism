// Package daemon provides sleep/wake monitoring integration
package daemon

import (
	"context"
	"fmt"
	"log"

	"github.com/scttfrdmn/prism/pkg/idle"
	"github.com/scttfrdmn/prism/pkg/sleepwake"
)

// daemonInstanceManager implements the sleepwake.InstanceManager interface
// by wrapping the daemon's AWS manager and idle detection system
type daemonInstanceManager struct {
	server *Server
}

// newInstanceManager creates a new InstanceManager for the daemon
func newInstanceManager(server *Server) sleepwake.InstanceManager {
	return &daemonInstanceManager{
		server: server,
	}
}

// ListInstances returns the names of all running instances
func (m *daemonInstanceManager) ListInstances(ctx context.Context) ([]string, error) {
	instances, err := m.server.awsManager.ListInstances()
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	// Filter to only running instances and extract names
	var runningNames []string
	for _, instance := range instances {
		if instance.State == "running" {
			runningNames = append(runningNames, instance.Name)
		}
	}

	return runningNames, nil
}

// IsInstanceIdle checks if an instance is currently idle using the idle detection system
func (m *daemonInstanceManager) IsInstanceIdle(ctx context.Context, instanceName string) (bool, error) {
	// Get instance details to retrieve instance ID
	instances, err := m.server.awsManager.ListInstances()
	if err != nil {
		return false, fmt.Errorf("failed to list instances: %w", err)
	}

	var instanceID string
	for _, instance := range instances {
		if instance.Name == instanceName {
			instanceID = instance.ID
			break
		}
	}

	if instanceID == "" {
		return false, fmt.Errorf("instance not found: %s", instanceName)
	}

	// Create metrics collector
	awsConfig := m.server.awsManager.GetAWSConfig()
	metricsCollector := idle.NewMetricsCollector(awsConfig)

	// Use a default idle detection schedule (configurable via sleep/wake config)
	schedule := &idle.Schedule{
		IdleMinutes:      10, // Check last 10 minutes
		CPUThreshold:     5.0,
		MemoryThreshold:  10.0,
		NetworkThreshold: 1000.0, // 1KB/s
	}

	// Check if instance is idle
	isIdle, err := metricsCollector.IsInstanceIdle(ctx, instanceID, schedule)
	if err != nil {
		return false, fmt.Errorf("failed to check idle status for %s: %w", instanceName, err)
	}

	if isIdle {
		log.Printf("Instance %s is idle (CPU < %.1f%%, network < %.0f B/s over %d minutes)",
			instanceName, schedule.CPUThreshold, schedule.NetworkThreshold, schedule.IdleMinutes)
	} else {
		log.Printf("Instance %s is active", instanceName)
	}

	return isIdle, nil
}

// HibernateInstance hibernates a single instance
func (m *daemonInstanceManager) HibernateInstance(ctx context.Context, instanceName string) error {
	if err := m.server.awsManager.HibernateInstance(instanceName); err != nil {
		return fmt.Errorf("failed to hibernate instance %s: %w", instanceName, err)
	}

	log.Printf("Successfully hibernated instance: %s", instanceName)
	return nil
}

// ResumeInstance resumes a hibernated instance
func (m *daemonInstanceManager) ResumeInstance(ctx context.Context, instanceName string) error {
	if err := m.server.awsManager.StartInstance(instanceName); err != nil {
		return fmt.Errorf("failed to resume instance %s: %w", instanceName, err)
	}

	log.Printf("Successfully resumed instance: %s", instanceName)
	return nil
}
