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
	http.HandleFunc("/api/metrics/query", queryHandler)
	http.HandleFunc("/api/metrics/range", queryRangeHandler)
	http.HandleFunc("/api/services", listServicesHandler)
	http.HandleFunc("/api/targets", listTargetsHandler)
	http.HandleFunc("/api/alerts/active", activeAlertsHandler)

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

// queryHandler consulta o último valor de uma métrica
func queryHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	service := r.URL.Query().Get("service")
	target := r.URL.Query().Get("target")

	if name == "" {
		http.Error(w, "Parameter 'name' is required", http.StatusBadRequest)
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
