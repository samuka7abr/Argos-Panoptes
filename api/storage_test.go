package main

import (
	"argos/shared"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *Storage {
	dsn := "postgres://postgres:postgres@localhost:5432/argos_test?sslmode=disable"
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skip("PostgreSQL not available for testing")
	}

	if err := db.Ping(); err != nil {
		t.Skip("PostgreSQL not available for testing")
	}

	_, err = db.Exec(`
		DROP TABLE IF EXISTS notifications CASCADE;
		DROP TABLE IF EXISTS alerts CASCADE;
		DROP TABLE IF EXISTS metrics CASCADE;
	`)
	if err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	schemaSQL := `
		CREATE TABLE metrics (
			ts TIMESTAMPTZ NOT NULL,
			service TEXT NOT NULL,
			target TEXT NOT NULL,
			name TEXT NOT NULL,
			value DOUBLE PRECISION NOT NULL,
			labels JSONB NOT NULL DEFAULT '{}'::jsonb,
			agent_id TEXT
		);
		CREATE INDEX idx_metrics_ts ON metrics (ts DESC);
		CREATE INDEX idx_metrics_by_name ON metrics (name, service, target, ts DESC);

		CREATE TABLE alerts (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			rule TEXT NOT NULL,
			severity TEXT NOT NULL,
			service TEXT NOT NULL,
			target TEXT NOT NULL,
			labels JSONB NOT NULL DEFAULT '{}'::jsonb,
			message TEXT,
			fired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			resolved_at TIMESTAMPTZ,
			notified BOOLEAN DEFAULT FALSE
		);
	`

	if _, err := db.Exec(schemaSQL); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	storage := &Storage{db: db}
	t.Cleanup(func() {
		storage.Close()
	})

	return storage
}

func TestInsertMetrics(t *testing.T) {
	storage := setupTestDB(t)

	metrics := []shared.Metric{
		{
			Service: "web",
			Target:  "site",
			Name:    "http_latency_ms",
			Value:   45.2,
			Labels:  map[string]string{"url": "https://exemplo.com"},
			TS:      time.Now(),
		},
		{
			Service: "web",
			Target:  "site",
			Name:    "http_up",
			Value:   1,
			Labels:  map[string]string{"url": "https://exemplo.com"},
			TS:      time.Now(),
		},
	}

	err := storage.InsertMetrics("agent-01", metrics)
	if err != nil {
		t.Fatalf("InsertMetrics failed: %v", err)
	}

	var count int
	err = storage.db.QueryRow("SELECT COUNT(*) FROM metrics").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count metrics: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 metrics, got %d", count)
	}
}

func TestQueryLatest(t *testing.T) {
	storage := setupTestDB(t)

	now := time.Now()
	metrics := []shared.Metric{
		{Service: "web", Target: "site", Name: "http_latency_ms", Value: 30.0, TS: now.Add(-2 * time.Minute)},
		{Service: "web", Target: "site", Name: "http_latency_ms", Value: 45.2, TS: now.Add(-1 * time.Minute)},
		{Service: "web", Target: "site", Name: "http_latency_ms", Value: 50.0, TS: now},
	}

	storage.InsertMetrics("agent-01", metrics)

	latest, err := storage.QueryLatest("http_latency_ms", "web", "site")
	if err != nil {
		t.Fatalf("QueryLatest failed: %v", err)
	}

	if latest == nil {
		t.Fatal("Expected metric, got nil")
	}

	if latest.Value != 50.0 {
		t.Errorf("Expected value 50.0, got %f", latest.Value)
	}
}

func TestQueryLatestNotFound(t *testing.T) {
	storage := setupTestDB(t)

	latest, err := storage.QueryLatest("nonexistent", "web", "site")
	if err != nil {
		t.Fatalf("QueryLatest failed: %v", err)
	}

	if latest != nil {
		t.Error("Expected nil for nonexistent metric")
	}
}

func TestQueryRange(t *testing.T) {
	storage := setupTestDB(t)

	now := time.Now()
	metrics := []shared.Metric{
		{Service: "web", Target: "site", Name: "http_latency_ms", Value: 30.0, TS: now.Add(-10 * time.Minute)},
		{Service: "web", Target: "site", Name: "http_latency_ms", Value: 40.0, TS: now.Add(-8 * time.Minute)},
		{Service: "web", Target: "site", Name: "http_latency_ms", Value: 50.0, TS: now.Add(-5 * time.Minute)},
		{Service: "web", Target: "site", Name: "http_latency_ms", Value: 60.0, TS: now.Add(-2 * time.Minute)},
	}

	storage.InsertMetrics("agent-01", metrics)

	start := now.Add(-15 * time.Minute)
	end := now

	dataPoints, err := storage.QueryRange("http_latency_ms", "web", "site", start, end, "1m")
	if err != nil {
		t.Fatalf("QueryRange failed: %v", err)
	}

	if len(dataPoints) == 0 {
		t.Error("Expected data points, got empty slice")
	}
}

func TestListServices(t *testing.T) {
	storage := setupTestDB(t)

	metrics := []shared.Metric{
		{Service: "web", Target: "site", Name: "http_up", Value: 1, TS: time.Now()},
		{Service: "db", Target: "postgres", Name: "db_up", Value: 1, TS: time.Now()},
		{Service: "dns", Target: "resolver", Name: "dns_up", Value: 1, TS: time.Now()},
	}

	storage.InsertMetrics("agent-01", metrics)

	services, err := storage.ListServices()
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	expected := []string{"db", "dns", "web"}
	if len(services) != len(expected) {
		t.Errorf("Expected %d services, got %d", len(expected), len(services))
	}

	for i, svc := range services {
		if svc != expected[i] {
			t.Errorf("Expected service %s at index %d, got %s", expected[i], i, svc)
		}
	}
}

func TestListTargets(t *testing.T) {
	storage := setupTestDB(t)

	metrics := []shared.Metric{
		{Service: "web", Target: "site-a", Name: "http_up", Value: 1, TS: time.Now()},
		{Service: "web", Target: "site-b", Name: "http_up", Value: 1, TS: time.Now()},
		{Service: "db", Target: "postgres", Name: "db_up", Value: 1, TS: time.Now()},
	}

	storage.InsertMetrics("agent-01", metrics)

	targets, err := storage.ListTargets("web")
	if err != nil {
		t.Fatalf("ListTargets failed: %v", err)
	}

	expected := []string{"site-a", "site-b"}
	if len(targets) != len(expected) {
		t.Errorf("Expected %d targets, got %d", len(expected), len(targets))
	}
}

func TestGetMetricsCount(t *testing.T) {
	storage := setupTestDB(t)

	metrics := []shared.Metric{
		{Service: "web", Target: "site", Name: "http_up", Value: 1, TS: time.Now()},
		{Service: "web", Target: "site", Name: "http_latency_ms", Value: 45, TS: time.Now()},
	}

	storage.InsertMetrics("agent-01", metrics)

	count, err := storage.GetMetricsCount()
	if err != nil {
		t.Fatalf("GetMetricsCount failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

