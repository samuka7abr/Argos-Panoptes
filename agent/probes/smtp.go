package probes

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"argos/shared"
)

type SMTPProbe struct {
	Name     string
	Host     string
	Port     int
	StartTLS bool
	Timeout  time.Duration
}

func NewSMTPProbe(name, host string, port int, startTLS bool, timeout time.Duration) *SMTPProbe {
	return &SMTPProbe{
		Name:     name,
		Host:     host,
		Port:     port,
		StartTLS: startTLS,
		Timeout:  timeout,
	}
}

func (p *SMTPProbe) Collect(ctx context.Context) []shared.Metric {
	start := time.Now()
	addr := net.JoinHostPort(p.Host, fmt.Sprint(p.Port))

	dialer := net.Dialer{Timeout: p.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)

	ts := time.Now()
	labels := map[string]string{
		"host": p.Host,
		"port": fmt.Sprint(p.Port),
	}

	if err != nil {
		return []shared.Metric{
			{Service: "smtp", Target: p.Name, Name: "smtp_up", Value: 0, Labels: labels, TS: ts},
		}
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, p.Host)
	if err != nil {
		return []shared.Metric{
			{Service: "smtp", Target: p.Name, Name: "smtp_up", Value: 0, Labels: labels, TS: ts},
		}
	}
	defer client.Quit()

	if p.StartTLS {
		tlsConfig := &tls.Config{
			ServerName: p.Host,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return []shared.Metric{
				{Service: "smtp", Target: p.Name, Name: "smtp_up", Value: 0, Labels: labels, TS: ts},
			}
		}
	}

	latency := time.Since(start).Seconds() * 1000

	return []shared.Metric{
		{Service: "smtp", Target: p.Name, Name: "smtp_up", Value: 1, Labels: labels, TS: ts},
		{Service: "smtp", Target: p.Name, Name: "smtp_handshake_ms", Value: latency, Labels: labels, TS: ts},
	}
}
