package probes

import (
	"context"
	"net"
	"time"

	"argos/shared"
)

type ICMPProbe struct {
	Name    string
	Host    string
	Timeout time.Duration
}

func NewICMPProbe(name, host string, timeout time.Duration) *ICMPProbe {
	return &ICMPProbe{
		Name:    name,
		Host:    host,
		Timeout: timeout,
	}
}

func (p *ICMPProbe) Collect(ctx context.Context) []shared.Metric {
	start := time.Now()

	dialer := net.Dialer{Timeout: p.Timeout}
	conn, err := dialer.DialContext(ctx, "ip4:icmp", p.Host)

	if err != nil {
		conn, err = dialer.DialContext(ctx, "tcp", net.JoinHostPort(p.Host, "80"))
	}

	latency := time.Since(start).Seconds() * 1000
	ts := time.Now()

	labels := map[string]string{
		"host": p.Host,
	}

	up := 1.0
	if err != nil {
		up = 0.0
	} else {
		conn.Close()
	}

	return []shared.Metric{
		{Service: "network", Target: p.Name, Name: "icmp_up", Value: up, Labels: labels, TS: ts},
		{Service: "network", Target: p.Name, Name: "icmp_rtt_ms", Value: latency, Labels: labels, TS: ts},
	}
}
