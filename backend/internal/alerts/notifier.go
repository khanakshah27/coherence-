package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"coherence.dev/backend/internal/models"
)

// AlertLevel represents the urgency of an alert
type AlertLevel string

const (
	AlertInfo     AlertLevel = "info"
	AlertWarning  AlertLevel = "warning"
	AlertCritical AlertLevel = "critical"
)

// Alert represents a notification to be sent
type Alert struct {
	Level       AlertLevel
	Title       string
	Message     string
	DriftItem   *models.DriftItem
	ScanID      string
	Timestamp   time.Time
	ActionURL   string
}

// Notifier defines the interface for notification backends
type Notifier interface {
	Send(ctx context.Context, alert Alert) error
	Name() string
}

// Manager manages multiple notification backends
type Manager struct {
	notifiers []Notifier
	logger    *logrus.Logger
}

// NewManager creates a new alert manager
func NewManager(logger *logrus.Logger) *Manager {
	return &Manager{logger: logger}
}

// AddNotifier registers a notification backend
func (m *Manager) AddNotifier(n Notifier) {
	m.notifiers = append(m.notifiers, n)
}

// SendAlert dispatches an alert to all registered notifiers
func (m *Manager) SendAlert(ctx context.Context, alert Alert) {
	for _, n := range m.notifiers {
		go func(notifier Notifier) {
			if err := notifier.Send(ctx, alert); err != nil {
				m.logger.WithFields(logrus.Fields{
					"notifier": notifier.Name(),
					"error":    err,
				}).Error("Failed to send alert")
			}
		}(n)
	}
}

// ─── Slack ────────────────────────────────────────────────────────────────

// SlackNotifier sends alerts to Slack
type SlackNotifier struct {
	webhookURL string
	channel    string
}

// NewSlackNotifier creates a Slack notifier
func NewSlackNotifier(webhookURL, channel string) *SlackNotifier {
	return &SlackNotifier{webhookURL: webhookURL, channel: channel}
}

func (s *SlackNotifier) Name() string { return "slack" }

func (s *SlackNotifier) Send(ctx context.Context, alert Alert) error {
	color := "#36a64f" // green
	switch alert.Level {
	case AlertWarning:
		color = "#f59e0b"
	case AlertCritical:
		color = "#dc2626"
	}

	payload := map[string]interface{}{
		"channel": s.channel,
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"title": alert.Title,
				"text":  alert.Message,
				"fields": []map[string]string{
					{"title": "Severity", "value": string(alert.Level), "short": "true"},
					{"title": "Time", "value": alert.Timestamp.Format(time.RFC3339), "short": "true"},
				},
				"actions": []map[string]string{
					{"type": "button", "text": "View in Coherence", "url": alert.ActionURL},
				},
				"footer": "Coherence · Infrastructure Drift Detection",
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("slack: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: non-200 response: %d", resp.StatusCode)
	}
	return nil
}

// ─── PagerDuty ───────────────────────────────────────────────────────────

// PagerDutyNotifier sends alerts via PagerDuty Events API v2
type PagerDutyNotifier struct {
	integrationKey string
	apiURL         string
}

// NewPagerDutyNotifier creates a PagerDuty notifier
func NewPagerDutyNotifier(integrationKey string) *PagerDutyNotifier {
	return &PagerDutyNotifier{
		integrationKey: integrationKey,
		apiURL:         "https://events.pagerduty.com/v2/enqueue",
	}
}

func (p *PagerDutyNotifier) Name() string { return "pagerduty" }

func (p *PagerDutyNotifier) Send(ctx context.Context, alert Alert) error {
	severity := "info"
	switch alert.Level {
	case AlertWarning:
		severity = "warning"
	case AlertCritical:
		severity = "critical"
	}

	payload := map[string]interface{}{
		"routing_key":  p.integrationKey,
		"event_action": "trigger",
		"payload": map[string]interface{}{
			"summary":   alert.Title,
			"severity":  severity,
			"source":    "coherence",
			"timestamp": alert.Timestamp.Format(time.RFC3339),
			"custom_details": map[string]string{
				"message":  alert.Message,
				"scan_id":  alert.ScanID,
				"action_url": alert.ActionURL,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("pagerduty: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("pagerduty: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pagerduty: non-2xx response: %d", resp.StatusCode)
	}
	return nil
}

// ─── Helper constructors ─────────────────────────────────────────────────

// NewCriticalDriftAlert creates an alert for critical drift detection
func NewCriticalDriftAlert(drift *models.DriftItem, dashboardURL string) Alert {
	return Alert{
		Level:     AlertCritical,
		Title:     fmt.Sprintf("🚨 Critical Drift Detected: %s", drift.ResourceID),
		Message:   drift.Description,
		DriftItem: drift,
		Timestamp: time.Now(),
		ActionURL: fmt.Sprintf("%s/drifts/%s", dashboardURL, drift.ID),
	}
}

// NewScanCompletedAlert creates an alert for scan completion
func NewScanCompletedAlert(scan *models.Scan, dashboardURL string) Alert {
	level := AlertInfo
	if scan.DriftCount > 0 {
		level = AlertWarning
	}
	return Alert{
		Level:     level,
		Title:     fmt.Sprintf("Scan completed: %d drift(s) found", scan.DriftCount),
		Message:   fmt.Sprintf("Scan %s finished scanning %s resources.", scan.ID, scan.CloudProvider),
		ScanID:    scan.ID,
		Timestamp: time.Now(),
		ActionURL: fmt.Sprintf("%s/scans/%s", dashboardURL, scan.ID),
	}
}

// NewRemediationFailedAlert creates an alert for a failed remediation
func NewRemediationFailedAlert(req *models.RemediationRequest, drift *models.DriftItem, dashboardURL string) Alert {
	return Alert{
		Level:     AlertCritical,
		Title:     fmt.Sprintf("❌ Remediation Failed: %s", drift.ResourceID),
		Message:   fmt.Sprintf("Remediation action '%s' failed for resource %s.", req.ActionType, drift.ResourceID),
		DriftItem: drift,
		Timestamp: time.Now(),
		ActionURL: fmt.Sprintf("%s/remediations/%s", dashboardURL, req.ID),
	}
}
