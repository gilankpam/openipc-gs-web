package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gilankpam/openipc-gs-web/internal/models"
)

// LoadTxProfiles loads TxProfiles from the specified file path
func (s *ServiceConfig) LoadTxProfiles() ([]models.TxProfile, error) {
	// For now, hardcode the path or prefer the one in ServiceConfig if it existed.
	// Since I didn't add TxProfilesPath to ServiceConfig yet, I'll use a default or assume it's passed.
	// Wait, I should have added it to ServiceConfig or passed it as argument.
	// Let's check ServiceConfig definition in config.go first.
	// For now let's assume I'll add it to ServiceConfig or I'll just use a constant for now if not present.
	// Actually, I should update ServiceConfig first. But for this file I'll assume s.TxProfilesPath exists.
	// I'll fix ServiceConfig in the next step.

	file, err := os.Open(s.TxProfilesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.TxProfile{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var profiles []models.TxProfile
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		// Expected format: start - end gi mcs feck fecn bitrate gop pwr roiqp bandwidth qpdelta
		// Count: 1 (start) + 1 (-) + 1 (end) + 1 (gi) + 1 (mcs) + 1 (feck) + 1 (fecn) + 1 (bitrate) + 1 (gop) + 1 (pwr) + 1 (roiqp) + 1 (bandwidth) + 1 (qpdelta) = 13 fields

		if len(fields) < 13 {
			// Maybe handle malformed lines gracefully or skip
			continue
		}

		// Parse fields
		start, _ := strconv.Atoi(fields[0])
		// fields[1] is "-"
		end, _ := strconv.Atoi(fields[2])
		gi := fields[3]
		mcs, _ := strconv.Atoi(fields[4])
		fecK, _ := strconv.Atoi(fields[5])
		fecN, _ := strconv.Atoi(fields[6])
		bitrate, _ := strconv.Atoi(fields[7])
		gop, _ := strconv.Atoi(fields[8])
		pwr, _ := strconv.Atoi(fields[9])
		roiQp := fields[10]
		bandwidth, _ := strconv.Atoi(fields[11])
		qpDelta, _ := strconv.Atoi(fields[12])

		profile := models.TxProfile{
			RangeStart: start,
			RangeEnd:   end,
			GI:         gi,
			MCS:        mcs,
			FecK:       fecK,
			FecN:       fecN,
			Bitrate:    bitrate,
			Gop:        gop,
			Pwr:        pwr,
			RoiQP:      roiQp,
			Bandwidth:  bandwidth,
			QpDelta:    qpDelta,
		}
		profiles = append(profiles, profile)
	}

	return profiles, scanner.Err()
}

// SaveTxProfiles saves the profiles to the file, overwriting it.
func (s *ServiceConfig) SaveTxProfiles(profiles []models.TxProfile) error {
	// We strictly overwrite and format as requested.
	// Format: "%d - %d %s %d %d %d %d %d %d %s %d %d"

	file, err := os.Create(s.TxProfilesPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header? User said "remove comments", but maybe a header is useful?
	// User said "You dont need to preserve the comment", which implies I can drop existing ones.
	// I'll stick to just data to be safe and minimalist as per "remove it".

	w := bufio.NewWriter(file)
	for _, p := range profiles {
		line := fmt.Sprintf("%d - %d %s %d %d %d %d %d %d %s %d %d\n",
			p.RangeStart, p.RangeEnd, p.GI, p.MCS, p.FecK, p.FecN,
			p.Bitrate, p.Gop, p.Pwr, p.RoiQP, p.Bandwidth, p.QpDelta)
		if _, err := w.WriteString(line); err != nil {
			return err
		}
	}
	return w.Flush()
}
