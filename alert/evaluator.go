package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"argos/shared"
)

type Evaluator struct {
	apiURL string
	client *http.Client
}

func NewEvaluator(apiURL string) *Evaluator {
	return &Evaluator{
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (e *Evaluator) Evaluate(ctx context.Context, rule *Rule) (bool, float64, error) {
	exprLower := strings.ToLower(rule.Expr)

	if strings.Contains(exprLower, "avg_over") {
		return e.evaluateAvgOver(ctx, rule)
	}

	if strings.Contains(exprLower, "last") {
		return e.evaluateLast(ctx, rule)
	}

	if strings.Contains(exprLower, "zscore") {
		return e.evaluateZScore(ctx, rule)
	}

	return false, 0, fmt.Errorf("unsupported expression: %s", rule.Expr)
}

func (e *Evaluator) evaluateAvgOver(ctx context.Context, rule *Rule) (bool, float64, error) {
	re := regexp.MustCompile(`avg_over\(([^,]+),\s*([^)]+)\)\s*([><=!]+)\s*([0-9.]+)`)
	matches := re.FindStringSubmatch(rule.Expr)
	if len(matches) < 5 {
		return false, 0, fmt.Errorf("invalid avg_over expression")
	}

	duration := matches[1]
	metricName := strings.TrimSpace(matches[2])
	operator := matches[3]
	thresholdStr := matches[4]

	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		return false, 0, err
	}

	avgValue, err := e.getAvgMetric(ctx, metricName, rule.Service, rule.Target, duration)
	if err != nil {
		return false, 0, err
	}

	triggered := compareValues(avgValue, operator, threshold)
	return triggered, avgValue, nil
}

func (e *Evaluator) evaluateLast(ctx context.Context, rule *Rule) (bool, float64, error) {
	re := regexp.MustCompile(`last\(([^,]+),\s*([^)]+)\)\s*([><=!]+)\s*([0-9.]+)`)
	matches := re.FindStringSubmatch(rule.Expr)
	if len(matches) < 5 {
		return false, 0, fmt.Errorf("invalid last expression")
	}

	metricName := strings.TrimSpace(matches[2])
	operator := matches[3]
	thresholdStr := matches[4]

	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		return false, 0, err
	}

	lastValue, err := e.getLastMetric(ctx, metricName, rule.Service, rule.Target)
	if err != nil {
		return false, 0, err
	}

	triggered := compareValues(lastValue, operator, threshold)
	return triggered, lastValue, nil
}

func (e *Evaluator) evaluateZScore(ctx context.Context, rule *Rule) (bool, float64, error) {
	re := regexp.MustCompile(`zscore\(([^,]+),\s*([^)]+)\)\s*([><=!]+)\s*([0-9.]+)`)
	matches := re.FindStringSubmatch(rule.Expr)
	if len(matches) < 5 {
		return false, 0, fmt.Errorf("invalid zscore expression")
	}

	duration := matches[1]
	metricName := strings.TrimSpace(matches[2])
	operator := matches[3]
	thresholdStr := matches[4]

	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		return false, 0, err
	}

	dataPoints, err := e.getRangeMetrics(ctx, metricName, rule.Service, rule.Target, duration)
	if err != nil {
		return false, 0, err
	}

	if len(dataPoints) < 2 {
		return false, 0, nil
	}

	values := make([]float64, len(dataPoints))
	for i, dp := range dataPoints {
		values[i] = dp.Value
	}

	currentValue := values[len(values)-1]
	zscore := shared.CalculateZScore(currentValue, values)

	triggered := compareValues(zscore, operator, threshold)
	return triggered, zscore, nil
}

func (e *Evaluator) getLastMetric(ctx context.Context, name, service, target string) (float64, error) {
	params := url.Values{}
	params.Add("name", name)
	if service != "" {
		params.Add("service", service)
	}
	if target != "" {
		params.Add("target", target)
	}

	reqURL := fmt.Sprintf("%s/api/metrics/query?%s", e.apiURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return 0, nil
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var metric shared.Metric
	if err := json.NewDecoder(resp.Body).Decode(&metric); err != nil {
		return 0, err
	}

	return metric.Value, nil
}

func (e *Evaluator) getAvgMetric(ctx context.Context, name, service, target, duration string) (float64, error) {
	dataPoints, err := e.getRangeMetrics(ctx, name, service, target, duration)
	if err != nil {
		return 0, err
	}

	if len(dataPoints) == 0 {
		return 0, nil
	}

	values := make([]float64, len(dataPoints))
	for i, dp := range dataPoints {
		values[i] = dp.Value
	}

	return shared.AggregateMetrics(values), nil
}

func (e *Evaluator) getRangeMetrics(ctx context.Context, name, service, target, duration string) ([]shared.DataPoint, error) {
	params := url.Values{}
	params.Add("name", name)
	params.Add("start", "-"+duration)
	if service != "" {
		params.Add("service", service)
	}
	if target != "" {
		params.Add("target", target)
	}

	reqURL := fmt.Sprintf("%s/api/metrics/range?%s", e.apiURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var result shared.QueryRangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func compareValues(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return math.Abs(value-threshold) < 0.0001
	case "!=":
		return math.Abs(value-threshold) >= 0.0001
	default:
		return false
	}
}
