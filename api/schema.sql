CREATE TABLE IF NOT EXISTS metrics (
    ts TIMESTAMPTZ NOT NULL,
    service TEXT NOT NULL,
    target TEXT NOT NULL,
    name TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    labels JSONB NOT NULL DEFAULT '{}'::jsonb,
    agent_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_metrics_ts ON metrics (ts DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_by_name ON metrics (name, service, target, ts DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_by_service ON metrics (service, ts DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_labels ON metrics USING GIN (labels);

CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    rule TEXT NOT NULL,
    severity TEXT NOT NULL,
    service TEXT NOT NULL,
    target TEXT NOT NULL,
    labels JSONB NOT NULL DEFAULT '{}'::jsonb,
    message TEXT,
    fired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    notified BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_alerts_active ON alerts (resolved_at) WHERE resolved_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts (severity, fired_at DESC);

CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    alert_id INTEGER REFERENCES alerts(id),
    channel TEXT NOT NULL,
    recipient TEXT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    success BOOLEAN NOT NULL,
    error TEXT
);

CREATE OR REPLACE VIEW latest_metrics AS
SELECT DISTINCT ON (service, target, name)
    ts,
    service,
    target,
    name,
    value,
    labels,
    agent_id
FROM metrics
ORDER BY service, target, name, ts DESC;

CREATE OR REPLACE FUNCTION cleanup_old_metrics()
RETURNS void AS $$
BEGIN
    DELETE FROM metrics WHERE ts < NOW() - INTERVAL '30 days';
    DELETE FROM alerts WHERE resolved_at < NOW() - INTERVAL '90 days';
    DELETE FROM notifications WHERE sent_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;
