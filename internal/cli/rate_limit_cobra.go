package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// RateLimitCobraCommands provides Cobra commands for rate limiting management
type RateLimitCobraCommands struct {
	app *App
}

// NewRateLimitCobraCommands creates a new rate limit commands instance
func NewRateLimitCobraCommands(app *App) *RateLimitCobraCommands {
	return &RateLimitCobraCommands{app: app}
}

// CreateRateLimitCommand creates the rate-limit command group
func (r *RateLimitCobraCommands) CreateRateLimitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-limit",
		Short: "Manage workspace launch rate limiting",
		Long: `Configure and monitor rate limiting for workspace launches.

Rate limiting prevents AWS API throttling and accidental mass launches by limiting
the number of workspaces that can be launched within a time window.

Examples:
  prism admin rate-limit status        # View current rate limit status
  prism admin rate-limit configure     # Configure rate limits
  prism admin rate-limit reset         # Reset rate limit counters`,
	}

	cmd.AddCommand(r.createStatusCommand())
	cmd.AddCommand(r.createConfigureCommand())
	cmd.AddCommand(r.createResetCommand())

	return cmd
}

// createStatusCommand creates the status subcommand
func (r *RateLimitCobraCommands) createStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show rate limit status",
		Long:  `Display the current rate limit configuration and usage statistics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.handleStatus()
		},
	}
}

// createConfigureCommand creates the configure subcommand
func (r *RateLimitCobraCommands) createConfigureCommand() *cobra.Command {
	var maxLaunches int
	var windowMinutes int
	var enabled bool
	var setEnabled bool

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure rate limiting",
		Long: `Update rate limit configuration including maximum launches per time window.

Examples:
  prism admin rate-limit configure --max-launches 5 --window 1
  prism admin rate-limit configure --enabled=false
  prism admin rate-limit configure --max-launches 10 --window 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.handleConfigure(maxLaunches, windowMinutes, enabled, setEnabled)
		},
	}

	cmd.Flags().IntVar(&maxLaunches, "max-launches", 0, "Maximum launches per window (1-100)")
	cmd.Flags().IntVar(&windowMinutes, "window", 0, "Time window in minutes (1-60)")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable or disable rate limiting")
	cmd.Flags().BoolVar(&setEnabled, "set-enabled", false, "Set enabled flag")

	return cmd
}

// createResetCommand creates the reset subcommand
func (r *RateLimitCobraCommands) createResetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset rate limit counters",
		Long:  `Reset the rate limit counters, clearing all launch history.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.handleReset()
		},
	}
}

// handleStatus displays the current rate limit status
func (r *RateLimitCobraCommands) handleStatus() error {
	daemonURL := r.app.config.Daemon.URL

	resp, err := http.Get(daemonURL + "/api/v1/rate-limit/status")
	if err != nil {
		return fmt.Errorf("failed to get rate limit status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get rate limit status: %s", string(body))
	}

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Display status
	fmt.Println("🚦 Rate Limit Status")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Enabled:            %v\n", status["enabled"])
	fmt.Printf("Max Launches:       %v per %v minute(s)\n", status["max_launches"], status["window_minutes"])
	fmt.Printf("Current Launches:   %v\n", status["current_launches"])
	fmt.Printf("Remaining:          %v\n", status["remaining_launches"])
	fmt.Printf("Quota Used:         %.1f%%\n", status["quota_used_percent"])

	if resetTime, ok := status["reset_time"]; ok && resetTime != nil {
		fmt.Printf("Next Reset:         %v\n", resetTime)
		if seconds, ok := status["seconds_until_reset"]; ok {
			fmt.Printf("Time Until Reset:   %v seconds\n", seconds)
		}
	}

	return nil
}

// handleConfigure updates the rate limit configuration
func (r *RateLimitCobraCommands) handleConfigure(maxLaunches, windowMinutes int, enabled, setEnabled bool) error {
	daemonURL := r.app.config.Daemon.URL

	// Build configuration request
	config := make(map[string]interface{})
	if maxLaunches > 0 {
		config["max_launches"] = maxLaunches
	}
	if windowMinutes > 0 {
		config["window_minutes"] = windowMinutes
	}
	if setEnabled {
		config["enabled"] = enabled
	}

	if len(config) == 0 {
		return fmt.Errorf("no configuration changes specified")
	}

	// Convert to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to encode configuration: %w", err)
	}

	// Send configuration request
	resp, err := http.Post(daemonURL+"/api/v1/rate-limit/configure", "application/json", strings.NewReader(string(configJSON)))
	if err != nil {
		return fmt.Errorf("failed to configure rate limit: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to configure rate limit: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Println("✅ Rate limit configuration updated")
	fmt.Println(result["message"])

	return nil
}

// handleReset resets the rate limit counters
func (r *RateLimitCobraCommands) handleReset() error {
	daemonURL := r.app.config.Daemon.URL

	resp, err := http.Post(daemonURL+"/api/v1/rate-limit/reset", "application/json", strings.NewReader("{}"))
	if err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to reset rate limit: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Println("✅ Rate limit reset successfully")
	fmt.Printf("Configuration: %v launches per %v minute(s)\n",
		result["max_launches"], result["window_minutes"])

	return nil
}
