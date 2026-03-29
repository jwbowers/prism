package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

// storageAnalyticsHTTP is the subset of HTTPClient used by analytics commands.
type storageAnalyticsHTTP interface {
	GetAllStorageAnalytics(ctx context.Context, period string) (map[string]interface{}, error)
	GetStorageAnalytics(ctx context.Context, name, period string) (map[string]interface{}, error)
}

func (a *App) storageAnalyticsClient() (storageAnalyticsHTTP, error) {
	hc, ok := a.apiClient.(storageAnalyticsHTTP)
	if !ok {
		return nil, fmt.Errorf("API client does not support storage analytics")
	}
	return hc, nil
}

func (a *App) storageAnalytics(name, period string) error {
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running — start with: prism admin daemon start")
	}
	hc, err := a.storageAnalyticsClient()
	if err != nil {
		return err
	}

	if name != "" {
		return a.storageAnalyticsOne(hc, name, period)
	}
	return a.storageAnalyticsAll(hc, period)
}

func (a *App) storageAnalyticsAll(hc storageAnalyticsHTTP, period string) error {
	result, err := hc.GetAllStorageAnalytics(a.ctx, period)
	if err != nil {
		return err
	}

	// Try to display as a table if resources array is present
	resources, hasResources := result["resources"].([]interface{})
	if !hasResources || len(resources) == 0 {
		fmt.Println("No storage analytics data available.")
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tTYPE\tTOTAL COST\tDAILY COST")
	for _, r := range resources {
		m, _ := r.(map[string]interface{})
		name, _ := m["storage_name"].(string)
		stype, _ := m["type"].(string)
		total, _ := m["total_cost"].(float64)
		daily, _ := m["daily_cost"].(float64)
		fmt.Fprintf(tw, "%s\t%s\t$%.4f\t$%.4f\n", name, stype, total, daily)
	}
	return tw.Flush()
}

func (a *App) storageAnalyticsOne(hc storageAnalyticsHTTP, name, period string) error {
	result, err := hc.GetStorageAnalytics(a.ctx, name, period)
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
	return nil
}
