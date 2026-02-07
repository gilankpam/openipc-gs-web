package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/openipc/ezconfig/internal/config"
	"github.com/openipc/ezconfig/internal/models"
)

type ConfigService struct {
	config    *config.ServiceConfig
	initDPath string
}

func NewConfigService(cfg *config.ServiceConfig) *ConfigService {
	path := "/etc/init.d/"
	if p, ok := os.LookupEnv("INIT_D_PATH"); ok {
		path = p
	}
	return &ConfigService{
		config:    cfg,
		initDPath: path,
	}
}

// --- Radio (WFB) ---

func (s *ConfigService) GetRadioSettings() (*models.RadioSettings, error) {
	wfb, err := s.config.LoadWFB()
	if err != nil {
		return nil, err
	}

	// User requirement was: use channel instead of frequency
	return &models.RadioSettings{
		Channel:   &wfb.Wireless.Channel,
		Bandwidth: &wfb.Wireless.Width,
		TxPower:   &wfb.Wireless.TxPower,
		McsIndex:  &wfb.Broadcast.McsIndex,
		FecK:      &wfb.Broadcast.FecK,
		FecN:      &wfb.Broadcast.FecN,
	}, nil
}

func (s *ConfigService) UpdateRadioSettings(settings *models.RadioSettings) error {
	wfb, err := s.config.LoadWFB()
	if err != nil {
		return err
	}

	// Update fields if present
	if settings.Channel != nil {
		wfb.Wireless.Channel = *settings.Channel
	}
	if settings.Bandwidth != nil {
		wfb.Wireless.Width = *settings.Bandwidth
	}
	if settings.TxPower != nil {
		wfb.Wireless.TxPower = *settings.TxPower
	}
	if settings.McsIndex != nil {
		wfb.Broadcast.McsIndex = *settings.McsIndex
	}
	if settings.FecK != nil {
		wfb.Broadcast.FecK = *settings.FecK
	}
	if settings.FecN != nil {
		wfb.Broadcast.FecN = *settings.FecN
	}

	if err := s.config.SaveWFB(wfb); err != nil {
		return err
	}

	s.restartWFBAsync()
	return nil
}

// --- Video (Majestic) ---

func (s *ConfigService) GetVideoSettings() (*models.VideoSettings, error) {
	conf, err := s.config.LoadMajestic()
	if err != nil {
		return nil, err
	}

	return &models.VideoSettings{
		Resolution: &conf.Video0.Size,
		Fps:        &conf.Video0.Fps,
		Codec:      &conf.Video0.Codec,
		Bitrate:    &conf.Video0.Bitrate,
		GopSize:    &conf.Video0.GopSize,
	}, nil
}

func (s *ConfigService) UpdateVideoSettings(settings *models.VideoSettings) error {
	conf, err := s.config.LoadMajestic()
	if err != nil {
		return err
	}

	if settings.Resolution != nil {
		conf.Video0.Size = *settings.Resolution
	}
	if settings.Fps != nil {
		conf.Video0.Fps = *settings.Fps
		conf.Isp.Exposure = 1000 / conf.Video0.Fps
	}
	if settings.Codec != nil {
		conf.Video0.Codec = *settings.Codec
	}
	if settings.Bitrate != nil {
		conf.Video0.Bitrate = *settings.Bitrate
	}
	if settings.GopSize != nil {
		conf.Video0.GopSize = *settings.GopSize
	}

	if err := s.config.SaveMajestic(conf); err != nil {
		return err
	}

	return s.restartService("S95majestic")
}

// --- Camera (Majestic) ---

func (s *ConfigService) GetCameraSettings() (*models.CameraSettings, error) {
	conf, err := s.config.LoadMajestic()
	if err != nil {
		return nil, err
	}

	return &models.CameraSettings{
		Contrast:   &conf.Image.Contrast,
		Saturation: &conf.Image.Saturation,
		Flip:       &conf.Image.Flip,
		Mirror:     &conf.Image.Mirror,
		Rotate:     &conf.Image.Rotate,
	}, nil
}

func (s *ConfigService) UpdateCameraSettings(settings *models.CameraSettings) error {
	conf, err := s.config.LoadMajestic()
	if err != nil {
		return err
	}

	if settings.Contrast != nil {
		conf.Image.Contrast = *settings.Contrast
	}
	if settings.Saturation != nil {
		conf.Image.Saturation = *settings.Saturation
	}
	if settings.Flip != nil {
		conf.Image.Flip = *settings.Flip
	}
	if settings.Mirror != nil {
		conf.Image.Mirror = *settings.Mirror
	}
	if settings.Rotate != nil {
		conf.Image.Rotate = *settings.Rotate
	}

	if err := s.config.SaveMajestic(conf); err != nil {
		return err
	}

	// Majestic usually needs restart or reload for image settings
	return s.restartService("S95majestic")
}

// --- Telemetry (WFB) ---

func (s *ConfigService) GetTelemetrySettings() (*models.TelemetrySettings, error) {
	wfb, err := s.config.LoadWFB()
	if err != nil {
		return nil, err
	}

	baudRate := 115200 // Placeholder

	return &models.TelemetrySettings{
		SerialPort: &wfb.Telemetry.Serial,
		Router:     &wfb.Telemetry.Router,
		BaudRate:   &baudRate,
	}, nil
}

func (s *ConfigService) UpdateTelemetrySettings(settings *models.TelemetrySettings) error {
	wfb, err := s.config.LoadWFB()
	if err != nil {
		return err
	}

	if settings.SerialPort != nil {
		wfb.Telemetry.Serial = *settings.SerialPort
	}
	if settings.Router != nil {
		wfb.Telemetry.Router = *settings.Router
	}
	// BaudRate might not be in wfb.yaml directly

	if err := s.config.SaveWFB(wfb); err != nil {
		return err
	}

	s.restartWFBAsync()
	return nil
}

// --- Adaptive Link (Alink) ---

func (s *ConfigService) GetAdaptiveLinkSettings() (*models.AdaptiveLinkSettings, error) {
	config, err := s.config.LoadAlink()
	if err != nil {
		return nil, err
	}

	// Check if enabled in rc.local
	enabled, err := s.isAlinkEnabledInRcLocal()
	if err != nil {
		return nil, err
	}

	// Map AlinkConfig to AdaptiveLinkSettings (API model)
	return &models.AdaptiveLinkSettings{
		Enabled:          &enabled,
		AllowSetPower:    &config.AllowSetPower,
		Use0To4TxPower:   &config.Use0To4TxPower,
		PowerLevel0To4:   &config.PowerLevel0To4,
		AllowSpikeFixFps: &config.AllowSpikeFixFps,
		OsdLevel:         &config.OsdLevel,
	}, nil
}

func (s *ConfigService) UpdateAdaptiveLinkSettings(settings *models.AdaptiveLinkSettings) error {
	// 1. Load existing config
	config, err := s.config.LoadAlink()
	if err != nil {
		return err
	}

	// 2. Update fields from API settings
	if settings.AllowSetPower != nil {
		config.AllowSetPower = *settings.AllowSetPower
	}
	if settings.Use0To4TxPower != nil {
		config.Use0To4TxPower = *settings.Use0To4TxPower
	}
	if settings.PowerLevel0To4 != nil {
		config.PowerLevel0To4 = *settings.PowerLevel0To4
	}
	if settings.AllowSpikeFixFps != nil {
		config.AllowSpikeFixFps = *settings.AllowSpikeFixFps
	}
	if settings.OsdLevel != nil {
		config.OsdLevel = *settings.OsdLevel
	}

	// 3. Save config
	if err := s.config.SaveAlink(config); err != nil {
		return err
	}

	// 4. Manage Enabled state in rc.local and process
	shouldBeEnabled := false
	if settings.Enabled != nil {
		shouldBeEnabled = *settings.Enabled
	} else {
		// If not specified in request, check current state
		var err error
		shouldBeEnabled, err = s.isAlinkEnabledInRcLocal()
		if err != nil {
			return err
		}
	}

	if shouldBeEnabled {
		if err := s.enableAlinkInRcLocal(); err != nil {
			return err
		}
		// Always restart if enabled, to apply new config
		_ = s.killAlinkProcess()
		return s.startAlinkProcess()
	} else {
		if err := s.disableAlinkInRcLocal(); err != nil {
			return err
		}
		return s.killAlinkProcess()
	}
}

// Helper functions for Alink

func (s *ConfigService) isAlinkEnabledInRcLocal() (bool, error) {
	content, err := os.ReadFile(s.config.RcLocalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return strings.Contains(string(content), "alink_drone"), nil
}

func (s *ConfigService) enableAlinkInRcLocal() error {
	content, err := os.ReadFile(s.config.RcLocalPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	text := string(content)
	if strings.Contains(text, "alink_drone") {
		return nil // Already enabled
	}

	// Insert before exit 0 if present
	if idx := strings.LastIndex(text, "exit 0"); idx != -1 {
		text = text[:idx] + "alink_drone &\n" + text[idx:]
	} else {
		// Append to rc.local
		if len(text) > 0 && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += "alink_drone &\n"
	}

	// Ensure rc.local is executable if we are creating it
	if err := os.WriteFile(s.config.RcLocalPath, []byte(text), 0755); err != nil {
		return err
	}
	return nil
}

func (s *ConfigService) disableAlinkInRcLocal() error {
	content, err := os.ReadFile(s.config.RcLocalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	for _, line := range lines {
		if strings.Contains(line, "alink_drone") {
			continue
		}
		newLines = append(newLines, line)
	}

	output := strings.Join(newLines, "\n")
	return os.WriteFile(s.config.RcLocalPath, []byte(output), 0755)
}

func (s *ConfigService) startAlinkProcess() error {
	cmd := exec.Command("sh", "-c", "alink_drone &")
	// If testing, we might want to mock the command or use a dummy script
	if s.initDPath != "/etc/init.d/" {
		// Testing mode: use dummy command
		cmd = exec.Command("sh", "-c", "echo 'Starting alink_drone'")
	}
	return cmd.Start()
}

func (s *ConfigService) killAlinkProcess() error {
	// Pkill is safer
	cmd := exec.Command("killall", "-9", "alink_drone")
	if s.initDPath != "/etc/init.d/" {
		// Testing mode
		cmd = exec.Command("sh", "-c", "echo 'Killing alink_drone'")
	}
	return cmd.Run()
}

// --- System ---

func (s *ConfigService) runServiceCommand(name, action string) error {
	cmd := exec.Command(filepath.Join(s.initDPath, name), action)
	return cmd.Run()
}

func (s *ConfigService) restartService(name string) error {
	return s.runServiceCommand(name, "restart")
}

func (s *ConfigService) restartWFBAsync() {
	go func() {
		// Wait a bit to ensure API response is sent
		time.Sleep(1 * time.Second)

		// S98wifibroadcast doesn't have restart, so stop then start
		_ = s.runServiceCommand("S98wifibroadcast", "stop")
		time.Sleep(1 * time.Second)
		_ = s.runServiceCommand("S98wifibroadcast", "start")
	}()
}

// --- TxProfiles ---

func (s *ConfigService) GetTxProfiles() ([]models.TxProfile, error) {
	return s.config.LoadTxProfiles()
}

func (s *ConfigService) UpdateTxProfiles(profiles []models.TxProfile) error {
	if err := s.config.SaveTxProfiles(profiles); err != nil {
		return err
	}

	// Restart alink if enabled to apply new profiles
	enabled, err := s.isAlinkEnabledInRcLocal()
	if err != nil {
		return err
	}

	if enabled {
		_ = s.killAlinkProcess()
		return s.startAlinkProcess()
	}

	return nil
}
