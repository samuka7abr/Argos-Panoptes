package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Pusher struct {
	Endpoint string
	Client   *http.Client
}

func NewPusher(endpoint string) *Pusher {
	return &Pusher{
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *Pusher) Push(agentID string, metrics []Metric) error {
	batch := Batch{
		AgentID: agentID,
		Items:   metrics,
	}

	buf, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", p.Endpoint, bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("request creation error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ingest failed with status %d", resp.StatusCode)
	}

	return nil
}
