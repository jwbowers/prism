package fixtures

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/api/client"
)

// SSHCommand executes a command on a remote instance via SSH
func SSHCommand(t *testing.T, host, user, command string) (string, error) {
	t.Helper()

	sshCmd := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=10",
		fmt.Sprintf("%s@%s", user, host),
		command)

	output, err := sshCmd.CombinedOutput()
	if err != nil {
		t.Logf("SSH command failed: %s", command)
		t.Logf("Output: %s", string(output))
		return string(output), err
	}

	return string(output), nil
}

// WaitForInstanceState waits for an instance to reach the specified state
func WaitForInstanceState(t *testing.T, apiClient client.PrismAPI, instanceName, targetState string, timeout time.Duration) error {
	t.Helper()

	ctx := context.Background()
	deadline := time.Now().Add(timeout)

	t.Logf("Waiting for instance %s to reach state '%s' (timeout: %v)", instanceName, targetState, timeout)

	for time.Now().Before(deadline) {
		instance, err := apiClient.GetInstance(ctx, instanceName)
		if err != nil {
			t.Logf("Error getting instance status: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		currentState := strings.ToLower(instance.State)
		if currentState == strings.ToLower(targetState) {
			t.Logf("✓ Instance %s reached state '%s'", instanceName, targetState)
			return nil
		}

		t.Logf("Instance %s current state: %s (waiting for %s)", instanceName, currentState, targetState)
		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("timeout waiting for instance %s to reach state %s after %v", instanceName, targetState, timeout)
}

// WaitForSSHReady waits for SSH to be accessible on an instance
func WaitForSSHReady(t *testing.T, host, user string, timeout time.Duration) error {
	t.Helper()

	deadline := time.Now().Add(timeout)
	t.Logf("Waiting for SSH to be ready on %s (timeout: %v)", host, timeout)

	for time.Now().Before(deadline) {
		_, err := SSHCommand(t, host, user, "echo 'SSH ready'")
		if err == nil {
			t.Logf("✓ SSH ready on %s", host)
			return nil
		}

		t.Logf("SSH not ready yet on %s, retrying...", host)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for SSH to be ready on %s after %v", host, timeout)
}

// ExecuteRemoteTest runs a test command on a remote instance and returns the output
func ExecuteRemoteTest(t *testing.T, host, user, testCommand string) (string, error) {
	t.Helper()

	t.Logf("Executing remote test: %s", testCommand)
	output, err := SSHCommand(t, host, user, testCommand)
	if err != nil {
		t.Logf("Remote test failed: %v", err)
		return output, err
	}

	t.Logf("Remote test output: %s", output)
	return output, nil
}

// VerifyServiceRunning checks if a service is running on a remote instance
func VerifyServiceRunning(t *testing.T, host, user, serviceName string) bool {
	t.Helper()

	command := fmt.Sprintf("systemctl is-active %s || ps aux | grep -v grep | grep %s", serviceName, serviceName)
	_, err := SSHCommand(t, host, user, command)
	return err == nil
}

// VerifyPortListening checks if a port is listening on a remote instance
func VerifyPortListening(t *testing.T, host, user string, port int) bool {
	t.Helper()

	command := fmt.Sprintf("netstat -tuln | grep ':%d ' || ss -tuln | grep ':%d '", port, port)
	output, err := SSHCommand(t, host, user, command)
	return err == nil && strings.Contains(output, fmt.Sprintf(":%d", port))
}
