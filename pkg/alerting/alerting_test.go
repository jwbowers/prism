package alerting

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── FormatBudgetThresholdAlert ─────────────────────────────────────────────

func TestFormatBudgetThresholdAlert_Warning(t *testing.T) {
	alert := FormatBudgetThresholdAlert("proj-1", "My Lab", 0.75, 600, 1000)

	assert.Equal(t, AlertSeverityWarning, alert.Severity)
	assert.Contains(t, alert.Title, "My Lab")
	assert.Contains(t, alert.Title, "60%")
	assert.Contains(t, alert.Body, "$600.00")
	assert.Contains(t, alert.Body, "$1000.00")
	assert.Equal(t, "proj-1", alert.ProjectID)
	assert.Equal(t, "threshold", alert.Tags["type"])
}

func TestFormatBudgetThresholdAlert_Critical(t *testing.T) {
	// 92% spent → critical
	alert := FormatBudgetThresholdAlert("proj-2", "Research Lab", 0.90, 920, 1000)

	assert.Equal(t, AlertSeverityCritical, alert.Severity)
	assert.Contains(t, alert.Title, "92%")
}

func TestFormatBudgetThresholdAlert_ZeroBudget(t *testing.T) {
	// totalBudget==0 should not panic (pct stays 0)
	alert := FormatBudgetThresholdAlert("proj-z", "Zero Lab", 0.50, 0, 0)

	assert.Equal(t, AlertSeverityWarning, alert.Severity) // pct==0, below 90
	assert.NotEmpty(t, alert.ID)
}

// ── FormatShortfallAlert ──────────────────────────────────────────────────

func TestFormatShortfallAlert(t *testing.T) {
	exhaustion := time.Now().Add(14 * 24 * time.Hour)
	alert := FormatShortfallAlert("proj-3", "Stats Group", 12.50, 14, exhaustion)

	assert.Equal(t, AlertSeverityWarning, alert.Severity)
	assert.Contains(t, alert.Title, "Stats Group")
	assert.Contains(t, alert.Title, "14 days")
	assert.Contains(t, alert.Body, "$12.50")
	assert.Equal(t, "shortfall", alert.Tags["type"])
	assert.Equal(t, "14", alert.Tags["days_until"])
}

func TestFormatShortfallAlert_ImmediateExhaustion(t *testing.T) {
	alert := FormatShortfallAlert("proj-4", "Urgent Lab", 100.0, 1, time.Now().Add(24*time.Hour))

	assert.Contains(t, alert.Title, "1 days")
}

// ── FormatCushionAlert ────────────────────────────────────────────────────

func TestFormatCushionAlert(t *testing.T) {
	alert := FormatCushionAlert("proj-5", "Cushion Lab", 0.10, 100, 1000, "hibernate")

	assert.Equal(t, AlertSeverityCritical, alert.Severity)
	assert.Contains(t, alert.Title, "Cushion Lab")
	assert.Contains(t, alert.Body, "hibernate")
	assert.Contains(t, alert.Body, "$100.00")
	assert.Equal(t, "cushion", alert.Tags["type"])
	assert.Equal(t, "hibernate", alert.Tags["mode"])
}

func TestFormatCushionAlert_StopMode(t *testing.T) {
	alert := FormatCushionAlert("proj-6", "Stop Lab", 0.05, 50, 1000, "stop")

	assert.Equal(t, "stop", alert.Tags["mode"])
	assert.Contains(t, alert.Body, "stop")
}

// ── FormatAutoActionAlert ─────────────────────────────────────────────────

func TestFormatAutoActionAlert_HibernateAll(t *testing.T) {
	alert := FormatAutoActionAlert("proj-7", "Action Lab", types.BudgetActionHibernateAll, 950, 1000)

	assert.Equal(t, AlertSeverityCritical, alert.Severity)
	assert.Contains(t, alert.Title, "Action Lab")
	assert.Contains(t, alert.Title, string(types.BudgetActionHibernateAll))
	assert.Equal(t, "auto_action", alert.Tags["type"])
	assert.Equal(t, string(types.BudgetActionHibernateAll), alert.Tags["action"])
}

func TestFormatAutoActionAlert_AllActionTypes(t *testing.T) {
	actions := []types.BudgetActionType{
		types.BudgetActionHibernateAll,
		types.BudgetActionStopAll,
		types.BudgetActionPreventLaunch,
		types.BudgetActionNotifyOnly,
	}
	for _, action := range actions {
		t.Run(string(action), func(t *testing.T) {
			alert := FormatAutoActionAlert("proj-x", "Lab", action, 100, 200)
			assert.Equal(t, string(action), alert.Tags["action"])
			assert.NotEmpty(t, alert.ID)
		})
	}
}

// ── LogDispatcher ─────────────────────────────────────────────────────────

func makeAlert(title string) Alert {
	return Alert{
		ID:       "test-" + title,
		Severity: AlertSeverityInfo,
		Title:    title,
		Body:     "test body",
	}
}

func TestLogDispatcher_Send(t *testing.T) {
	d := NewLogDispatcher()
	ctx := context.Background()

	err := d.Send(ctx, makeAlert("hello"))
	require.NoError(t, err)

	history := d.History(10)
	require.Len(t, history, 1)
	assert.Equal(t, "hello", history[0].Alert.Title)
	assert.True(t, history[0].Succeeded)
}

func TestLogDispatcher_SendBatch(t *testing.T) {
	d := NewLogDispatcher()
	ctx := context.Background()

	alerts := []Alert{makeAlert("a"), makeAlert("b"), makeAlert("c")}
	err := d.SendBatch(ctx, alerts)
	require.NoError(t, err)

	history := d.History(10)
	assert.Len(t, history, 3)
}

func TestLogDispatcher_History_Limit(t *testing.T) {
	d := NewLogDispatcher()
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		_ = d.Send(ctx, makeAlert("item"))
	}

	// Only ask for 3
	history := d.History(3)
	assert.Len(t, history, 3)
}

func TestLogDispatcher_History_ZeroLimit(t *testing.T) {
	d := NewLogDispatcher()
	ctx := context.Background()

	_ = d.Send(ctx, makeAlert("x"))
	_ = d.Send(ctx, makeAlert("y"))

	// limit=0 returns all
	history := d.History(0)
	assert.Len(t, history, 2)
}

// ── WebhookDispatcher ─────────────────────────────────────────────────────

func TestWebhookDispatcher_Send_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := NewWebhookDispatcher(WebhookConfig{
		Channels:   []Channel{{Type: ChannelWebhook, Target: srv.URL}},
		RetryCount: 0,
	})

	err := d.Send(context.Background(), makeAlert("webhook-ok"))
	require.NoError(t, err)

	history := d.History(5)
	require.Len(t, history, 1)
	assert.True(t, history[0].Succeeded)
}

func TestWebhookDispatcher_Send_Retry(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	d := NewWebhookDispatcher(WebhookConfig{
		Channels:   []Channel{{Type: ChannelWebhook, Target: srv.URL}},
		RetryCount: 2,
	})

	err := d.Send(context.Background(), makeAlert("retry-ok"))
	require.NoError(t, err) // 3rd attempt succeeds

	assert.EqualValues(t, 3, callCount.Load())
}

func TestWebhookDispatcher_Send_AllFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	d := NewWebhookDispatcher(WebhookConfig{
		Channels:   []Channel{{Type: ChannelWebhook, Target: srv.URL}},
		RetryCount: 2,
	})

	err := d.Send(context.Background(), makeAlert("all-fail"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")

	history := d.History(5)
	require.Len(t, history, 1)
	assert.False(t, history[0].Succeeded)
}

func TestWebhookDispatcher_Slack_Payload(t *testing.T) {
	var receivedBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		receivedBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := NewWebhookDispatcher(WebhookConfig{
		Channels:   []Channel{{Type: ChannelSlack, Target: srv.URL}},
		RetryCount: 0,
	})

	alert := Alert{
		ID:       "slack-test",
		Severity: AlertSeverityCritical,
		Title:    "Budget Critical",
		Body:     "You are over budget",
	}
	err := d.Send(context.Background(), alert)
	require.NoError(t, err)

	// Verify Slack JSON shape
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(receivedBody), &payload))
	text, ok := payload["text"].(string)
	require.True(t, ok, "payload must have 'text' field")
	assert.True(t, strings.Contains(text, "Budget Critical"), "text must contain title")
	assert.True(t, strings.Contains(text, "over budget"), "text must contain body")
}

func TestWebhookDispatcher_Email_NoError(t *testing.T) {
	// Email channel logs but does not return an error even without SMTP
	d := NewWebhookDispatcher(WebhookConfig{
		Channels:   []Channel{{Type: ChannelEmail, Target: "admin@example.com"}},
		RetryCount: 0,
	})

	err := d.Send(context.Background(), makeAlert("email-skip"))
	assert.NoError(t, err)
}
