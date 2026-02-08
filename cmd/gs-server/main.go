package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gilankpam/openipc-gs-web/gs"
)

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

	// Initialize Radio Handler
	radioHandler := gs.NewRadioHandler(proxy, *configFile)

	// Serve Static Files or Proxy API
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Log request
		log.Printf("%s %s", r.Method, r.URL.Path)

		// Check if it's an API request
		if strings.HasPrefix(r.URL.Path, "/api/") {
			// Special handling for Radio Settings update
			if r.URL.Path == "/api/v1/radio" {
				radioHandler.ServeHTTP(w, r)
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
