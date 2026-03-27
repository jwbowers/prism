package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	apiclient "github.com/scttfrdmn/prism/pkg/api/client"
)

// workshopClient returns the underlying *HTTPClient for workshop commands.
func (a *App) workshopClient() (*apiclient.HTTPClient, error) {
	hc, ok := a.apiClient.(*apiclient.HTTPClient)
	if !ok {
		return nil, fmt.Errorf("workshop commands require daemon connection")
	}
	return hc, nil
}

// Workshop handles the 'prism workshop' command dispatcher.
func (a *App) Workshop(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism workshop <action> [args]")
	}
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running. Start with: prism admin daemon start")
	}
	hc, err := a.workshopClient()
	if err != nil {
		return err
	}
	action := args[0]
	rest := args[1:]
	switch action {
	case "list":
		return a.workshopList(hc, rest)
	case "create":
		return a.workshopCreate(hc, rest)
	case "show":
		return a.workshopShow(hc, rest)
	case "delete":
		return a.workshopDelete(hc, rest)
	case "provision":
		return a.workshopProvision(hc, rest)
	case "dashboard":
		return a.workshopDashboard(hc, rest)
	case "end":
		return a.workshopEnd(hc, rest)
	case "download":
		return a.workshopDownload(hc, rest)
	case "config":
		return a.workshopConfig(hc, rest)
	default:
		return fmt.Errorf("unknown workshop action: %s", action)
	}
}

func (a *App) workshopList(hc *apiclient.HTTPClient, args []string) error {
	params := ""
	for i := 0; i < len(args)-1; i++ {
		switch args[i] {
		case "--owner":
			params += "owner=" + args[i+1] + "&"
			i++
		case "--status":
			params += "status=" + args[i+1] + "&"
			i++
		}
	}
	if len(params) > 0 {
		params = params[:len(params)-1]
	}

	result, err := hc.ListWorkshops(a.ctx, params)
	if err != nil {
		return fmt.Errorf("failed to list workshops: %w", err)
	}

	ws, _ := result["workshops"].([]interface{})
	if len(ws) == 0 {
		fmt.Println("No workshops found.")
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tTITLE\tSTATUS\tSTART\tEND\tPARTICIPANTS")
	for _, raw := range ws {
		w, _ := raw.(map[string]interface{})
		participants, _ := w["participants"].([]interface{})
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\n",
			str(w["id"]),
			str(w["title"]),
			str(w["status"]),
			formatTimestamp(str(w["start_time"])),
			formatTimestamp(str(w["end_time"])),
			len(participants),
		)
	}
	return tw.Flush()
}

func (a *App) workshopCreate(hc *apiclient.HTTPClient, args []string) error {
	req := map[string]interface{}{}
	for i := 0; i < len(args)-1; i++ {
		switch args[i] {
		case "--title":
			req["title"] = args[i+1]
			i++
		case "--template":
			req["template"] = args[i+1]
			i++
		case "--owner":
			req["owner"] = args[i+1]
			i++
		case "--start":
			t, err := parseFlexibleTime(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid start time: %w", err)
			}
			req["start_time"] = t.UTC().Format(time.RFC3339)
			i++
		case "--end":
			t, err := parseFlexibleTime(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid end time: %w", err)
			}
			req["end_time"] = t.UTC().Format(time.RFC3339)
			i++
		case "--max-participants":
			n, _ := strconv.Atoi(args[i+1])
			req["max_participants"] = n
			i++
		case "--budget-per-participant":
			f, _ := strconv.ParseFloat(args[i+1], 64)
			req["budget_per_participant"] = f
			i++
		case "--early-access":
			n, _ := strconv.Atoi(args[i+1])
			req["early_access_hours"] = n
			i++
		case "--description":
			req["description"] = args[i+1]
			i++
		}
	}

	result, err := hc.CreateWorkshop(a.ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create workshop: %w", err)
	}
	fmt.Printf("Workshop created: %s\n", str(result["id"]))
	fmt.Printf("  Title:      %s\n", str(result["title"]))
	fmt.Printf("  Status:     %s\n", str(result["status"]))
	fmt.Printf("  Join Token: %s\n", str(result["join_token"]))
	return nil
}

func (a *App) workshopShow(hc *apiclient.HTTPClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism workshop show <id>")
	}
	result, err := hc.GetWorkshop(a.ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to get workshop: %w", err)
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
	return nil
}

func (a *App) workshopDelete(hc *apiclient.HTTPClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism workshop delete <id>")
	}
	if err := hc.DeleteWorkshop(a.ctx, args[0]); err != nil {
		return fmt.Errorf("failed to delete workshop: %w", err)
	}
	fmt.Printf("Workshop %s deleted.\n", args[0])
	return nil
}

func (a *App) workshopProvision(hc *apiclient.HTTPClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism workshop provision <id>")
	}
	result, err := hc.ProvisionWorkshop(a.ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to provision workshop: %w", err)
	}
	fmt.Printf("Provisioning complete:\n")
	fmt.Printf("  Provisioned: %v\n", result["provisioned"])
	fmt.Printf("  Skipped:     %v\n", result["skipped"])
	if errs, ok := result["errors"].([]interface{}); ok && len(errs) > 0 {
		fmt.Printf("  Errors (%d):\n", len(errs))
		for _, e := range errs {
			fmt.Printf("    - %v\n", e)
		}
	}
	return nil
}

func (a *App) workshopDashboard(hc *apiclient.HTTPClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism workshop dashboard <id>")
	}
	result, err := hc.GetWorkshopDashboard(a.ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to get dashboard: %w", err)
	}

	fmt.Printf("Workshop Dashboard: %s\n", str(result["title"]))
	fmt.Printf("  Status:          %s\n", str(result["status"]))
	fmt.Printf("  Total:           %v participants\n", result["total_participants"])
	fmt.Printf("  Active:          %v instances\n", result["active_instances"])
	fmt.Printf("  Stopped:         %v instances\n", result["stopped_instances"])
	fmt.Printf("  Pending:         %v instances\n", result["pending_instances"])
	fmt.Printf("  Time Remaining:  %s\n", str(result["time_remaining"]))
	fmt.Printf("  Total Spent:     $%.2f\n", floatVal(result["total_spent"]))

	if paxRaw, ok := result["participants"].([]interface{}); ok && len(paxRaw) > 0 {
		fmt.Println("\nParticipants:")
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  USER ID\tSTATUS\tINSTANCE")
		for _, raw := range paxRaw {
			p, _ := raw.(map[string]interface{})
			fmt.Fprintf(tw, "  %s\t%s\t%s\n",
				str(p["user_id"]),
				str(p["status"]),
				str(p["instance_name"]),
			)
		}
		_ = tw.Flush()
	}
	return nil
}

func (a *App) workshopEnd(hc *apiclient.HTTPClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism workshop end <id>")
	}
	result, err := hc.EndWorkshop(a.ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to end workshop: %w", err)
	}
	fmt.Printf("Workshop ended. Stopped %v instances.\n", result["stopped"])
	if errs, ok := result["errors"].([]interface{}); ok && len(errs) > 0 {
		fmt.Printf("Errors (%d):\n", len(errs))
		for _, e := range errs {
			fmt.Printf("  - %v\n", e)
		}
	}
	return nil
}

func (a *App) workshopDownload(hc *apiclient.HTTPClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism workshop download <id>")
	}
	result, err := hc.GetWorkshopDownload(a.ctx, args[0])
	if err != nil {
		return fmt.Errorf("failed to get download manifest: %w", err)
	}

	fmt.Printf("Download Manifest — Workshop %s\n\n", str(result["workshop_id"]))
	pax, _ := result["participants"].([]interface{})
	for _, raw := range pax {
		p, _ := raw.(map[string]interface{})
		fmt.Printf("  [%s] %s\n    %s\n\n",
			str(p["user_id"]),
			str(p["display_name"]),
			str(p["download_note"]),
		)
	}
	return nil
}

func (a *App) workshopConfig(hc *apiclient.HTTPClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism workshop config <save|list|use> [args]")
	}
	switch args[0] {
	case "save":
		return a.workshopConfigSave(hc, args[1:])
	case "list":
		return a.workshopConfigList(hc)
	case "use":
		return a.workshopConfigUse(hc, args[1:])
	default:
		return fmt.Errorf("unknown config subcommand: %s", args[0])
	}
}

func (a *App) workshopConfigSave(hc *apiclient.HTTPClient, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: prism workshop config save <workshop-id> <config-name>")
	}
	result, err := hc.SaveWorkshopConfig(a.ctx, args[0], args[1])
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("Config saved: %s (template: %s, %dh, max %d pax)\n",
		str(result["name"]), str(result["template"]),
		intVal(result["duration_hours"]), intVal(result["max_participants"]))
	return nil
}

func (a *App) workshopConfigList(hc *apiclient.HTTPClient) error {
	result, err := hc.ListWorkshopConfigs(a.ctx)
	if err != nil {
		return fmt.Errorf("failed to list configs: %w", err)
	}
	configs, _ := result["configs"].([]interface{})
	if len(configs) == 0 {
		fmt.Println("No saved workshop configs.")
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tTEMPLATE\tDURATION\tMAX PAX\tBUDGET/PAX")
	for _, raw := range configs {
		c, _ := raw.(map[string]interface{})
		fmt.Fprintf(tw, "%s\t%s\t%dh\t%d\t$%.2f\n",
			str(c["name"]), str(c["template"]),
			intVal(c["duration_hours"]), intVal(c["max_participants"]),
			floatVal(c["budget_per_participant"]),
		)
	}
	return tw.Flush()
}

func (a *App) workshopConfigUse(hc *apiclient.HTTPClient, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: prism workshop config use <config-name> --title <title> --start <time>")
	}
	configName := args[0]
	req := map[string]interface{}{}
	for i := 1; i < len(args)-1; i++ {
		switch args[i] {
		case "--title":
			req["title"] = args[i+1]
			i++
		case "--start":
			t, err := parseFlexibleTime(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid start time: %w", err)
			}
			req["start_time"] = t.UTC().Format(time.RFC3339)
			i++
		case "--owner":
			req["owner"] = args[i+1]
			i++
		}
	}

	result, err := hc.CreateWorkshopFromConfig(a.ctx, configName, req)
	if err != nil {
		return fmt.Errorf("failed to create workshop from config: %w", err)
	}
	fmt.Printf("Workshop created from config %q: %s\n", configName, str(result["id"]))
	fmt.Printf("  Title:      %s\n", str(result["title"]))
	fmt.Printf("  Join Token: %s\n", str(result["join_token"]))
	return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func str(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func intVal(v interface{}) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}

func itoa(n int) string     { return strconv.Itoa(n) }
func ftoa(f float64) string { return strconv.FormatFloat(f, 'f', 2, 64) }

// parseFlexibleTime parses RFC3339 or "YYYY-MM-DDTHH:MM:SS" format.
func parseFlexibleTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unrecognized time format %q (use RFC3339 or YYYY-MM-DDTHH:MM:SS)", s)
}

// formatTimestamp shortens an RFC3339 timestamp for tabular display.
func formatTimestamp(s string) string {
	if len(s) >= 16 {
		return s[:16]
	}
	return s
}
