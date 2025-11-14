package probes

import (
	"context"
	"net/http"
	"time"

	"argos/shared"
)

type HTTPProbe struct {
	Name    string
	URL     string
	Method  string
	Timeout time.Duration
	client  *http.Client
}

func NewHTTPProbe(name, url, method string, timeout time.Duration) *HTTPProbe {
	return &HTTPProbe{
		Name:    name,
		URL:     url,
		Method:  method,
		Timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

func (p *HTTPProbe) Collect(ctx context.Context) []shared.Metric {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, p.Method, p.URL, nil)
	if err != nil {
		return p.errorMetrics(start)
	}

	resp, err := p.client.Do(req)
	latency := time.Since(start).Seconds() * 1000
	ts := time.Now()

	labels := map[string]string{
		"url":    p.URL,
		"method": p.Method,
	}

	if err != nil {
		return []shared.Metric{
			{Service: "web", Target: p.Name, Name: "http_up", Value: 0, Labels: labels, TS: ts},
			{Service: "web", Target: p.Name, Name: "http_latency_ms", Value: latency, Labels: labels, TS: ts},
		}
	}
	defer resp.Body.Close()

	metrics := []shared.Metric{
		{Service: "web", Target: p.Name, Name: "http_up", Value: 1, Labels: labels, TS: ts},
		{Service: "web", Target: p.Name, Name: "http_latency_ms", Value: latency, Labels: labels, TS: ts},
		{Service: "web", Target: p.Name, Name: "http_status_code", Value: float64(resp.StatusCode), Labels: labels, TS: ts},
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		metrics = append(metrics, shared.Metric{
			Service: "web",
			Target:  p.Name,
			Name:    "http_errors_4xx",
			Value:   1,
			Labels:  labels,
			TS:      ts,
		})
	}

	if resp.StatusCode >= 500 {
		metrics = append(metrics, shared.Metric{
			Service: "web",
			Target:  p.Name,
			Name:    "http_errors_5xx",
			Value:   1,
			Labels:  labels,
			TS:      ts,
		})
	}

	return metrics
}

func (p *HTTPProbe) errorMetrics(start time.Time) []shared.Metric {
	latency := time.Since(start).Seconds() * 1000
	ts := time.Now()
	labels := map[string]string{
		"url":    p.URL,
		"method": p.Method,
	}

	return []shared.Metric{
		{Service: "web", Target: p.Name, Name: "http_up", Value: 0, Labels: labels, TS: ts},
		{Service: "web", Target: p.Name, Name: "http_latency_ms", Value: latency, Labels: labels, TS: ts},
	}
}
