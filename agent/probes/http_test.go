package probes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPProbeSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	probe := NewHTTPProbe("test-server", server.URL, "GET", 5*time.Second)
	metrics := probe.Collect(context.Background())

	if len(metrics) < 2 {
		t.Fatalf("Expected at least 2 metrics, got %d", len(metrics))
	}

	var foundUp, foundLatency, foundStatus bool
	for _, m := range metrics {
		switch m.Name {
		case "http_up":
			foundUp = true
			if m.Value != 1 {
				t.Errorf("Expected http_up=1, got %f", m.Value)
			}
		case "http_latency_ms":
			foundLatency = true
			if m.Value <= 0 {
				t.Errorf("Expected positive latency, got %f", m.Value)
			}
		case "http_status_code":
			foundStatus = true
			if m.Value != 200 {
				t.Errorf("Expected status 200, got %f", m.Value)
			}
		}
	}

	if !foundUp {
		t.Error("Missing http_up metric")
	}
	if !foundLatency {
		t.Error("Missing http_latency_ms metric")
	}
	if !foundStatus {
		t.Error("Missing http_status_code metric")
	}
}

func TestHTTPProbe4xxError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	probe := NewHTTPProbe("test-server", server.URL, "GET", 5*time.Second)
	metrics := probe.Collect(context.Background())

	var found4xx bool
	for _, m := range metrics {
		if m.Name == "http_errors_4xx" {
			found4xx = true
			if m.Value != 1 {
				t.Errorf("Expected http_errors_4xx=1, got %f", m.Value)
			}
		}
	}

	if !found4xx {
		t.Error("Expected http_errors_4xx metric for 404 response")
	}
}

func TestHTTPProbe5xxError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	probe := NewHTTPProbe("test-server", server.URL, "GET", 5*time.Second)
	metrics := probe.Collect(context.Background())

	var found5xx bool
	for _, m := range metrics {
		if m.Name == "http_errors_5xx" {
			found5xx = true
			if m.Value != 1 {
				t.Errorf("Expected http_errors_5xx=1, got %f", m.Value)
			}
		}
	}

	if !found5xx {
		t.Error("Expected http_errors_5xx metric for 500 response")
	}
}

func TestHTTPProbeTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	probe := NewHTTPProbe("test-server", server.URL, "GET", 100*time.Millisecond)
	metrics := probe.Collect(context.Background())

	var foundUp bool
	for _, m := range metrics {
		if m.Name == "http_up" {
			foundUp = true
			if m.Value != 0 {
				t.Errorf("Expected http_up=0 for timeout, got %f", m.Value)
			}
		}
	}

	if !foundUp {
		t.Error("Missing http_up metric")
	}
}

func TestHTTPProbeInvalidURL(t *testing.T) {
	probe := NewHTTPProbe("test-server", "http://invalid-host-that-does-not-exist.local", "GET", 1*time.Second)
	metrics := probe.Collect(context.Background())

	var foundUp bool
	for _, m := range metrics {
		if m.Name == "http_up" {
			foundUp = true
			if m.Value != 0 {
				t.Errorf("Expected http_up=0 for invalid host, got %f", m.Value)
			}
		}
	}

	if !foundUp {
		t.Error("Missing http_up metric")
	}
}
