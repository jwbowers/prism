package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// WebhookDispatcher sends alerts to one or more HTTP endpoints (Slack, generic webhook).
// Falls back to LogDispatcher for any channel that fails after retries.
type WebhookDispatcher struct {
	client     *http.Client
	channels   []Channel
	retryCount int
	fallback   *LogDispatcher
	mu         sync.Mutex
	history    []SentAlert
}

// WebhookConfig configures the WebhookDispatcher.
type WebhookConfig struct {
	Channels   []Channel
	RetryCount int           // default 2
	Timeout    time.Duration // default 10s
}

// NewWebhookDispatcher creates a dispatcher that delivers to all configured channels.
func NewWebhookDispatcher(cfg WebhookConfig) *WebhookDispatcher {
	if cfg.RetryCount == 0 {
		cfg.RetryCount = 2
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &WebhookDispatcher{
		client:     &http.Client{Timeout: timeout},
		channels:   cfg.Channels,
		retryCount: cfg.RetryCount,
		fallback:   NewLogDispatcher(),
	}
}

// Send delivers the alert to all configured channels.
func (d *WebhookDispatcher) Send(ctx context.Context, alert Alert) error {
	// Always log locally via fallback.
	_ = d.fallback.Send(ctx, alert)

	var lastErr error
	for _, ch := range d.channels {
		if err := d.sendToChannel(ctx, alert, ch); err != nil {
			lastErr = err
			d.record(alert, false, err.Error())
		} else {
			d.record(alert, true, "")
		}
	}
	return lastErr
}

// SendBatch sends multiple alerts.
func (d *WebhookDispatcher) SendBatch(ctx context.Context, alerts []Alert) error {
	var lastErr error
	for _, a := range alerts {
		if err := d.Send(ctx, a); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// History returns the most recent sent alerts.
func (d *WebhookDispatcher) History(limit int) []SentAlert {
	d.mu.Lock()
	defer d.mu.Unlock()
	if limit <= 0 || limit > len(d.history) {
		limit = len(d.history)
	}
	start := len(d.history) - limit
	result := make([]SentAlert, limit)
	copy(result, d.history[start:])
	return result
}

func (d *WebhookDispatcher) sendToChannel(ctx context.Context, alert Alert, ch Channel) error {
	switch ch.Type {
	case ChannelSlack:
		return d.sendSlack(ctx, alert, ch.Target)
	case ChannelWebhook:
		return d.sendWebhook(ctx, alert, ch.Target)
	case ChannelEmail:
		// Email requires SMTP setup; log and skip gracefully.
		fmt.Printf("📧 Email alert (unconfigured SMTP): %s — %s\n", alert.Title, alert.Body)
		return nil
	default:
		return nil
	}
}

type slackPayload struct {
	Text string `json:"text"`
}

func (d *WebhookDispatcher) sendSlack(ctx context.Context, alert Alert, webhookURL string) error {
	icon := severityIcon(alert.Severity)
	payload := slackPayload{Text: fmt.Sprintf("%s *%s*\n%s", icon, alert.Title, alert.Body)}
	return d.postJSON(ctx, webhookURL, payload)
}

func (d *WebhookDispatcher) sendWebhook(ctx context.Context, alert Alert, url string) error {
	return d.postJSON(ctx, url, alert)
}

func (d *WebhookDispatcher) postJSON(ctx context.Context, url string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal alert: %w", err)
	}
	var lastErr error
	for i := 0; i <= d.retryCount; i++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		resp, err := d.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}
	return lastErr
}

func (d *WebhookDispatcher) record(alert Alert, ok bool, errMsg string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.history = append(d.history, SentAlert{
		Alert:     alert,
		SentAt:    time.Now(),
		Succeeded: ok,
		Error:     errMsg,
	})
	if len(d.history) > 200 {
		d.history = d.history[len(d.history)-200:]
	}
}
