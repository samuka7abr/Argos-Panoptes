package probes

import (
	"context"
	"net"
	"time"

	"argos/shared"
)

type DNSProbe struct {
	Name   string
	FQDN   string
	Server string
}

func NewDNSProbe(name, fqdn, server string) *DNSProbe {
	return &DNSProbe{
		Name:   name,
		FQDN:   fqdn,
		Server: server,
	}
}

func (p *DNSProbe) Collect(ctx context.Context) []shared.Metric {
	start := time.Now()

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			dialer := net.Dialer{Timeout: 5 * time.Second}
			return dialer.DialContext(ctx, "udp", p.Server)
		},
	}

	_, err := resolver.LookupHost(ctx, p.FQDN)
	latency := time.Since(start).Seconds() * 1000
	ts := time.Now()

	labels := map[string]string{
		"fqdn":   p.FQDN,
		"server": p.Server,
	}

	up := 1.0
	if err != nil {
		up = 0.0
	}

	return []shared.Metric{
		{Service: "dns", Target: p.Name, Name: "dns_up", Value: up, Labels: labels, TS: ts},
		{Service: "dns", Target: p.Name, Name: "dns_lookup_ms", Value: latency, Labels: labels, TS: ts},
	}
}
