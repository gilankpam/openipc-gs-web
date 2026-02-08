package gs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/gilankpam/openipc-gs-web/internal/models"
)

type CapturingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
}

func (w *CapturingResponseWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *CapturingResponseWriter) Write(b []byte) (int, error) {
	if w.Body != nil {
		w.Body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

type RadioHandler struct {
	Proxy      *httputil.ReverseProxy
	ConfigPath string
}

func NewRadioHandler(proxy *httputil.ReverseProxy, configPath string) *RadioHandler {
	return &RadioHandler{
		Proxy:      proxy,
		ConfigPath: configPath,
	}
}

func (h *RadioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.handleGet(w, r)
		return
	}
	if r.Method == http.MethodPost {
		h.handlePost(w, r)
		return
	}
	h.Proxy.ServeHTTP(w, r)
}

func (h *RadioHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	// Clone proxy to intercept handling
	proxyCopy := *h.Proxy

	// Intercept 5xx responses to trigger fallback
	proxyCopy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode >= 500 {
			return fmt.Errorf("backend returned %d", resp.StatusCode)
		}
		return nil
	}

	proxyCopy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Printf("Proxy error for %s: %v. Falling back to local config.", req.URL.Path, err)
		h.serveLocalConfig(rw)
	}

	proxyCopy.ServeHTTP(w, r)
}

func (h *RadioHandler) serveLocalConfig(w http.ResponseWriter) {
	content, err := os.ReadFile(h.ConfigPath)
	// If file doesn't exist, we still want to show UI in "Local Mode" (maybe with defaults)
	// rather than failing completely.
	if err != nil {
		log.Printf("Failed to read local config at %s: %v. Using defaults.", h.ConfigPath, err)
		// Don't return error to client, return empty/default settings with local flag
		// This ensures UI loads and shows "Local Only" alert
	}

	// Parse channel from config (if content exists)
	channel := 0
	if content != nil {
		re := regexp.MustCompile(`(?m)^(\s*wifi_channel\s*=\s*)(\d+)(.*)$`)
		matches := re.FindStringSubmatch(string(content))
		if len(matches) > 2 {
			channel, _ = strconv.Atoi(matches[2])
		}
	}

	// Create partial RadioSettings
	settings := models.RadioSettings{
		Channel: &channel,
		// Other fields will be empty/null, frontend handles this
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-GS-Data-Source", "local")
	w.Header().Set("Access-Control-Expose-Headers", "X-GS-Data-Source")
	w.WriteHeader(http.StatusOK) // Return 200 OK
	json.NewEncoder(w).Encode(settings)
}

func (h *RadioHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	// 1. Read body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 2. Try proxy with capturing
	cw := &CapturingResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}

	proxySuccess := false
	proxyCopy := *h.Proxy

	// Trigger fallback on 5xx
	proxyCopy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode >= 500 {
			return fmt.Errorf("backend returned %d", resp.StatusCode)
		}
		return nil
	}

	proxyCopy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Printf("Proxy error for POST %s: %v. Treating as offline update.", req.URL.Path, err)
		// Mark as 502 internally so we check it later
		cw.StatusCode = http.StatusBadGateway
	}

	proxyCopy.ServeHTTP(cw, r)

	// Check result
	if cw.StatusCode >= 200 && cw.StatusCode < 300 {
		proxySuccess = true
	}

	// Parse settings
	var settings models.RadioSettings
	if err := json.Unmarshal(bodyBytes, &settings); err != nil {
		log.Printf("Error unmarshaling radio settings: %v", err)
		return
	}

	// Update local config
	updated, err := h.updateLocalConfig(settings)
	if err != nil {
		log.Printf("Error updating local config: %v", err)
		if !proxySuccess {
			// Only error out if BOTH proxy and local failed
			// Use original writer 'w' effectively via 'cw' if not already written?
			// If ErrorHandler was called, nothing was written to wire yet usually?
			// But cw captures it.
			// If proxySuccess is false, it means either ErrorHandler called (cw.StatusCode=502)
			// OR ModifyResponse error -> ErrorHandler called.
		}
		// If local update failed, we can't do much.
	} else if updated {
		if err := h.restartService(); err != nil {
			log.Printf("Error restarting service: %v", err)
		} else {
			log.Printf("Service restarted")
		}
	}

	// If proxy failed, we need to overwrite the failure with success
	if !proxySuccess {
		// If ErrorHandler was called, it might typically write a response?
		// We set ErrorHandler to JUST set status code in cw, not write to rw.
		// Wait, my ErrorHandler in handlePost DOES NOT write to rw.

		// If we haven't writtenheaders yet:
		if cw.Body == nil {
			// We need to write response now
			w.Header().Set("X-GS-Data-Source", "local")
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Expose-Headers", "X-GS-Data-Source")
			w.WriteHeader(http.StatusOK)
			w.Write(bodyBytes)
		}
	}
}

func (h *RadioHandler) updateLocalConfig(settings models.RadioSettings) (bool, error) {
	content, err := os.ReadFile(h.ConfigPath)
	if err != nil {
		return false, fmt.Errorf("failed to read config file: %w", err)
	}

	text := string(content)
	updated := false

	// Update Channel
	if settings.Channel != nil {
		re := regexp.MustCompile(`(?m)^(\s*wifi_channel\s*=\s*)(\d+)(.*)$`)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 2 {
			currentVal, err := strconv.Atoi(matches[2])
			if err == nil && currentVal == *settings.Channel {
				// Same value
			} else {
				text = re.ReplaceAllString(text, fmt.Sprintf("${1}%d${3}", *settings.Channel))
				updated = true
			}
		} else {
			// Append if missing? (Simplified: warn only)
			log.Printf("Warning: wifi_channel key not found in config")
		}
	}

	if !updated {
		return false, nil
	}

	if err := os.WriteFile(h.ConfigPath, []byte(text), 0644); err != nil {
		return false, fmt.Errorf("failed to write config file: %w", err)
	}

	return true, nil
}

func (h *RadioHandler) restartService() error {
	log.Println("Internal: Restarting wifibroadcast service...")
	// In production this might need sudo or specific permissions
	cmd := exec.Command("/etc/init.d/S98wifibroadcast", "restart")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart service: %v, output: %s", err, string(output))
	}
	return nil
}
