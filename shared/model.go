package shared

import "time"

type Metric struct {
	Service string            `json:"service"`
	Target  string            `json:"target"`
	Name    string            `json:"name"`
	Value   float64           `json:"value"`
	Labels  map[string]string `json:"labels"`
	TS      time.Time         `json:"ts"`
}

type Batch struct {
	AgentID string   `json:"agent_id"`
	Items   []Metric `json:"items"`
}

type QueryRequest struct {
	Name    string `json:"name"`
	Service string `json:"service,omitempty"`
	Target  string `json:"target,omitempty"`
}

type QueryRangeRequest struct {
	Name    string    `json:"name"`
	Service string    `json:"service,omitempty"`
	Target  string    `json:"target,omitempty"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	Step    string    `json:"step"`
}

type DataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type QueryRangeResponse struct {
	Service string      `json:"service"`
	Target  string      `json:"target"`
	Name    string      `json:"name"`
	Data    []DataPoint `json:"data"`
}

type Alert struct {
	Name     string            `json:"name"`
	Rule     string            `json:"rule"`
	Severity string            `json:"severity"`
	Service  string            `json:"service"`
	Target   string            `json:"target"`
	Since    time.Time         `json:"since"`
	Labels   map[string]string `json:"labels"`
	Message  string            `json:"message"`
}

type HealthResponse struct {
	Status       string    `json:"status"`
	Uptime       string    `json:"uptime"`
	MetricsCount int64     `json:"metrics_count"`
	LastIngest   time.Time `json:"last_ingest"`
	Version      string    `json:"version"`
}
