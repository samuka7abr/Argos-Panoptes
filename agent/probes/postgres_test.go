package probes

import (
	"context"
	"testing"
)

func TestPostgresProbeInvalidDSN(t *testing.T) {
	probe := NewPostgresProbe("test-db", "postgres://invalid:invalid@localhost:9999/nonexistent", 100, "SELECT 1")
	metrics := probe.Collect(context.Background())

	if len(metrics) == 0 {
		t.Fatal("Expected at least 1 metric")
	}

	var foundUp bool
	for _, m := range metrics {
		if m.Name == "db_up" {
			foundUp = true
			if m.Value != 0 {
				t.Errorf("Expected db_up=0 for invalid DSN, got %f", m.Value)
			}
		}
	}

	if !foundUp {
		t.Error("Missing db_up metric")
	}
}

func TestPostgresProbeService(t *testing.T) {
	probe := NewPostgresProbe("test-db", "postgres://user:pass@localhost:5432/db", 100, "SELECT 1")
	metrics := probe.Collect(context.Background())

	for _, m := range metrics {
		if m.Service != "db" {
			t.Errorf("Expected service 'db', got %s", m.Service)
		}
		if m.Target != "test-db" {
			t.Errorf("Expected target 'test-db', got %s", m.Target)
		}
	}
}

func TestPostgresProbeErrorMetrics(t *testing.T) {
	probe := NewPostgresProbe("test-db", "invalid-dsn", 100, "SELECT 1")
	metrics := probe.Collect(context.Background())

	if len(metrics) != 1 {
		t.Errorf("Expected exactly 1 metric on error, got %d", len(metrics))
	}

	if metrics[0].Name != "db_up" {
		t.Errorf("Expected db_up metric on error, got %s", metrics[0].Name)
	}
}
