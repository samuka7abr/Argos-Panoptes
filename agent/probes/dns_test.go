package probes

import (
	"context"
	"testing"
)

func TestDNSProbeSuccess(t *testing.T) {
	probe := NewDNSProbe("test-dns", "google.com", "8.8.8.8:53")
	metrics := probe.Collect(context.Background())

	if len(metrics) != 2 {
		t.Fatalf("Expected 2 metrics, got %d", len(metrics))
	}

	var foundUp, foundLatency bool
	for _, m := range metrics {
		switch m.Name {
		case "dns_up":
			foundUp = true
			if m.Value != 1 {
				t.Errorf("Expected dns_up=1, got %f", m.Value)
			}
		case "dns_lookup_ms":
			foundLatency = true
			if m.Value <= 0 {
				t.Errorf("Expected positive latency, got %f", m.Value)
			}
		}
	}

	if !foundUp {
		t.Error("Missing dns_up metric")
	}
	if !foundLatency {
		t.Error("Missing dns_lookup_ms metric")
	}
}

func TestDNSProbeInvalidFQDN(t *testing.T) {
	probe := NewDNSProbe("test-dns", "this-domain-absolutely-does-not-exist-12345.com", "8.8.8.8:53")
	metrics := probe.Collect(context.Background())

	var foundUp bool
	for _, m := range metrics {
		if m.Name == "dns_up" {
			foundUp = true
			if m.Value != 0 {
				t.Errorf("Expected dns_up=0 for invalid FQDN, got %f", m.Value)
			}
		}
	}

	if !foundUp {
		t.Error("Missing dns_up metric")
	}
}

func TestDNSProbeInvalidServer(t *testing.T) {
	probe := NewDNSProbe("test-dns", "google.com", "192.0.2.1:53")
	metrics := probe.Collect(context.Background())

	var foundUp bool
	for _, m := range metrics {
		if m.Name == "dns_up" {
			foundUp = true
		}
	}

	if !foundUp {
		t.Error("Missing dns_up metric")
	}
}

func TestDNSProbeLabels(t *testing.T) {
	fqdn := "example.com"
	server := "1.1.1.1:53"
	probe := NewDNSProbe("test-dns", fqdn, server)
	metrics := probe.Collect(context.Background())

	for _, m := range metrics {
		if m.Labels["fqdn"] != fqdn {
			t.Errorf("Expected FQDN label %s, got %s", fqdn, m.Labels["fqdn"])
		}
		if m.Labels["server"] != server {
			t.Errorf("Expected server label %s, got %s", server, m.Labels["server"])
		}
	}
}
