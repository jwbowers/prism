package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

// capBlockHTTP is the subset of HTTPClient used by capacity-block commands.
type capBlockHTTP interface {
	ListCapacityBlocks(ctx context.Context) ([]interface{}, error)
	ReserveCapacityBlock(ctx context.Context, req map[string]interface{}) (map[string]interface{}, error)
	DescribeCapacityBlock(ctx context.Context, id string) (map[string]interface{}, error)
	CancelCapacityBlock(ctx context.Context, id string) error
}

func (a *App) capacityBlockList() error {
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running — start with: prism admin daemon start")
	}
	hc, err := a.capacityBlockClient()
	if err != nil {
		return err
	}
	blocks, err := hc.ListCapacityBlocks(a.ctx)
	if err != nil {
		return err
	}
	if len(blocks) == 0 {
		fmt.Println("No capacity blocks found.")
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tTYPE\tCOUNT\tAZ\tSTATE\tSTART\tEND")
	for _, b := range blocks {
		m, _ := b.(map[string]interface{})
		id, _ := m["id"].(string)
		itype, _ := m["instance_type"].(string)
		count, _ := m["instance_count"].(float64)
		az, _ := m["availability_zone"].(string)
		state, _ := m["state"].(string)
		start, _ := m["start_time"].(string)
		end, _ := m["end_time"].(string)
		fmt.Fprintf(tw, "%s\t%s\t%.0f\t%s\t%s\t%s\t%s\n", id, itype, count, az, state, start, end)
	}
	return tw.Flush()
}

func (a *App) capacityBlockReserve(instanceType, az, start string, count, hours int) error {
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running — start with: prism admin daemon start")
	}
	hc, err := a.capacityBlockClient()
	if err != nil {
		return err
	}
	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return fmt.Errorf("invalid start time (use RFC3339, e.g. 2026-04-01T09:00:00Z): %w", err)
	}
	req := map[string]interface{}{
		"instance_type":  instanceType,
		"instance_count": count,
		"start_time":     startTime.Format(time.RFC3339),
		"duration_hours": hours,
	}
	if az != "" {
		req["availability_zone"] = az
	}
	result, err := hc.ReserveCapacityBlock(a.ctx, req)
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
	return nil
}

func (a *App) capacityBlockShow(id string) error {
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running — start with: prism admin daemon start")
	}
	hc, err := a.capacityBlockClient()
	if err != nil {
		return err
	}
	result, err := hc.DescribeCapacityBlock(a.ctx, id)
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
	return nil
}

func (a *App) capacityBlockCancel(id string) error {
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running — start with: prism admin daemon start")
	}
	hc, err := a.capacityBlockClient()
	if err != nil {
		return err
	}
	if err := hc.CancelCapacityBlock(a.ctx, id); err != nil {
		return err
	}
	fmt.Printf("Capacity block %s cancelled.\n", id)
	return nil
}
