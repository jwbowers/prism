package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

// fileOpsHTTP is the subset of HTTPClient used by file-ops commands.
type fileOpsHTTP interface {
	ListInstanceFiles(ctx context.Context, instanceName, path string) ([]interface{}, error)
	PushFileToInstance(ctx context.Context, instanceName, localPath, remotePath string) (map[string]interface{}, error)
	PullFileFromInstance(ctx context.Context, instanceName, remotePath, localPath string) (map[string]interface{}, error)
}

func (a *App) fileOpsList(instanceName, remotePath string) error {
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running — start with: prism admin daemon start")
	}
	hc, err := a.fileOpsClient()
	if err != nil {
		return err
	}
	entries, err := hc.ListInstanceFiles(a.ctx, instanceName, remotePath)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Println("No files found.")
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PERMISSIONS\tSIZE\tMODIFIED\tPATH")
	for _, e := range entries {
		m, _ := e.(map[string]interface{})
		perms, _ := m["permissions"].(string)
		size, _ := m["size_bytes"].(float64)
		mod, _ := m["modified_at"].(string)
		path, _ := m["path"].(string)
		isDir, _ := m["is_dir"].(bool)
		if isDir {
			path += "/"
		}
		fmt.Fprintf(tw, "%s\t%.0f\t%s\t%s\n", perms, size, mod, path)
	}
	return tw.Flush()
}

func (a *App) fileOpsPush(instanceName, localPath, remotePath string) error {
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running — start with: prism admin daemon start")
	}
	hc, err := a.fileOpsClient()
	if err != nil {
		return err
	}
	fmt.Printf("Pushing %s → %s:%s ...\n", localPath, instanceName, remotePath)
	result, err := hc.PushFileToInstance(a.ctx, instanceName, localPath, remotePath)
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
	return nil
}

func (a *App) fileOpsPull(instanceName, remotePath, localPath string) error {
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running — start with: prism admin daemon start")
	}
	hc, err := a.fileOpsClient()
	if err != nil {
		return err
	}
	fmt.Printf("Pulling %s:%s → %s ...\n", instanceName, remotePath, localPath)
	result, err := hc.PullFileFromInstance(a.ctx, instanceName, remotePath, localPath)
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
	return nil
}
