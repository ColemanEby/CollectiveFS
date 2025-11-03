package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// ServerReqHandler handles HTTP requests, matching Python's ServerReqHandler class.
type ServerReqHandler struct {
	programPath string
	configPath  string // Path to config directory for serving files
}

// NewServerReqHandler creates a new HTTP request handler.
func NewServerReqHandler(programPath, configPath string) *ServerReqHandler {
	return &ServerReqHandler{
		programPath: programPath,
		configPath:  configPath,
	}
}

// ServeHTTP implements http.Handler interface.
func (h *ServerReqHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle API endpoints
	if r.URL.Path == "/api/status" {
		HandleStatus(w, r)
		return
	}

	// Handle regular requests
	switch r.Method {
	case http.MethodGet:
		h.doGET(w, r)
	case http.MethodHead:
		h.doHEAD(w, r)
	case http.MethodPost:
		h.doPOST(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// doGET handles GET requests - serves index.html, matching Python's do_GET.
func (h *ServerReqHandler) doGET(w http.ResponseWriter, r *http.Request) {
	h.setHeaders(w)

	indexPath := filepath.Join(h.programPath, "index.html")
	f, err := os.Open(indexPath)
	if err != nil {
		log.Printf("failed to open index.html: %v", err)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	// Copy file contents to response
	if _, err := io.Copy(w, f); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

// doHEAD handles HEAD requests, matching Python's do_HEAD.
func (h *ServerReqHandler) doHEAD(w http.ResponseWriter, r *http.Request) {
	h.setHeaders(w)
}

// doPOST handles POST requests.
// For now, this is a placeholder for future API functionality.
// The original Python code wrote to test123456.json and returned for_presen.py,
// which were clearly test/debug artifacts, so we've removed those.
func (h *ServerReqHandler) doPOST(w http.ResponseWriter, r *http.Request) {
	// Read and parse request body
	dataString, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var data interface{}
	if err := json.Unmarshal(dataString, &data); err != nil {
		log.Printf("failed to parse JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received POST data: %v", data)

	// Return JSON response acknowledging receipt
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"status":  "received",
		"message": "POST request received (API functionality coming soon)",
	}
	json.NewEncoder(w).Encode(response)
}

// setHeaders sets response headers, matching Python's _set_headers.
func (h *ServerReqHandler) setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-type", "text/html")
	w.WriteHeader(http.StatusOK)
}

// RunServer starts the HTTP server in a goroutine, matching Python's run_server.
func RunServer(host string, port int, programPath, configPath string) {
	handler := NewServerReqHandler(programPath, configPath)

	addr := fmt.Sprintf("%s:%d", host, port)
	httpd := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		log.Printf("Starting HTTP server on %s", addr)
		log.Printf("  - Web UI: http://%s/", addr)
		log.Printf("  - Status API: http://%s/api/status", addr)
		if err := httpd.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()
}
