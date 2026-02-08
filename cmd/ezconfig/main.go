package main

import (
	"log"
	"net/http"

	"github.com/gilankpam/openipc-gs-web/internal/air_unit/handler"
	"github.com/gilankpam/openipc-gs-web/internal/air_unit/service"
	"github.com/gilankpam/openipc-gs-web/internal/config"
)

func main() {
	// Initialize Config
	cfg := config.NewServiceConfig()

	// Initialize Service
	svc := service.NewConfigService(cfg)

	// Initialize Handler
	h := handler.NewHandler(svc)

	// Setup Routes
	mux := http.NewServeMux()

	// Radio
	mux.HandleFunc("/api/v1/radio", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetRadio(w, r)
		case http.MethodPost:
			h.UpdateRadio(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Video
	mux.HandleFunc("/api/v1/video", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetVideo(w, r)
		case http.MethodPost:
			h.UpdateVideo(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Camera
	mux.HandleFunc("/api/v1/camera", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetCamera(w, r)
		case http.MethodPost:
			h.UpdateCamera(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Telemetry
	mux.HandleFunc("/api/v1/telemetry", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetTelemetry(w, r)
		case http.MethodPost:
			h.UpdateTelemetry(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Adaptive Link
	mux.HandleFunc("/api/v1/adaptive-link", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetAdaptiveLink(w, r)
		case http.MethodPost:
			h.UpdateAdaptiveLink(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// TxProfiles
	mux.HandleFunc("/api/v1/txprofiles", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetTxProfiles(w, r)
		case http.MethodPost:
			h.UpdateTxProfiles(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Ping
	mux.HandleFunc("/api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	// Start Server
	log.Println("Starting OpenIPC EZConfig API on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
