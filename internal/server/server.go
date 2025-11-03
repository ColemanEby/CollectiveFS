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
}

// NewServerReqHandler creates a new HTTP request handler.
func NewServerReqHandler(programPath string) *ServerReqHandler {
	return &ServerReqHandler{
		programPath: programPath,
	}
}

// ServeHTTP implements http.Handler interface.
func (h *ServerReqHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

// doPOST handles POST requests, matching Python's do_POST.
func (h *ServerReqHandler) doPOST(w http.ResponseWriter, r *http.Request) {
	log.Println("in post method")

	h.setHeaders(w)

	// Read request body
	dataString, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON (matching Python's simplejson.loads)
	var data interface{}
	if err := json.Unmarshal(dataString, &data); err != nil {
		log.Printf("failed to parse JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Write to test123456.json (matching Python behavior)
	testFilePath := filepath.Join(h.programPath, "test123456.json")
	if err := os.WriteFile(testFilePath, dataString, 0644); err != nil {
		log.Printf("failed to write test file: %v", err)
	}

	log.Printf("%v", data)

	// Try to read for_presen.py and send it (matching Python behavior)
	presentPath := filepath.Join(h.programPath, "for_presen.py")
	if f, err := os.Open(presentPath); err == nil {
		defer f.Close()
		io.Copy(w, f)
	} else {
		// File doesn't exist, just return OK (matching Python behavior if file missing)
		w.WriteHeader(http.StatusOK)
	}
}

// setHeaders sets response headers, matching Python's _set_headers.
func (h *ServerReqHandler) setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-type", "text/html")
	w.WriteHeader(http.StatusOK)
}

// RunServer starts the HTTP server in a goroutine, matching Python's run_server.
func RunServer(host string, port int, programPath string) {
	handler := NewServerReqHandler(programPath)

	addr := fmt.Sprintf("%s:%d", host, port)
	httpd := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		log.Printf("Starting HTTP server on %s", addr)
		if err := httpd.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()
}
