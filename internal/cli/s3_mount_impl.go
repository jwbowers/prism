package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

// s3MountHTTP is the subset of HTTPClient used by S3 mount commands.
type s3MountHTTP interface {
	ListInstanceS3Mounts(ctx context.Context, instanceName string) ([]interface{}, error)
	MountS3Bucket(ctx context.Context, instanceName, bucket, mountPath, method string, readOnly bool) (map[string]interface{}, error)
	UnmountS3Bucket(ctx context.Context, instanceName, mountPath string) error
}

func (a *App) s3MountClient() (s3MountHTTP, error) {
	hc, ok := a.apiClient.(s3MountHTTP)
	if !ok {
		return nil, fmt.Errorf("API client does not support S3 mount operations")
	}
	return hc, nil
}

func (a *App) s3MountList(instanceName string) error {
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running — start with: prism admin daemon start")
	}
	hc, err := a.s3MountClient()
	if err != nil {
		return err
	}
	mounts, err := hc.ListInstanceS3Mounts(a.ctx, instanceName)
	if err != nil {
		return err
	}
	if len(mounts) == 0 {
		fmt.Printf("No S3 mounts on %s.\n", instanceName)
		return nil
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "BUCKET\tMOUNT PATH\tMETHOD\tREAD-ONLY")
	for _, m := range mounts {
		entry, _ := m.(map[string]interface{})
		bucket, _ := entry["bucket_name"].(string)
		path, _ := entry["mount_path"].(string)
		method, _ := entry["mount_method"].(string)
		ro, _ := entry["read_only"].(bool)
		roStr := "no"
		if ro {
			roStr = "yes"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", bucket, path, method, roStr)
	}
	return tw.Flush()
}

func (a *App) s3MountMount(instanceName, bucket, mountPath, method string, readOnly bool) error {
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running — start with: prism admin daemon start")
	}
	hc, err := a.s3MountClient()
	if err != nil {
		return err
	}
	fmt.Printf("Mounting s3://%s on %s:%s ...\n", bucket, instanceName, mountPath)
	result, err := hc.MountS3Bucket(a.ctx, instanceName, bucket, mountPath, method, readOnly)
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
	return nil
}

func (a *App) s3MountUnmount(instanceName, mountPath string) error {
	if err := a.apiClient.Ping(a.ctx); err != nil {
		return fmt.Errorf("daemon not running — start with: prism admin daemon start")
	}
	hc, err := a.s3MountClient()
	if err != nil {
		return err
	}
	if err := hc.UnmountS3Bucket(a.ctx, instanceName, mountPath); err != nil {
		return err
	}
	fmt.Printf("Unmounted %s from %s.\n", mountPath, instanceName)
	return nil
}
