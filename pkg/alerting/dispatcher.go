// Package alerting provides a unified notification dispatcher for Prism.
//
// All budget alerts, shortfall warnings, and cushion triggers route through
// AlertDispatcher so the delivery backend can be swapped without changing callers.
// The default backend (LogDispatcher) always writes to stdout/stderr — no external
// dependencies required for basic operation.
package alerting

import (
	"context"
	"time"
)

// AlertSeverity classifies the urgency of an alert.
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// ChannelType identifies a notification delivery mechanism.
type ChannelType string

const (
	ChannelLog     ChannelType = "log" // always active — stdout/stderr
	ChannelEmail   ChannelType = "email"
	ChannelSlack   ChannelType = "slack"
	ChannelWebhook ChannelType = "webhook"
	ChannelDesktop ChannelType = "desktop" // OS desktop notification
)

// Channel defines a single notification destination.
type Channel struct {
	Type    ChannelType       `json:"type"`
	Target  string            `json:"target"` // email addr, webhook URL, etc.
	Options map[string]string `json:"options,omitempty"`
}

// Alert is the canonical notification payload.
type Alert struct {
	ID          string            `json:"id"`
	Severity    AlertSeverity     `json:"severity"`
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	ProjectID   string            `json:"project_id,omitempty"`
	ProjectName string            `json:"project_name,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Channels    []Channel         `json:"channels"`
	CreatedAt   time.Time         `json:"created_at"`
}

// SentAlert records a delivery attempt.
type SentAlert struct {
	Alert     Alert     `json:"alert"`
	SentAt    time.Time `json:"sent_at"`
	Succeeded bool      `json:"succeeded"`
	Error     string    `json:"error,omitempty"`
}

// AlertDispatcher is the unified notification interface.
type AlertDispatcher interface {
	Send(ctx context.Context, alert Alert) error
	SendBatch(ctx context.Context, alerts []Alert) error
	// History returns the most recent sent alerts (up to limit).
	History(limit int) []SentAlert
}
