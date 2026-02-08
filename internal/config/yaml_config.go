package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gilankpam/openipc-gs-web/internal/models"
	"gopkg.in/yaml.v3"
)

// ServiceConfig holds paths to config files
type ServiceConfig struct {
	WFBPath        string
	MajesticPath   string
	AlinkPath      string
	RcLocalPath    string
	TxProfilesPath string
}

// NewServiceConfig creates a new config handler with default paths or from env
func NewServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		WFBPath:        getEnv("WFB_PATH", "/etc/wfb.yaml"),
		MajesticPath:   getEnv("MAJESTIC_PATH", "/etc/majestic.yaml"),
		AlinkPath:      getEnv("ALINK_PATH", "/etc/alink.conf"),
		RcLocalPath:    getEnv("RC_LOCAL_PATH", "/etc/rc.local"),
		TxProfilesPath: getEnv("TXPROFILES_PATH", "/etc/txprofiles.conf"),
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
	if err := s.saveYaml(s.WFBPath, config); err != nil {
		return fmt.Errorf("failed to save wfb config: %w", err)
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
	if err := s.saveYaml(s.MajesticPath, config); err != nil {
		return fmt.Errorf("failed to save majestic config: %w", err)
	}
	return nil
}

func (s *ServiceConfig) saveYaml(path string, config interface{}) error {
	var b bytes.Buffer
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	// Write to temp file first
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, b.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write temp config: %w", err)
	}

	// Rename temp file to actual file (atomic update)
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to replace config: %w", err)
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
