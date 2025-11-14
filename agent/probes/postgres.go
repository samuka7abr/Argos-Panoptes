package probes

import (
	"context"
	"database/sql"
	"time"

	"argos/shared"

	_ "github.com/lib/pq"
)

type PostgresProbe struct {
	Name    string
	DSN     string
	SlowMS  int
	PingSQL string
}

func NewPostgresProbe(name, dsn string, slowMS int, pingSQL string) *PostgresProbe {
	return &PostgresProbe{
		Name:    name,
		DSN:     dsn,
		SlowMS:  slowMS,
		PingSQL: pingSQL,
	}
}

func (p *PostgresProbe) Collect(ctx context.Context) []shared.Metric {
	db, err := sql.Open("postgres", p.DSN)
	if err != nil {
		return p.errorMetrics()
	}
	defer db.Close()

	start := time.Now()
	var result int
	err = db.QueryRowContext(ctx, p.PingSQL).Scan(&result)
	latency := time.Since(start).Seconds() * 1000
	ts := time.Now()

	labels := map[string]string{}

	if err != nil {
		return p.errorMetrics()
	}

	metrics := []shared.Metric{
		{Service: "db", Target: p.Name, Name: "db_up", Value: 1, Labels: labels, TS: ts},
		{Service: "db", Target: p.Name, Name: "db_query_ms", Value: latency, Labels: labels, TS: ts},
	}

	var connections int
	err = db.QueryRowContext(ctx, "SELECT count(*) FROM pg_stat_activity").Scan(&connections)
	if err == nil {
		metrics = append(metrics, shared.Metric{
			Service: "db",
			Target:  p.Name,
			Name:    "db_connections",
			Value:   float64(connections),
			Labels:  labels,
			TS:      ts,
		})
	}

	var slowQueries int64
	err = db.QueryRowContext(ctx,
		"SELECT count(*) FROM pg_stat_activity WHERE state = 'active' AND query_start < NOW() - INTERVAL '1 second'",
	).Scan(&slowQueries)
	if err == nil {
		metrics = append(metrics, shared.Metric{
			Service: "db",
			Target:  p.Name,
			Name:    "db_slow_queries",
			Value:   float64(slowQueries),
			Labels:  labels,
			TS:      ts,
		})
	}

	return metrics
}

func (p *PostgresProbe) errorMetrics() []shared.Metric {
	ts := time.Now()
	labels := map[string]string{}

	return []shared.Metric{
		{Service: "db", Target: p.Name, Name: "db_up", Value: 0, Labels: labels, TS: ts},
	}
}
