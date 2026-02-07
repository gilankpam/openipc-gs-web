package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openipc/ezconfig/internal/models"
	"gopkg.in/yaml.v3"
)

// ServiceConfig holds paths to config files
type ServiceConfig struct {
	WFBPath      string
	MajesticPath string
	AlinkPath    string
	RcLocalPath  string
}

// NewServiceConfig creates a new config handler with default paths or from env
func NewServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		WFBPath:      getEnv("WFB_PATH", "/etc/wfb.yaml"),
		MajesticPath: getEnv("MAJESTIC_PATH", "/etc/majestic.yaml"),
		AlinkPath:    getEnv("ALINK_PATH", "/etc/alink.conf"),
		RcLocalPath:  getEnv("RC_LOCAL_PATH", "/etc/rc.local"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// LoadWFB loads WFB configuration
func (s *ServiceConfig) LoadWFB() (*models.WFBConfig, error) {
	data, err := os.ReadFile(s.WFBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wfb config: %w", err)
	}

	var config models.WFBConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wfb config: %w", err)
	}

	return &config, nil
}

// SaveWFB saves WFB configuration
func (s *ServiceConfig) SaveWFB(config *models.WFBConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal wfb config: %w", err)
	}

	// Write to temp file first
	tmpPath := s.WFBPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp wfb config: %w", err)
	}

	// Rename temp file to actual file (atomic update)
	if err := os.Rename(tmpPath, s.WFBPath); err != nil {
		return fmt.Errorf("failed to replace wfb config: %w", err)
	}

	return nil
}

// LoadMajestic loads Majestic configuration
func (s *ServiceConfig) LoadMajestic() (*models.MajesticConfig, error) {
	data, err := os.ReadFile(s.MajesticPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read majestic config: %w", err)
	}

	var config models.MajesticConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal majestic config: %w", err)
	}

	return &config, nil
}

// SaveMajestic saves Majestic configuration
func (s *ServiceConfig) SaveMajestic(config *models.MajesticConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal majestic config: %w", err)
	}

	tmpPath := s.MajesticPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp majestic config: %w", err)
	}

	if err := os.Rename(tmpPath, s.MajesticPath); err != nil {
		return fmt.Errorf("failed to replace majestic config: %w", err)
	}

	return nil
}

// Helper to ensure directory exists (mostly for testing/local dev)
func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}
