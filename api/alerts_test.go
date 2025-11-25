package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"argos/shared"
)

type mockStorageWithAlertRules struct {
	rules []AlertRule
}

func (m *mockStorageWithAlertRules) GetAlertRules() ([]AlertRule, error) {
	return m.rules, nil
}

func (m *mockStorageWithAlertRules) GetAlertRule(id int) (*AlertRule, error) {
	for _, r := range m.rules {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, nil
}

func (m *mockStorageWithAlertRules) CreateAlertRule(rule *AlertRule) error {
	rule.ID = len(m.rules) + 1
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	m.rules = append(m.rules, *rule)
	return nil
}

func (m *mockStorageWithAlertRules) UpdateAlertRule(rule *AlertRule) error {
	for i := range m.rules {
		if m.rules[i].ID == rule.ID {
			m.rules[i] = *rule
			return nil
		}
	}
	return nil
}

func (m *mockStorageWithAlertRules) DeleteAlertRule(id int) error {
	for i, r := range m.rules {
		if r.ID == id {
			m.rules = append(m.rules[:i], m.rules[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockStorageWithAlertRules) InsertMetrics(agentID string, metrics []shared.Metric) error {
	return nil
}
func (m *mockStorageWithAlertRules) QueryLatest(name, service, target string) (*shared.Metric, error) {
	return nil, nil
}
func (m *mockStorageWithAlertRules) QueryRange(name, service, target string, start, end time.Time, step string) ([]shared.DataPoint, error) {
	return nil, nil
}
func (m *mockStorageWithAlertRules) ListServices() ([]string, error)              { return nil, nil }
func (m *mockStorageWithAlertRules) ListTargets(service string) ([]string, error) { return nil, nil }
func (m *mockStorageWithAlertRules) GetMetricsCount() (int64, error)              { return 0, nil }
func (m *mockStorageWithAlertRules) GetLastIngestTime() (time.Time, error)        { return time.Time{}, nil }
func (m *mockStorageWithAlertRules) GetActiveAlerts() ([]shared.Alert, error)     { return nil, nil }
func (m *mockStorageWithAlertRules) Close() error                                 { return nil }

func TestListAlertRules(t *testing.T) {
	mock := &mockStorageWithAlertRules{
		rules: []AlertRule{
			{ID: 1, Name: "test-rule", Expr: "last(1m, http_up) == 0", Severity: "critical", EmailTo: []string{"test@example.com"}},
		},
	}
	storage = mock

	req := httptest.NewRequest("GET", "/api/alert-rules", nil)
	w := httptest.NewRecorder()

	alertsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	rules := response["rules"].([]interface{})
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
}

func TestCreateAlertRule(t *testing.T) {
	mock := &mockStorageWithAlertRules{rules: []AlertRule{}}
	storage = mock

	rule := AlertRule{
		Name:        "new-rule",
		Description: "Test rule",
		Expr:        "last(1m, http_up) == 0",
		Service:     "web",
		ForDuration: "1m",
		Severity:    "critical",
		EmailTo:     []string{"test@example.com"},
	}

	body, _ := json.Marshal(rule)
	req := httptest.NewRequest("POST", "/api/alert-rules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	alertsHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	if len(mock.rules) != 1 {
		t.Errorf("Expected 1 rule created, got %d", len(mock.rules))
	}
}

func TestCreateAlertRuleMissingName(t *testing.T) {
	mock := &mockStorageWithAlertRules{rules: []AlertRule{}}
	storage = mock

	rule := AlertRule{
		Expr:    "last(1m, http_up) == 0",
		EmailTo: []string{"test@example.com"},
	}

	body, _ := json.Marshal(rule)
	req := httptest.NewRequest("POST", "/api/alert-rules", bytes.NewReader(body))
	w := httptest.NewRecorder()

	alertsHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdateAlertRule(t *testing.T) {
	mock := &mockStorageWithAlertRules{
		rules: []AlertRule{
			{ID: 1, Name: "test-rule", Expr: "last(1m, http_up) == 0", EmailTo: []string{"old@example.com"}},
		},
	}
	storage = mock

	updated := AlertRule{
		Name:    "test-rule",
		Expr:    "last(2m, http_up) == 0",
		EmailTo: []string{"new@example.com"},
	}

	body, _ := json.Marshal(updated)
	req := httptest.NewRequest("PUT", "/api/alert-rules/1", bytes.NewReader(body))
	w := httptest.NewRecorder()

	alertsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if mock.rules[0].Expr != "last(2m, http_up) == 0" {
		t.Error("Rule was not updated")
	}
}

func TestDeleteAlertRule(t *testing.T) {
	mock := &mockStorageWithAlertRules{
		rules: []AlertRule{
			{ID: 1, Name: "test-rule", Expr: "last(1m, http_up) == 0", EmailTo: []string{"test@example.com"}},
		},
	}
	storage = mock

	req := httptest.NewRequest("DELETE", "/api/alert-rules/1", nil)
	w := httptest.NewRecorder()

	alertsHandler(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	if len(mock.rules) != 0 {
		t.Errorf("Expected rule to be deleted, got %d rules", len(mock.rules))
	}
}
