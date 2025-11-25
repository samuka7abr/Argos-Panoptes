package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadRulesFromAPI(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/alert-rules" {
			t.Errorf("Expected path /api/alert-rules, got %s", r.URL.Path)
		}

		response := APIResponse{
			Rules: []APIAlertRule{
				{
					ID:          1,
					Name:        "test-rule",
					Description: "Test rule",
					Expr:        "last(1m, http_up) == 0",
					Service:     "web",
					Target:      "example.com",
					ForDuration: "2m",
					Severity:    "critical",
					EmailTo:     []string{"test@example.com"},
					Enabled:     true,
				},
			},
			Count: 1,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockAPI.Close()

	cfg, err := LoadRulesFromAPI(mockAPI.URL)
	if err != nil {
		t.Fatalf("LoadRulesFromAPI failed: %v", err)
	}

	if len(cfg.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(cfg.Rules))
	}

	rule := cfg.Rules[0]
	if rule.Name != "test-rule" {
		t.Errorf("Expected name 'test-rule', got '%s'", rule.Name)
	}

	if rule.Expr != "last(1m, http_up) == 0" {
		t.Errorf("Expected expr 'last(1m, http_up) == 0', got '%s'", rule.Expr)
	}

	if rule.Severity != "critical" {
		t.Errorf("Expected severity 'critical', got '%s'", rule.Severity)
	}

	if len(rule.EmailTo) != 1 || rule.EmailTo[0] != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %v", rule.EmailTo)
	}
}

func TestLoadRulesFromAPIEmpty(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := APIResponse{
			Rules: []APIAlertRule{},
			Count: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockAPI.Close()

	_, err := LoadRulesFromAPI(mockAPI.URL)
	if err == nil {
		t.Error("Expected error for empty rules, got nil")
	}
}

func TestLoadRulesFromAPIServerError(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockAPI.Close()

	_, err := LoadRulesFromAPI(mockAPI.URL)
	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestLoadRulesHybrid(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := APIResponse{
			Rules: []APIAlertRule{
				{
					ID:          1,
					Name:        "api-rule",
					Expr:        "last(1m, http_up) == 0",
					ForDuration: "1m",
					Severity:    "warning",
					EmailTo:     []string{"api@example.com"},
				},
			},
			Count: 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockAPI.Close()

	cfg, err := LoadRulesHybrid(mockAPI.URL, "rules.example.yaml")
	if err != nil {
		t.Fatalf("LoadRulesHybrid failed: %v", err)
	}

	if len(cfg.Rules) == 0 {
		t.Fatal("Expected at least 1 rule")
	}

	if cfg.Email.From == "" {
		t.Error("Expected SMTP config to be loaded from YAML")
	}
}

func TestLoadRulesHybridAPIDown(t *testing.T) {
	cfg, err := LoadRulesHybrid("http://localhost:99999", "rules.example.yaml")
	if err != nil {
		t.Fatalf("LoadRulesHybrid should fallback to YAML: %v", err)
	}

	if len(cfg.Rules) == 0 {
		t.Fatal("Expected rules from YAML fallback")
	}
}
