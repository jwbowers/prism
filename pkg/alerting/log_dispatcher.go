package alerting

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LogDispatcher is the default no-dependency AlertDispatcher.
// It writes every alert to stdout/stderr and keeps an in-memory history.
// Used when no external notification channels are configured.
type LogDispatcher struct {
	mu      sync.Mutex
	history []SentAlert
}

// NewLogDispatcher creates a LogDispatcher.
func NewLogDispatcher() *LogDispatcher {
	return &LogDispatcher{}
}

// Send writes the alert to stdout and records it in history.
func (d *LogDispatcher) Send(_ context.Context, alert Alert) error {
	icon := severityIcon(alert.Severity)
	fmt.Printf("%s [ALERT] %s — %s\n", icon, alert.Title, alert.Body)
	d.record(alert, true, "")
	return nil
}

// SendBatch sends multiple alerts sequentially.
func (d *LogDispatcher) SendBatch(ctx context.Context, alerts []Alert) error {
	for _, a := range alerts {
		if err := d.Send(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

// History returns the most recent sent alerts up to limit.
func (d *LogDispatcher) History(limit int) []SentAlert {
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

func (d *LogDispatcher) record(alert Alert, ok bool, errMsg string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.history = append(d.history, SentAlert{
		Alert:     alert,
		SentAt:    time.Now(),
		Succeeded: ok,
		Error:     errMsg,
	})
	// Keep last 200 entries.
	if len(d.history) > 200 {
		d.history = d.history[len(d.history)-200:]
	}
}

func severityIcon(s AlertSeverity) string {
	switch s {
	case AlertSeverityCritical:
		return "🚨"
	case AlertSeverityWarning:
		return "⚠️ "
	default:
		return "💡"
	}
}
