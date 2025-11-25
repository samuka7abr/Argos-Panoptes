package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type APIAlertRule struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Expr        string   `json:"expr"`
	Service     string   `json:"service"`
	Target      string   `json:"target"`
	ForDuration string   `json:"for_duration"`
	Severity    string   `json:"severity"`
	EmailTo     []string `json:"email_to"`
	Enabled     bool     `json:"enabled"`
}

type APIResponse struct {
	Rules []APIAlertRule `json:"rules"`
	Count int            `json:"count"`
}

func LoadRulesFromAPI(apiURL string) (*RulesConfig, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(apiURL + "/api/alert-rules")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rules from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	if len(apiResp.Rules) == 0 {
		return nil, fmt.Errorf("no rules found in API")
	}

	rules := make([]Rule, 0, len(apiResp.Rules))
	for _, apiRule := range apiResp.Rules {
		rules = append(rules, Rule{
			Name:        apiRule.Name,
			Description: apiRule.Description,
			Expr:        apiRule.Expr,
			Service:     apiRule.Service,
			Target:      apiRule.Target,
			For:         apiRule.ForDuration,
			Severity:    apiRule.Severity,
			EmailTo:     apiRule.EmailTo,
		})
	}

	cfg := &RulesConfig{
		Rules: rules,
		Email: EmailConfig{},
	}

	return cfg, nil
}

func LoadRulesHybrid(apiURL, yamlPath string) (*RulesConfig, error) {
	yamlCfg, yamlErr := LoadRules(yamlPath)

	apiCfg, apiErr := LoadRulesFromAPI(apiURL)
	if apiErr == nil && len(apiCfg.Rules) > 0 {
		apiCfg.Email = yamlCfg.Email
		return apiCfg, nil
	}

	if yamlErr != nil {
		return nil, fmt.Errorf("failed to load from API and YAML: api=%v, yaml=%v", apiErr, yamlErr)
	}

	return yamlCfg, nil
}
