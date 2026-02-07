package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWFBConfigRoundTrip(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "ezconfig_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	wfbPath := filepath.Join(tmpDir, "wfb.yaml")
	
	// Create initial YAML matching our struct definition
	initialYaml := `
wireless:
  txpower: 55
  channel: 161
  width: 20
  mlink: 1
  wlan_adapter: adapter
  link_control: ctrl
broadcast:
  mcs_index: 1
  tun_index: 2
  fec_k: 3
  fec_n: 4
  stbc: 1
  ldpc: 1
  link_id: 12345
telemetry:
  router: router
  serial: serial
  osd_fps: 30
`
	if err := os.WriteFile(wfbPath, []byte(initialYaml), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize ServiceConfig
	cfg := &ServiceConfig{
		WFBPath: wfbPath,
	}

	// Load config
	wfb, err := cfg.LoadWFB()
	if err != nil {
		t.Fatalf("LoadWFB failed: %v", err)
	}

	if wfb.Wireless.Channel != 161 {
		t.Errorf("Expected channel 161, got %d", wfb.Wireless.Channel)
	}

	// Modify a field
	wfb.Wireless.Channel = 36
	wfb.Broadcast.McsIndex = 5
	
	// Save config
	if err := cfg.SaveWFB(wfb); err != nil {
		t.Fatalf("SaveWFB failed: %v", err)
	}

	// Reload to verify
	wfb2, err := cfg.LoadWFB()
	if err != nil {
		t.Fatalf("LoadWFB failed 2nd time: %v", err)
	}

	if wfb2.Wireless.Channel != 36 {
		t.Errorf("Expected channel 36, got %d", wfb2.Wireless.Channel)
	}
	if wfb2.Broadcast.McsIndex != 5 {
		t.Errorf("Expected McsIndex 5, got %d", wfb2.Broadcast.McsIndex)
	}
	// Verify other fields preserved
	if wfb2.Wireless.TxPower != 55 {
		t.Errorf("Expected TxPower 55, got %d", wfb2.Wireless.TxPower)
	}
}

func TestAlinkConfigRoundTrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ezconfig_test_alink")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	alinkPath := filepath.Join(tmpDir, "alink.conf")
	
	initialConf := `
### Some comment
allow_set_power=1
use_0_to_4_txpower=0
power_level_0_to_4=3
some_unknown_key=value
`
	if err := os.WriteFile(alinkPath, []byte(initialConf), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &ServiceConfig{
		AlinkPath: alinkPath,
	}

	// Load
	alink, err := cfg.LoadAlink()
	if err != nil {
		t.Fatalf("LoadAlink failed: %v", err)
	}

	if !alink.AllowSetPower {
		t.Error("Expected AllowSetPower true")
	}
	if alink.Use0To4TxPower {
		t.Error("Expected Use0To4TxPower false")
	}
	if alink.PowerLevel0To4 != 3 {
		t.Errorf("Expected PowerLevel 3, got %d", alink.PowerLevel0To4)
	}

	// Update
	alink.PowerLevel0To4 = 4
	alink.AllowSpikeFixFps = true // Was missing in file, defaults false in struct, we set true

	// Save
	if err := cfg.SaveAlink(alink); err != nil {
		t.Fatalf("SaveAlink failed: %v", err)
	}

	// Check file content for comment validation and new keys
	bytes, err := os.ReadFile(alinkPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(bytes)

	if !strings.Contains(content, "### Some comment") {
		t.Error("Comment lost")
	}
	if !strings.Contains(content, "power_level_0_to_4=4") {
		t.Error("Power level not updated")
	}
	if !strings.Contains(content, "allow_spike_fix_fps=1") { // converted to 1
		t.Error("New key not added")
	}
	if !strings.Contains(content, "some_unknown_key=value") {
		t.Error("Unknown key lost (should be preserved by read/write loop)")
	}
}
