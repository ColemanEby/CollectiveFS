package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// StatusResponse represents the status API response.
type StatusResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime,omitempty"`
}

var startTime = time.Now()

// HandleStatus returns server status information.
func HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uptime := time.Since(startTime)
	response := StatusResponse{
		Status:    "running",
		Version:   "1.0.0",
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode status response: %v", err)
	}
}
