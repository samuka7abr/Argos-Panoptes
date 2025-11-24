package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"argos/shared"
)

func TestCompareValues(t *testing.T) {
	tests := []struct {
		value     float64
		operator  string
		threshold float64
		expected  bool
	}{
		{100, ">", 50, true},
		{100, ">", 100, false},
		{100, ">", 150, false},
		{100, ">=", 100, true},
		{100, ">=", 50, true},
		{100, "<", 150, true},
		{100, "<", 100, false},
		{100, "<=", 100, true},
		{100, "==", 100, true},
		{100, "==", 101, false},
		{100, "!=", 101, true},
		{100, "!=", 100, false},
	}

	for _, tt := range tests {
		result := compareValues(tt.value, tt.operator, tt.threshold)
		if result != tt.expected {
			t.Errorf("compareValues(%.0f, %s, %.0f) = %v, want %v",
				tt.value, tt.operator, tt.threshold, result, tt.expected)
		}
	}
}

func TestCompareValuesInvalidOperator(t *testing.T) {
	result := compareValues(100, "invalid", 50)
	if result != false {
		t.Error("Expected false for invalid operator")
	}
}

func TestEvaluateLastSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/metrics/query" {
			t.Errorf("Expected path /api/metrics/query, got %s", r.URL.Path)
		}

		name := r.URL.Query().Get("name")
		if name != "http_up" {
			t.Errorf("Expected name=http_up, got %s", name)
		}

		metric := shared.Metric{
			Service: "web",
			Target:  "test",
			Name:    "http_up",
			Value:   0,
			TS:      time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metric)
	}))
	defer server.Close()

	evaluator := NewEvaluator(server.URL)
	rule := &Rule{
		Name:     "test-rule",
		Expr:     "last(1m, http_up) == 0",
		Service:  "web",
		Target:   "test",
		Severity: "critical",
	}

	triggered, value, err := evaluator.Evaluate(context.Background(), rule)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if !triggered {
		t.Error("Expected rule to trigger")
	}

	if value != 0 {
		t.Errorf("Expected value 0, got %f", value)
	}
}

func TestEvaluateLastNotTriggered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metric := shared.Metric{
			Service: "web",
			Target:  "test",
			Name:    "http_up",
			Value:   1,
			TS:      time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metric)
	}))
	defer server.Close()

	evaluator := NewEvaluator(server.URL)
	rule := &Rule{
		Name:    "test-rule",
		Expr:    "last(1m, http_up) == 0",
		Service: "web",
	}

	triggered, value, err := evaluator.Evaluate(context.Background(), rule)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if triggered {
		t.Error("Expected rule not to trigger")
	}

	if value != 1 {
		t.Errorf("Expected value 1, got %f", value)
	}
}

func TestEvaluateAvgOverSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/metrics/range" {
			t.Errorf("Expected path /api/metrics/range, got %s", r.URL.Path)
		}

		response := shared.QueryRangeResponse{
			Service: "web",
			Target:  "test",
			Name:    "http_latency_ms",
			Data: []shared.DataPoint{
				{Timestamp: time.Now().Unix() - 300, Value: 400},
				{Timestamp: time.Now().Unix() - 240, Value: 450},
				{Timestamp: time.Now().Unix() - 180, Value: 500},
				{Timestamp: time.Now().Unix() - 120, Value: 480},
				{Timestamp: time.Now().Unix() - 60, Value: 470},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	evaluator := NewEvaluator(server.URL)
	rule := &Rule{
		Name:    "latency-high",
		Expr:    "avg_over(5m, http_latency_ms) > 400",
		Service: "web",
	}

	triggered, value, err := evaluator.Evaluate(context.Background(), rule)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if !triggered {
		t.Error("Expected rule to trigger (avg > 400)")
	}

	if value <= 400 {
		t.Errorf("Expected value > 400, got %f", value)
	}
}

func TestEvaluateAvgOverNotTriggered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := shared.QueryRangeResponse{
			Service: "web",
			Name:    "http_latency_ms",
			Data: []shared.DataPoint{
				{Timestamp: time.Now().Unix() - 300, Value: 100},
				{Timestamp: time.Now().Unix() - 240, Value: 150},
				{Timestamp: time.Now().Unix() - 180, Value: 120},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	evaluator := NewEvaluator(server.URL)
	rule := &Rule{
		Name:    "latency-high",
		Expr:    "avg_over(5m, http_latency_ms) > 400",
		Service: "web",
	}

	triggered, _, err := evaluator.Evaluate(context.Background(), rule)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if triggered {
		t.Error("Expected rule not to trigger (avg < 400)")
	}
}

func TestEvaluateZScoreSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := shared.QueryRangeResponse{
			Service: "web",
			Name:    "http_rps",
			Data: []shared.DataPoint{
				{Timestamp: time.Now().Unix() - 300, Value: 100},
				{Timestamp: time.Now().Unix() - 240, Value: 100},
				{Timestamp: time.Now().Unix() - 180, Value: 100},
				{Timestamp: time.Now().Unix() - 120, Value: 100},
				{Timestamp: time.Now().Unix() - 60, Value: 100},
				{Timestamp: time.Now().Unix(), Value: 10000},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	evaluator := NewEvaluator(server.URL)
	rule := &Rule{
		Name:    "rps-spike",
		Expr:    "zscore(5m, http_rps) > 2",
		Service: "web",
	}

	triggered, value, err := evaluator.Evaluate(context.Background(), rule)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if !triggered {
		t.Errorf("Expected rule to trigger (spike detected), z-score: %f", value)
	}

	if value <= 2 {
		t.Errorf("Expected z-score > 2, got %f", value)
	}
}

func TestEvaluateAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	evaluator := NewEvaluator(server.URL)
	rule := &Rule{
		Name:    "test-rule",
		Expr:    "last(1m, http_up) == 0",
		Service: "web",
	}

	_, _, err := evaluator.Evaluate(context.Background(), rule)
	if err == nil {
		t.Error("Expected error for API failure")
	}
}

func TestEvaluateMetricNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	evaluator := NewEvaluator(server.URL)
	rule := &Rule{
		Name:    "test-rule",
		Expr:    "last(1m, nonexistent_metric) > 0",
		Service: "web",
	}

	triggered, value, err := evaluator.Evaluate(context.Background(), rule)
	if err != nil {
		t.Fatalf("Expected no error for 404, got: %v", err)
	}

	if triggered {
		t.Errorf("Expected rule not to trigger for missing metric, got value: %f", value)
	}

	if value != 0 {
		t.Errorf("Expected value 0 for missing metric, got %f", value)
	}
}

func TestEvaluateInvalidExpression(t *testing.T) {
	evaluator := NewEvaluator("http://localhost")
	rule := &Rule{
		Name: "invalid-rule",
		Expr: "invalid_function(1m, metric) > 0",
	}

	_, _, err := evaluator.Evaluate(context.Background(), rule)
	if err == nil {
		t.Error("Expected error for invalid expression")
	}
}

func TestEvaluateEmptyDataPoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := shared.QueryRangeResponse{
			Service: "web",
			Name:    "http_latency_ms",
			Data:    []shared.DataPoint{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	evaluator := NewEvaluator(server.URL)
	rule := &Rule{
		Name:    "test-rule",
		Expr:    "avg_over(5m, http_latency_ms) > 400",
		Service: "web",
	}

	triggered, value, err := evaluator.Evaluate(context.Background(), rule)
	if err != nil {
		t.Fatalf("Expected no error for empty data, got: %v", err)
	}

	if triggered {
		t.Error("Expected rule not to trigger for empty data")
	}

	if value != 0 {
		t.Errorf("Expected value 0 for empty data, got %f", value)
	}
}
