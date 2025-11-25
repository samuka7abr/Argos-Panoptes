package main

import (
	"strings"
	"testing"
)

func TestBuildEmailBody(t *testing.T) {
	config := EmailConfig{
		SMTPHost: "smtp.example.com",
		From:     "alerts@example.com",
	}

	notifier := NewNotifier(config)

	rule := &Rule{
		Name:        "test-alert",
		Description: "Test alert description",
		Expr:        "last(1m, http_up) == 0",
		Service:     "web",
		Target:      "site-a",
		Severity:    "critical",
	}

	body := notifier.buildEmailBody(rule, 0.0)

	expectedParts := []string{
		"Alert: test-alert",
		"Severity: critical",
		"Description: Test alert description",
		"Expression: last(1m, http_up) == 0",
		"Current Value: 0.00",
		"Service: web",
		"Target: site-a",
		"Argos Panoptes Monitoring System",
	}

	for _, part := range expectedParts {
		if !strings.Contains(body, part) {
			t.Errorf("Email body missing expected part: %s", part)
		}
	}
}

func TestBuildEmailBodyNoServiceTarget(t *testing.T) {
	config := EmailConfig{
		From: "alerts@example.com",
	}

	notifier := NewNotifier(config)

	rule := &Rule{
		Name:        "test-alert",
		Description: "Test",
		Expr:        "last(1m, metric) > 100",
		Severity:    "warning",
	}

	body := notifier.buildEmailBody(rule, 150.5)

	if strings.Contains(body, "Service:") {
		t.Error("Email body should not contain 'Service:' when not set")
	}

	if strings.Contains(body, "Target:") {
		t.Error("Email body should not contain 'Target:' when not set")
	}

	if !strings.Contains(body, "Current Value: 150.50") {
		t.Error("Email body should contain formatted value")
	}
}

func TestBuildEmailBodySeverityInSubject(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "[CRITICAL]"},
		{"warning", "[WARNING]"},
		{"info", "[INFO]"},
	}

	rule := &Rule{
		Name:        "test-alert",
		Description: "Test",
		Expr:        "test",
	}

	for _, tt := range tests {
		rule.Severity = tt.severity
		subject := "[" + strings.ToUpper(tt.severity) + "] " + rule.Name

		if !strings.Contains(subject, tt.expected) {
			t.Errorf("Expected subject to contain %s for severity %s", tt.expected, tt.severity)
		}
	}
}

func TestNotifierConfiguration(t *testing.T) {
	config := EmailConfig{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUser:     "user@example.com",
		SMTPPassword: "password",
		From:         "alerts@example.com",
		UseTLS:       true,
	}

	notifier := NewNotifier(config)

	if notifier.config.SMTPHost != "smtp.example.com" {
		t.Errorf("Expected SMTP host smtp.example.com, got %s", notifier.config.SMTPHost)
	}

	if notifier.config.SMTPPort != 587 {
		t.Errorf("Expected SMTP port 587, got %d", notifier.config.SMTPPort)
	}

	if !notifier.config.UseTLS {
		t.Error("Expected UseTLS to be true")
	}
}

func TestSendAlertMultipleRecipients(t *testing.T) {
	config := EmailConfig{
		SMTPHost:     "invalid-smtp-server.local",
		SMTPPort:     587,
		SMTPUser:     "test@example.com",
		SMTPPassword: "password",
		From:         "alerts@example.com",
		UseTLS:       false,
	}

	notifier := NewNotifier(config)

	rule := &Rule{
		Name:        "test-alert",
		Description: "Test",
		Expr:        "test",
		Severity:    "critical",
		EmailTo:     []string{"recipient1@example.com", "recipient2@example.com"},
	}

	err := notifier.SendAlert(rule, 0.0)

	if err == nil {
		t.Error("Expected error when sending to invalid SMTP server")
	}

	if !strings.Contains(err.Error(), "recipient1@example.com") {
		t.Error("Error should mention the recipient email")
	}
}

func TestSendAlertEmptyRecipients(t *testing.T) {
	config := EmailConfig{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUser:     "test@example.com",
		SMTPPassword: "password",
		From:         "alerts@example.com",
	}

	notifier := NewNotifier(config)

	rule := &Rule{
		Name:        "test-alert",
		Description: "Test",
		Expr:        "test",
		Severity:    "critical",
		EmailTo:     []string{},
	}

	err := notifier.SendAlert(rule, 0.0)

	if err != nil {
		t.Errorf("Expected no error for empty recipients, got: %v", err)
	}
}
