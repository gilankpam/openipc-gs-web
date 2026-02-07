package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/openipc/ezconfig/internal/models"
)

type CapturingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *CapturingResponseWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func main() {
	var (
		listenAddr  = flag.String("listen", ":8081", "Address to listen on")
		airUnitAddr = flag.String("airunit", "http://192.168.1.10:8080", "Address of the Air Unit API")
		staticDir   = flag.String("static", "./web/dist", "Directory containing static frontend files")
		configFile  = flag.String("config", "/etc/wifibroadcast.cfg", "Path to wifibroadcast.cfg")
	)
	flag.Parse()

	// Parse Air Unit URL
	airUnitURL, err := url.Parse(*airUnitAddr)
	if err != nil {
		log.Fatalf("Invalid Air Unit URL: %v", err)
	}

	// Create API Proxy
	proxy := httputil.NewSingleHostReverseProxy(airUnitURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Ensure Host header matches the target
		req.Host = airUnitURL.Host
	}

	// Serve Static Files or Proxy API
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Log request
		log.Printf("%s %s", r.Method, r.URL.Path)

		// Check if it's an API request
		if strings.HasPrefix(r.URL.Path, "/api/") {
			// Special handling for Radio Settings update
			if r.Method == http.MethodPost && r.URL.Path == "/api/v1/radio" {
				handleRadioUpdate(w, r, proxy, *configFile)
				return
			}
			proxy.ServeHTTP(w, r)
			return
		}

		// Serve Static Files
		// If file exists, serve it. Otherwise, serve index.html (SPA)
		path := filepath.Join(*staticDir, r.URL.Path)
		// Clean the path to prevent directory traversal
		path = filepath.Clean(path)

		// Check if file exists
		info, err := os.Stat(path) // Use os.Stat, not fs.Stat
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}

		// Fallback to index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(*staticDir, "index.html"))
	})

	log.Printf("Starting GS Server on %s", *listenAddr)
	log.Printf("Proxying API requests to %s", *airUnitAddr)
	log.Printf("Serving static files from %s", *staticDir)
	log.Printf("Managing local config at %s", *configFile)

	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleRadioUpdate(w http.ResponseWriter, r *http.Request, proxy *httputil.ReverseProxy, configPath string) {
	// 1. Read the body to capture settings
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Restore the body for the proxy
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 2. Proxy the request and capture the response status
	cw := &CapturingResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
	proxy.ServeHTTP(cw, r)

	// 3. If the Air Unit update was successful (2xx), update local config
	if cw.StatusCode >= 200 && cw.StatusCode < 300 {
		var settings models.RadioSettings
		if err := json.Unmarshal(bodyBytes, &settings); err != nil {
			log.Printf("Error unmarshaling radio settings: %v", err)
			return
		}

		// Update local config file
		updated, err := updateLocalConfig(configPath, settings)
		if err != nil {
			log.Printf("Error updating local config: %v", err)
			return
		}

		if updated {
			// Restart service
			if err := restartService(); err != nil {
				log.Printf("Error restarting wifibroadcast service: %v", err)
			} else {
				log.Printf("Successfully updated local config and restarted service")
			}
		} else {
			log.Printf("Local config already matches, no restart needed")
		}
	}
}

func updateLocalConfig(path string, settings models.RadioSettings) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("failed to read config file: %w", err)
	}

	text := string(content)
	updated := false

	// Update Channel
	if settings.Channel != nil {
		re := regexp.MustCompile(`(?m)^(\s*wifi_channel\s*=\s*)(\d+)(.*)$`)
		// Check for existing value
		matches := re.FindStringSubmatch(text)
		if len(matches) > 2 {
			currentVal, err := strconv.Atoi(matches[2])
			if err == nil && currentVal == *settings.Channel {
				// Value is the same, no update needed
			} else {
				// Value differs, update it
				text = re.ReplaceAllString(text, fmt.Sprintf("${1}%d${3}", *settings.Channel))
				updated = true
			}
		} else {
			// If key doesn't exist, we might want to append it, but for now assuming it exists
			log.Printf("Warning: wifi_channel key not found in config")
		}
	}

	if !updated {
		return false, nil
	}

	// Write back to file
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		return false, fmt.Errorf("failed to write config file: %w", err)
	}

	return true, nil
}

func restartService() error {
	fmt.Println("Restarting wifibroadcast service...")
	// cmd := exec.Command("/etc/init.d/S98wifibroadcast", "restart")
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	return fmt.Errorf("failed to restart service: %v, output: %s", err, string(output))
	// }
	return nil
}
