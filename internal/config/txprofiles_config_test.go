package config

import (
	"os"
	"testing"
)

func TestLoadTxProfiles(t *testing.T) {
	// Create a temporary file with sample content
	content := `# <ra - nge> <gi> <mcs> <fecK> <fecN> <bitrate> <gop> <Pwr> <roiQP> <bandwidth> <qpDelta>
999  -  999  long 0 2 3    1000 10 30   0,0,0,0 20 -12
1000 - 1050  long 0 2 3    2000 10 30   0,0,0,0 20 -12
`
	tmpfile, err := os.CreateTemp("", "txprofiles_test.conf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Create ServiceConfig with the temp file path
	cfg := &ServiceConfig{
		TxProfilesPath: tmpfile.Name(),
	}

	// Load profiles
	profiles, err := cfg.LoadTxProfiles()
	if err != nil {
		t.Fatalf("LoadTxProfiles failed: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("Expected 2 profiles, got %d", len(profiles))
	}

	// Check first profile
	p1 := profiles[0]
	if p1.RangeStart != 999 || p1.RangeEnd != 999 || p1.GI != "long" || p1.MCS != 0 ||
		p1.FecK != 2 || p1.FecN != 3 || p1.Bitrate != 1000 || p1.Gop != 10 ||
		p1.Pwr != 30 || p1.RoiQP != "0,0,0,0" || p1.Bandwidth != 20 || p1.QpDelta != -12 {
		t.Errorf("First profile mismatch: %+v", p1)
	}

	// Test Save
	// Modify a value
	profiles[0].Bitrate = 9999
	if err := cfg.SaveTxProfiles(profiles); err != nil {
		t.Fatalf("SaveTxProfiles failed: %v", err)
	}

	// Load again
	profiles2, err := cfg.LoadTxProfiles()
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	if profiles2[0].Bitrate != 9999 {
		t.Errorf("Expected bitrate 9999, got %d", profiles2[0].Bitrate)
	}
}
