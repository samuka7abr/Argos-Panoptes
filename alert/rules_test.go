package main

import (
	"os"
	"testing"
	"time"
)

func TestLoadRules(t *testing.T) {
	yamlContent := `
rules:
  - name: test-rule
    description: "Test rule"
    expr: "last(1m, http_up) == 0"
    service: web
    for: 2m
    severity: critical
    email_to:
      - test@example.com

email:
  smtp_host: smtp.example.com
  smtp_port: 587
  smtp_user: user@example.com
  smtp_password: password
  from: alerts@example.com
  use_tls: true
`

	tmpFile, err := os.CreateTemp("", "rules-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := LoadRules(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load rules: %v", err)
	}

	if len(cfg.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(cfg.Rules))
	}

	rule := cfg.Rules[0]
	if rule.Name != "test-rule" {
		t.Errorf("Expected rule name 'test-rule', got %s", rule.Name)
	}

	if rule.Severity != "critical" {
		t.Errorf("Expected severity 'critical', got %s", rule.Severity)
	}

	if rule.For != "2m" {
		t.Errorf("Expected for '2m', got %s", rule.For)
	}

	if len(rule.EmailTo) != 1 {
		t.Errorf("Expected 1 email recipient, got %d", len(rule.EmailTo))
	}

	if cfg.Email.SMTPHost != "smtp.example.com" {
		t.Errorf("Expected SMTP host 'smtp.example.com', got %s", cfg.Email.SMTPHost)
	}

	if cfg.Email.SMTPPort != 587 {
		t.Errorf("Expected SMTP port 587, got %d", cfg.Email.SMTPPort)
	}
}

func TestLoadRulesDefaults(t *testing.T) {
	yamlContent := `
rules:
  - name: minimal-rule
    expr: "last(1m, http_up) == 0"

email:
  smtp_host: smtp.example.com
`

	tmpFile, err := os.CreateTemp("", "rules-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := LoadRules(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load rules: %v", err)
	}

	rule := cfg.Rules[0]

	if rule.For != "1m" {
		t.Errorf("Expected default for '1m', got %s", rule.For)
	}

	if rule.Severity != "warning" {
		t.Errorf("Expected default severity 'warning', got %s", rule.Severity)
	}

	if cfg.Email.SMTPPort != 587 {
		t.Errorf("Expected default SMTP port 587, got %d", cfg.Email.SMTPPort)
	}
}

func TestRuleForDuration(t *testing.T) {
	tests := []struct {
		forStr   string
		expected time.Duration
	}{
		{"1m", 1 * time.Minute},
		{"5m", 5 * time.Minute},
		{"1h", 1 * time.Hour},
		{"30s", 30 * time.Second},
		{"invalid", 1 * time.Minute},
	}

	for _, tt := range tests {
		rule := Rule{For: tt.forStr}
		result := rule.ForDuration()
		if result != tt.expected {
			t.Errorf("ForDuration(%s) = %v, want %v", tt.forStr, result, tt.expected)
		}
	}
}

func TestLoadRulesFileNotFound(t *testing.T) {
	_, err := LoadRules("nonexistent-file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoadRulesInvalidYAML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rules-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte("invalid: yaml: content: [")); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	_, err = LoadRules(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}
