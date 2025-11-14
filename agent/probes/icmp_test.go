package probes

import (
	"context"
	"testing"
	"time"
)

func TestICMPProbeLocalhost(t *testing.T) {
	probe := NewICMPProbe("test-icmp", "127.0.0.1", 2*time.Second)
	metrics := probe.Collect(context.Background())

	if len(metrics) != 2 {
		t.Fatalf("Expected 2 metrics, got %d", len(metrics))
	}

	var foundUp, foundRTT bool
	for _, m := range metrics {
		switch m.Name {
		case "icmp_up":
			foundUp = true
		case "icmp_rtt_ms":
			foundRTT = true
			if m.Value < 0 {
				t.Errorf("Expected non-negative RTT, got %f", m.Value)
			}
		}
	}

	if !foundUp {
		t.Error("Missing icmp_up metric")
	}
	if !foundRTT {
		t.Error("Missing icmp_rtt_ms metric")
	}
}

func TestICMPProbeInvalidHost(t *testing.T) {
	probe := NewICMPProbe("test-icmp", "192.0.2.1", 1*time.Second)
	metrics := probe.Collect(context.Background())

	var foundUp bool
	for _, m := range metrics {
		if m.Name == "icmp_up" {
			foundUp = true
			if m.Value != 0 {
				t.Errorf("Expected icmp_up=0 for unreachable host, got %f", m.Value)
			}
		}
	}

	if !foundUp {
		t.Error("Missing icmp_up metric")
	}
}

func TestICMPProbeLabels(t *testing.T) {
	host := "8.8.8.8"
	probe := NewICMPProbe("test-icmp", host, 2*time.Second)
	metrics := probe.Collect(context.Background())

	for _, m := range metrics {
		if m.Labels["host"] != host {
			t.Errorf("Expected host label %s, got %s", host, m.Labels["host"])
		}
	}
}

func TestICMPProbeService(t *testing.T) {
	probe := NewICMPProbe("test-icmp", "127.0.0.1", 2*time.Second)
	metrics := probe.Collect(context.Background())

	for _, m := range metrics {
		if m.Service != "network" {
			t.Errorf("Expected service 'network', got %s", m.Service)
		}
		if m.Target != "test-icmp" {
			t.Errorf("Expected target 'test-icmp', got %s", m.Target)
		}
	}
}

func TestICMPProbeTimeout(t *testing.T) {
	probe := NewICMPProbe("test-icmp", "192.0.2.254", 100*time.Millisecond)
	metrics := probe.Collect(context.Background())

	var foundUp bool
	for _, m := range metrics {
		if m.Name == "icmp_up" {
			foundUp = true
		}
	}

	if !foundUp {
		t.Error("Missing icmp_up metric")
	}
}
