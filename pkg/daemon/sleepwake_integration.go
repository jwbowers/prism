// Package daemon provides sleep/wake monitoring integration
package daemon

import (
	"context"
	"fmt"
	"log"

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

	// Idle detection is now handled on-instance by spored (#588).
	// The daemon cannot accurately determine idle status from the control plane.
	// spored monitors CPU, network, disk I/O, GPU, terminals, users, and recent
	// activity directly on the instance. Return false (active) as safe default.
	_ = instances
	log.Printf("IsInstanceIdle for %s: delegated to spored on-instance agent", instanceName)
	return false, nil
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
