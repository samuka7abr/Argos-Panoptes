package probes

import (
	"context"
	"testing"
	"time"
)

func TestSMTPProbeInvalidHost(t *testing.T) {
	probe := NewSMTPProbe("test-smtp", "invalid-smtp-host-12345.local", 25, false, 2*time.Second)
	metrics := probe.Collect(context.Background())

	if len(metrics) == 0 {
		t.Fatal("Expected at least 1 metric")
	}

	var foundUp bool
	for _, m := range metrics {
		if m.Name == "smtp_up" {
			foundUp = true
			if m.Value != 0 {
				t.Errorf("Expected smtp_up=0 for invalid host, got %f", m.Value)
			}
		}
	}

	if !foundUp {
		t.Error("Missing smtp_up metric")
	}
}

func TestSMTPProbeTimeout(t *testing.T) {
	probe := NewSMTPProbe("test-smtp", "192.0.2.1", 25, false, 100*time.Millisecond)
	metrics := probe.Collect(context.Background())

	var foundUp bool
	for _, m := range metrics {
		if m.Name == "smtp_up" {
			foundUp = true
			if m.Value != 0 {
				t.Errorf("Expected smtp_up=0 for timeout, got %f", m.Value)
			}
		}
	}

	if !foundUp {
		t.Error("Missing smtp_up metric")
	}
}

func TestSMTPProbeLabels(t *testing.T) {
	host := "smtp.example.com"
	port := 587
	probe := NewSMTPProbe("test-smtp", host, port, true, 5*time.Second)
	metrics := probe.Collect(context.Background())

	for _, m := range metrics {
		if m.Labels["host"] != host {
			t.Errorf("Expected host label %s, got %s", host, m.Labels["host"])
		}
		if m.Labels["port"] != "587" {
			t.Errorf("Expected port label 587, got %s", m.Labels["port"])
		}
	}
}

func TestSMTPProbeService(t *testing.T) {
	probe := NewSMTPProbe("test-smtp", "invalid.local", 25, false, 1*time.Second)
	metrics := probe.Collect(context.Background())

	for _, m := range metrics {
		if m.Service != "smtp" {
			t.Errorf("Expected service 'smtp', got %s", m.Service)
		}
		if m.Target != "test-smtp" {
			t.Errorf("Expected target 'test-smtp', got %s", m.Target)
		}
	}
}
