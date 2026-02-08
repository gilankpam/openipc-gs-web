package service

import (
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

func TestGetStatsTCP(t *testing.T) {
	// Start mock server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Send initial stats
		// p['all'] = [100, 0]
		// antenna 1: [pkt=50, rssi_min=-65, rssi_avg=-60, rssi_max=-55, snr_min=15, snr_avg=20, snr_max=25]

		// Manual construction of RX Ant Stats MsgPack
		// Key: [[freq, mcs, bw], ant_id]
		// Value: [pkt_s, rssi_min, rssi_avg, rssi_max, snr_min, snr_avg, snr_max]

		antKey := []interface{}{
			[]int{5825, 1, 20}, // freq, mcs, bw
			1,                  // ant_id
		}

		antVal := []int{50, -65, -60, -55, 15, 20, 25}

		// Encode key and value separately
		keyBytes, _ := msgpack.Marshal(antKey)
		valBytes, _ := msgpack.Marshal(antVal)

		// Construct map payload: [0x81 (map size 1)][key][val]
		rxAntStatsPayload := append([]byte{0x81}, keyBytes...)
		rxAntStatsPayload = append(rxAntStatsPayload, valBytes...)

		msg := map[string]interface{}{
			"type": "rx",
			"packets": map[string][]int{
				"all":       {100, 0},
				"lost":      {5, 0},
				"all_bytes": {1024000, 0}, // ~1MB/s
			},
			"session": map[string]int{
				"fec_k": 8,
				"fec_n": 12,
			},
			"rx_ant_stats": msgpack.RawMessage(rxAntStatsPayload),
		}

		sendMsg(t, conn, msg)

		// Wait and send update
		time.Sleep(100 * time.Millisecond)

		// Update value
		antVal2 := []int{60, -65, -58, -55, 15, 22, 25}
		valBytes2, _ := msgpack.Marshal(antVal2)

		rxAntStatsPayload2 := append([]byte{0x81}, keyBytes...)
		rxAntStatsPayload2 = append(rxAntStatsPayload2, valBytes2...)

		msg["packets"] = map[string][]int{
			"all":       {120, 0},
			"lost":      {5, 0},
			"all_bytes": {2048000, 0},
		}
		msg["rx_ant_stats"] = msgpack.RawMessage(rxAntStatsPayload2)

		sendMsg(t, conn, msg)

		// Keep connection open
		time.Sleep(1 * time.Second)
	}()

	// Initialize service with test address
	s := NewWFBStatsService().WithAddress(ln.Addr().String())
	s.Start()
	defer s.Stop()

	// Wait for connection and data
	time.Sleep(200 * time.Millisecond)

	// Get stats
	stats, err := s.GetStats()
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	// Verify
	if stats.VideoPacketsPerSec != 120 {
		t.Errorf("Expected 120 packets/sec, got %d", stats.VideoPacketsPerSec)
	}

	// Verify transmission info
	if stats.Frequency != 5825 {
		t.Errorf("Expected Freq 5825, got %d", stats.Frequency)
	}
	if stats.McsIndex != 1 {
		t.Errorf("Expected MCS 1, got %d", stats.McsIndex)
	}
	if stats.Bandwidth != 20 {
		t.Errorf("Expected BW 20, got %d", stats.Bandwidth)
	}
	if stats.FecK != 8 || stats.FecN != 12 {
		t.Errorf("Expected FEC 8/12, got %d/%d", stats.FecK, stats.FecN)
	}
	if stats.LinkFlowBytesPerSec != 2048000 {
		t.Errorf("Expected Flow 2048000, got %d", stats.LinkFlowBytesPerSec)
	}
}

func sendMsg(t *testing.T, conn net.Conn, msg interface{}) {
	payload, err := msgpack.Marshal(msg)
	if err != nil {
		t.Errorf("Failed to marshal msg: %v", err)
		return
	}

	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(payload)))

	if _, err := conn.Write(header); err != nil {
		t.Logf("Write header failed: %v", err)
		return
	}
	if _, err := conn.Write(payload); err != nil {
		t.Logf("Write payload failed: %v", err)
		return
	}
}
