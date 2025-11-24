package client

import (
	"context"
	"fmt"

	"github.com/scttfrdmn/prism/pkg/storage"
)

// TransferRequest represents a request to start a file transfer
type TransferRequest struct {
	Type      string                   `json:"type"`
	LocalPath string                   `json:"local_path"`
	S3Bucket  string                   `json:"s3_bucket"`
	S3Key     string                   `json:"s3_key"`
	Options   *storage.TransferOptions `json:"options,omitempty"`
}

// TransferResponse represents the response from a transfer request
type TransferResponse struct {
	TransferID string                    `json:"transfer_id"`
	Status     storage.TransferStatus    `json:"status"`
	Progress   *storage.TransferProgress `json:"progress,omitempty"`
	Error      string                    `json:"error,omitempty"`
}

// StartTransfer initiates a file transfer (upload or download)
func (c *HTTPClient) StartTransfer(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
	resp, err := c.makeRequest(ctx, "POST", "/api/v1/storage/transfer", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result TransferResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTransferStatus retrieves the current status of a transfer
func (c *HTTPClient) GetTransferStatus(ctx context.Context, transferID string) (*storage.TransferProgress, error) {
	url := fmt.Sprintf("/api/v1/storage/transfer/%s", transferID)
	resp, err := c.makeRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var progress storage.TransferProgress
	if err := c.handleResponse(resp, &progress); err != nil {
		return nil, err
	}

	return &progress, nil
}

// ListTransfers lists all active transfers
func (c *HTTPClient) ListTransfers(ctx context.Context) ([]*storage.TransferProgress, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/v1/storage/transfer", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var transfers []*storage.TransferProgress
	if err := c.handleResponse(resp, &transfers); err != nil {
		return nil, err
	}

	return transfers, nil
}

// CancelTransfer cancels a transfer in progress
func (c *HTTPClient) CancelTransfer(ctx context.Context, transferID string) error {
	url := fmt.Sprintf("/api/v1/storage/transfer/%s", transferID)
	resp, err := c.makeRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// For DELETE operations that return 200 OK, handleResponse with nil target is sufficient
	if err := c.handleResponse(resp, nil); err != nil {
		return err
	}

	return nil
}
