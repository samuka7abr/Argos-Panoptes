package main

import (
	"argos/shared"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockStorage struct {
	metrics []shared.Metric
}

func (m *mockStorage) InsertMetrics(agentID string, metrics []shared.Metric) error {
	m.metrics = append(m.metrics, metrics...)
	return nil
}

func (m *mockStorage) QueryLatest(name, service, target string) (*shared.Metric, error) {
	for i := len(m.metrics) - 1; i >= 0; i-- {
		metric := m.metrics[i]
		if metric.Name == name {
			if service != "" && metric.Service != service {
				continue
			}
			if target != "" && metric.Target != target {
				continue
			}
			return &metric, nil
		}
	}
	return nil, nil
}

func (m *mockStorage) QueryRange(name, service, target string, start, end time.Time, step string) ([]shared.DataPoint, error) {
	return []shared.DataPoint{
		{Timestamp: time.Now().Unix(), Value: 45.2},
	}, nil
}

func (m *mockStorage) ListServices() ([]string, error) {
	return []string{"web", "db"}, nil
}

func (m *mockStorage) ListTargets(service string) ([]string, error) {
	return []string{"site-a", "site-b"}, nil
}

func (m *mockStorage) GetMetricsCount() (int64, error) {
	return int64(len(m.metrics)), nil
}

func (m *mockStorage) GetLastIngestTime() (time.Time, error) {
	if len(m.metrics) == 0 {
		return time.Time{}, nil
	}
	return m.metrics[len(m.metrics)-1].TS, nil
}

func (m *mockStorage) GetActiveAlerts() ([]shared.Alert, error) {
	return []shared.Alert{}, nil
}

func (m *mockStorage) GetAlertRules() ([]AlertRule, error) {
	return nil, nil
}

func (m *mockStorage) GetAlertRule(id int) (*AlertRule, error) {
	return nil, nil
}

func (m *mockStorage) CreateAlertRule(rule *AlertRule) error {
	return nil
}

func (m *mockStorage) UpdateAlertRule(rule *AlertRule) error {
	return nil
}

func (m *mockStorage) DeleteAlertRule(id int) error {
	return nil
}

func (m *mockStorage) GetLatestMetrics() ([]shared.Metric, error) {
	return m.metrics, nil
}

// Security methods
func (m *mockStorage) GetSecurityEvents(limit int) ([]SecurityEvent, error) {
	return nil, nil
}

func (m *mockStorage) CreateSecurityEvent(event *SecurityEvent) error {
	return nil
}

func (m *mockStorage) GetFailedLoginsByIP(limit int) ([]struct {
	IPAddress string `json:"ip_address"`
	Count     int    `json:"count"`
}, error) {
	return nil, nil
}

func (m *mockStorage) GetTotalFailedLogins() (int64, error) {
	return 0, nil
}

func (m *mockStorage) RecordFailedLogin(ip, username, service, userAgent string) error {
	return nil
}

func (m *mockStorage) GetConfigChanges(limit int) ([]ConfigChange, error) {
	return nil, nil
}

func (m *mockStorage) RecordConfigChange(change *ConfigChange) error {
	return nil
}

func (m *mockStorage) GetVulnerabilities() ([]Vulnerability, error) {
	return nil, nil
}

func (m *mockStorage) GetTrafficAnomalies(limit int) (int64, error) {
	return 0, nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestIngestHandler(t *testing.T) {
	mock := &mockStorage{}
	storage = mock
	startTime = time.Now()

	batch := shared.Batch{
		AgentID: "agent-test",
		Items: []shared.Metric{
			{
				Service: "web",
				Target:  "site",
				Name:    "http_latency_ms",
				Value:   45.2,
				TS:      time.Now(),
			},
		},
	}

	body, _ := json.Marshal(batch)
	req := httptest.NewRequest("POST", "/ingest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ingestHandler(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	if len(mock.metrics) != 1 {
		t.Errorf("Expected 1 metric stored, got %d", len(mock.metrics))
	}
}

func TestIngestHandlerInvalidMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/ingest", nil)
	w := httptest.NewRecorder()

	ingestHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestIngestHandlerEmptyBatch(t *testing.T) {
	mock := &mockStorage{}
	storage = mock

	batch := shared.Batch{
		AgentID: "agent-test",
		Items:   []shared.Metric{},
	}

	body, _ := json.Marshal(batch)
	req := httptest.NewRequest("POST", "/ingest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ingestHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHealthHandler(t *testing.T) {
	mock := &mockStorage{
		metrics: []shared.Metric{
			{TS: time.Now()},
		},
	}
	storage = mock
	startTime = time.Now().Add(-5 * time.Minute)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var health shared.HealthResponse
	json.NewDecoder(w.Body).Decode(&health)

	if health.Status == "" {
		t.Error("Expected health status, got empty")
	}

	if health.MetricsCount != 1 {
		t.Errorf("Expected metrics count 1, got %d", health.MetricsCount)
	}
}

func TestQueryHandler(t *testing.T) {
	mock := &mockStorage{
		metrics: []shared.Metric{
			{
				Service: "web",
				Target:  "site",
				Name:    "http_latency_ms",
				Value:   45.2,
				TS:      time.Now(),
			},
		},
	}
	storage = mock

	req := httptest.NewRequest("GET", "/api/metrics/query?name=http_latency_ms&service=web", nil)
	w := httptest.NewRecorder()

	queryHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var metric shared.Metric
	json.NewDecoder(w.Body).Decode(&metric)

	if metric.Value != 45.2 {
		t.Errorf("Expected value 45.2, got %f", metric.Value)
	}
}

func TestQueryHandlerMissingName(t *testing.T) {
	mock := &mockStorage{}
	storage = mock

	req := httptest.NewRequest("GET", "/api/metrics/query", nil)
	w := httptest.NewRecorder()

	queryHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestQueryHandlerNotFound(t *testing.T) {
	mock := &mockStorage{}
	storage = mock

	req := httptest.NewRequest("GET", "/api/metrics/query?name=nonexistent", nil)
	w := httptest.NewRecorder()

	queryHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestQueryRangeHandler(t *testing.T) {
	mock := &mockStorage{}
	storage = mock

	req := httptest.NewRequest("GET", "/api/metrics/range?name=http_latency_ms&start=-1h", nil)
	w := httptest.NewRecorder()

	queryRangeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response shared.QueryRangeResponse
	json.NewDecoder(w.Body).Decode(&response)

	if len(response.Data) == 0 {
		t.Error("Expected data points, got empty")
	}
}

func TestListServicesHandler(t *testing.T) {
	mock := &mockStorage{}
	storage = mock

	req := httptest.NewRequest("GET", "/api/services", nil)
	w := httptest.NewRecorder()

	listServicesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	services := response["services"].([]interface{})
	if len(services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(services))
	}
}

func TestListTargetsHandler(t *testing.T) {
	mock := &mockStorage{}
	storage = mock

	req := httptest.NewRequest("GET", "/api/targets?service=web", nil)
	w := httptest.NewRecorder()

	listTargetsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	targets := response["targets"].([]interface{})
	if len(targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(targets))
	}
}

func TestActiveAlertsHandler(t *testing.T) {
	mock := &mockStorage{}
	storage = mock

	req := httptest.NewRequest("GET", "/api/alerts/active", nil)
	w := httptest.NewRecorder()

	activeAlertsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	count := int(response["count"].(float64))
	if count != 0 {
		t.Errorf("Expected 0 alerts, got %d", count)
	}
}
