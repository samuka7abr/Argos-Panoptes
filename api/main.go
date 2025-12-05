package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"argos/shared"
)

var (
	storage   StorageInterface
	startTime time.Time
)

func main() {
	startTime = time.Now()

	// Configuração
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://argos:argos123@localhost:5432/argos?sslmode=disable"
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8081"
	}

	// Conectar ao banco
	var err error
	storage, err = NewStorage(databaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	log.Printf("Connected to database successfully")

	// Rotas
	http.HandleFunc("/ingest", ingestHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/metrics/latest", latestMetricsHandler)
	http.HandleFunc("/api/metrics/query", queryHandler)
	http.HandleFunc("/api/metrics/range", queryRangeHandler)
	http.HandleFunc("/api/metrics/services", listServicesHandler)
	http.HandleFunc("/api/metrics/targets", listTargetsHandler)
	http.HandleFunc("/api/alerts/active", activeAlertsHandler)
	http.HandleFunc("/api/alert-rules", alertsHandler)
	http.HandleFunc("/api/alert-rules/", alertsHandler)

	// Security endpoints
	http.HandleFunc("/api/security/events", securityEventsHandler)
	http.HandleFunc("/api/security/failed-logins", failedLoginsHandler)
	http.HandleFunc("/api/security/config-changes", configChangesHandler)
	http.HandleFunc("/api/security/vulnerabilities", vulnerabilitiesHandler)
	http.HandleFunc("/api/security/stats", securityStatsHandler)
	http.HandleFunc("/api/security/record-event", recordSecurityEventHandler)
	http.HandleFunc("/api/security/record-failed-login", recordFailedLoginHandler)
	http.HandleFunc("/api/security/record-config-change", recordConfigChangeHandler)
	http.HandleFunc("/api/security/record-vulnerability", recordVulnerabilityHandler)

	// CORS middleware
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		http.NotFound(w, r)
	})

	// Iniciar servidor
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Argos API listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// ingestHandler recebe métricas dos agentes
func ingestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var batch shared.Batch
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if len(batch.Items) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	if err := storage.InsertMetrics(batch.AgentID, batch.Items); err != nil {
		log.Printf("Failed to insert metrics: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Received %d metrics from agent %s", len(batch.Items), batch.AgentID)
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "accepted",
		"count":    len(batch.Items),
		"agent_id": batch.AgentID,
	})
}

// healthHandler retorna o status de saúde da API
func healthHandler(w http.ResponseWriter, r *http.Request) {
	count, _ := storage.GetMetricsCount()
	lastIngest, _ := storage.GetLastIngestTime()

	status := "ok"
	if time.Since(lastIngest) > 5*time.Minute {
		status = "degraded"
	}

	health := shared.HealthResponse{
		Status:       status,
		Uptime:       shared.FormatUptime(time.Since(startTime)),
		MetricsCount: count,
		LastIngest:   lastIngest,
		Version:      "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// queryHandler consulta o último valor de uma métrica ou série temporal se duration for fornecido
func queryHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	metricName := r.URL.Query().Get("metric_name") // Frontend usa metric_name
	if name == "" {
		name = metricName
	}
	service := r.URL.Query().Get("service")
	target := r.URL.Query().Get("target")
	duration := r.URL.Query().Get("duration")

	// Se duration for fornecido, retornar série temporal
	if duration != "" {
		var start time.Time
		var err error

		// Parse duration (ex: "1h", "30m", "24h")
		if duration[0] == '-' {
			start, err = shared.ParseRelativeTime(duration)
		} else {
			// Se não começar com -, assumir que é relativo (ex: "1h" -> "-1h")
			start, err = shared.ParseRelativeTime("-" + duration)
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid duration: %v", err), http.StatusBadRequest)
			return
		}

		end := time.Now()
		step := "1m" // Default step

		dataPoints, err := storage.QueryRange(name, service, target, start, end, step)
		if err != nil {
			log.Printf("Query range error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response := shared.QueryRangeResponse{
			Service: service,
			Target:  target,
			Name:    name,
			Data:    dataPoints,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Comportamento original: retornar último valor
	if name == "" {
		http.Error(w, "Parameter 'name' or 'metric_name' is required", http.StatusBadRequest)
		return
	}

	metric, err := storage.QueryLatest(name, service, target)
	if err != nil {
		log.Printf("Query error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if metric == nil {
		http.Error(w, "No data found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metric)
}

// queryRangeHandler consulta uma série temporal
func queryRangeHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	service := r.URL.Query().Get("service")
	target := r.URL.Query().Get("target")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	step := r.URL.Query().Get("step")

	if name == "" {
		http.Error(w, "Parameter 'name' is required", http.StatusBadRequest)
		return
	}

	// Parse start time (suporta relativo como "-1h" ou absoluto)
	var start time.Time
	var err error
	if startStr != "" {
		if startStr[0] == '-' {
			start, err = shared.ParseRelativeTime(startStr)
		} else {
			start, err = time.Parse(time.RFC3339, startStr)
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid start time: %v", err), http.StatusBadRequest)
			return
		}
	} else {
		start = time.Now().Add(-1 * time.Hour) // default: última hora
	}

	// Parse end time
	var end time.Time
	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid end time: %v", err), http.StatusBadRequest)
			return
		}
	} else {
		end = time.Now()
	}

	if step == "" {
		step = "1m"
	}

	dataPoints, err := storage.QueryRange(name, service, target, start, end, step)
	if err != nil {
		log.Printf("Query range error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := shared.QueryRangeResponse{
		Service: service,
		Target:  target,
		Name:    name,
		Data:    dataPoints,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func latestMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics, err := storage.GetLatestMetrics()
	if err != nil {
		log.Printf("Get latest metrics error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	grouped := make(map[string]map[string]interface{})

	for _, m := range metrics {
		key := m.Service + ":" + m.Target
		if grouped[key] == nil {
			grouped[key] = map[string]interface{}{
				"service":   m.Service,
				"target":    m.Target,
				"metrics":   make(map[string]float64),
				"timestamp": m.TS.Format(time.RFC3339),
			}
		}
		grouped[key]["metrics"].(map[string]float64)[m.Name] = m.Value
	}

	result := make([]map[string]interface{}, 0, len(grouped))
	for _, v := range grouped {
		result = append(result, v)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// listServicesHandler lista todos os serviços monitorados
func listServicesHandler(w http.ResponseWriter, r *http.Request) {
	services, err := storage.ListServices()
	if err != nil {
		log.Printf("List services error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"services": services,
	})
}

// listTargetsHandler lista todos os targets de um serviço
func listTargetsHandler(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	if service == "" {
		http.Error(w, "Parameter 'service' is required", http.StatusBadRequest)
		return
	}

	targets, err := storage.ListTargets(service)
	if err != nil {
		log.Printf("List targets error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": service,
		"targets": targets,
	})
}

// activeAlertsHandler retorna alertas ativos
func activeAlertsHandler(w http.ResponseWriter, r *http.Request) {
	alerts, err := storage.GetActiveAlerts()
	if err != nil {
		log.Printf("Get active alerts error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// Security handlers
func securityEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	events, err := storage.GetSecurityEvents(limit)
	if err != nil {
		log.Printf("Get security events error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"count":  len(events),
	})
}

func failedLoginsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	byIP, err := storage.GetFailedLoginsByIP(limit)
	if err != nil {
		log.Printf("Get failed logins error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	total, err := storage.GetTotalFailedLogins()
	if err != nil {
		log.Printf("Get total failed logins error: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"by_ip": byIP,
		"total": total,
	})
}

func configChangesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	changes, err := storage.GetConfigChanges(limit)
	if err != nil {
		log.Printf("Get config changes error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"changes": changes,
		"count":   len(changes),
	})
}

func vulnerabilitiesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vulns, err := storage.GetVulnerabilities()
	if err != nil {
		log.Printf("Get vulnerabilities error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vulnerabilities": vulns,
		"count":           len(vulns),
	})
}

func securityStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	failedLogins, _ := storage.GetTotalFailedLogins()
	anomalies, _ := storage.GetTrafficAnomalies(100)
	configChanges, _ := storage.GetConfigChanges(1)
	vulns, _ := storage.GetVulnerabilities()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"failed_logins":     failedLogins,
		"traffic_anomalies": anomalies,
		"config_changes":    len(configChanges),
		"vulnerabilities":   len(vulns),
	})
}

func recordSecurityEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event SecurityEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if event.Type == "" || event.Severity == "" || event.Description == "" {
		http.Error(w, "Missing required fields: type, severity, description", http.StatusBadRequest)
		return
	}

	if err := storage.CreateSecurityEvent(&event); err != nil {
		log.Printf("Create security event error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func recordFailedLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		IPAddress string `json:"ip_address"`
		Username  string `json:"username"`
		Service   string `json:"service"`
		UserAgent string `json:"user_agent"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if req.IPAddress == "" {
		http.Error(w, "Missing required field: ip_address", http.StatusBadRequest)
		return
	}

	if err := storage.RecordFailedLogin(req.IPAddress, req.Username, req.Service, req.UserAgent); err != nil {
		log.Printf("Record failed login error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}

func recordConfigChangeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var change ConfigChange
	if err := json.NewDecoder(r.Body).Decode(&change); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if change.FilePath == "" || change.ChangeType == "" {
		http.Error(w, "Missing required fields: file_path, change_type", http.StatusBadRequest)
		return
	}

	if err := storage.RecordConfigChange(&change); err != nil {
		log.Printf("Record config change error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(change)
}

func recordVulnerabilityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var vuln struct {
		Service     string `json:"service"`
		CVE         string `json:"cve"`
		Severity    string `json:"severity"`
		Description string `json:"description"`
		Version     string `json:"version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&vuln); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if vuln.Service == "" || vuln.Severity == "" {
		http.Error(w, "Missing required fields: service, severity", http.StatusBadRequest)
		return
	}

	// Inserir vulnerabilidade no banco
	if s, ok := storage.(*Storage); ok {
		_, err := s.DB().Exec(`
			INSERT INTO vulnerabilities (service, cve, severity, description, version)
			VALUES ($1, $2, $3, $4, $5)
		`, vuln.Service, vuln.CVE, vuln.Severity, vuln.Description, vuln.Version)

		if err != nil {
			log.Printf("Record vulnerability error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Storage not available", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "recorded"})
}
