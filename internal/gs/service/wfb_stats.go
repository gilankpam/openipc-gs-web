package service

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

// WFBStats holds the aggregated statistics for the UI
type WFBStats struct {
	// Antenna stats
	Rssi []int8 `json:"rssi"`
	Snr  []int8 `json:"snr"`

	// Link stats (rates)
	VideoPacketsPerSec int `json:"video_packets_per_sec"`
	FecPacketsPerSec   int `json:"fec_packets_per_sec"`
	LostPacketsPerSec  int `json:"lost_packets_per_sec"`
	BadBlocksPerSec    int `json:"bad_blocks_per_sec"`

	// Transmission Info
	Frequency           uint32 `json:"frequency"`
	McsIndex            int    `json:"mcs_index"`
	Bandwidth           int    `json:"bandwidth"`
	FecK                int    `json:"fec_k"`
	FecN                int    `json:"fec_n"`
	LinkFlowBytesPerSec int    `json:"link_flow_bytes_per_sec"`

	// Raw stats for debug or accumulation
	TotalPackets uint32 `json:"total_packets"`
	TotalLost    uint32 `json:"total_lost"`
}

// Internal MsgPack structures
type WFBMessage struct {
	Type       string                 `msgpack:"type"`
	Packets    map[string][]int64     `msgpack:"packets"`      // keys: "all", "lost", "fec_rec", "bad", "all_bytes"
	RxAntStats msgpack.RawMessage     `msgpack:"rx_ant_stats"` // Complex key map, decode manually
	Session    map[string]interface{} `msgpack:"session"`
}

// WFBStatsService handles reading stats via TCP
type WFBStatsService struct {
	mu           sync.Mutex
	currentStats *WFBStats
	address      string
	running      bool
}

func NewWFBStatsService() *WFBStatsService {
	// Default port for GS stats in wfb-ng master.cfg is 8003
	return &WFBStatsService{
		address: "127.0.0.1:8003",
		currentStats: &WFBStats{
			Rssi: []int8{},
			Snr:  []int8{},
		},
	}
}

// WithAddress allows setting custom address for testing
func (s *WFBStatsService) WithAddress(addr string) *WFBStatsService {
	s.address = addr
	return s
}

func (s *WFBStatsService) Start() {
	s.running = true
	go s.runLoop()
}

func (s *WFBStatsService) Stop() {
	s.running = false
}

func (s *WFBStatsService) GetStats() (*WFBStats, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Return a copy
	stats := *s.currentStats
	return &stats, nil
}

func (s *WFBStatsService) runLoop() {
	for s.running {
		conn, err := net.DialTimeout("tcp", s.address, 2*time.Second)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		s.handleConnection(conn)
		// Connection closed, retry
		time.Sleep(1 * time.Second)
	}
}

func (s *WFBStatsService) handleConnection(conn net.Conn) {
	defer conn.Close()

	// wfb-ng sends: [4 bytes length][msgpack payload]
	header := make([]byte, 4)

	for s.running {
		// Read length
		if _, err := io.ReadFull(conn, header); err != nil {
			return
		}
		length := binary.BigEndian.Uint32(header)

		// Sanity check length
		if length > 1024*1024 { // 1MB limit
			return
		}

		// Read payload
		payload := make([]byte, length)
		if _, err := io.ReadFull(conn, payload); err != nil {
			return
		}

		// Decode
		var msg WFBMessage
		if err := msgpack.Unmarshal(payload, &msg); err != nil {
			log.Printf("Failed to decode stats msg: %v", err)
			continue
		}

		if msg.Type == "rx" {
			s.updateStats(msg)
		}
	}
}

func (s *WFBStatsService) updateStats(msg WFBMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newStats := &WFBStats{
		Rssi: make([]int8, 0, 4),
		Snr:  make([]int8, 0, 4),
	}

	// 1. Parse Rates
	getInt := func(key string) int {
		if val, ok := msg.Packets[key]; ok && len(val) > 0 {
			return int(val[0])
		}
		return 0
	}

	newStats.VideoPacketsPerSec = getInt("all")
	newStats.LostPacketsPerSec = getInt("lost")
	newStats.FecPacketsPerSec = getInt("fec_rec")
	newStats.BadBlocksPerSec = getInt("bad")
	newStats.LinkFlowBytesPerSec = getInt("all_bytes")

	// 2. Parse Session Info (FEC)
	if msg.Session != nil {
		if k, ok := msg.Session["fec_k"]; ok {
			newStats.FecK = int(convertToInt64(k))
		}
		if n, ok := msg.Session["fec_n"]; ok {
			newStats.FecN = int(convertToInt64(n))
		}
	}

	// 3. Parse Antenna Stats (Complex keys) & Transmission Info
	if len(msg.RxAntStats) > 0 {
		dec := msgpack.NewDecoder(bytes.NewReader(msg.RxAntStats))
		n, err := dec.DecodeMapLen()
		if err == nil {
			// Temporary map to sort by antenna ID
			type antData struct {
				rssi int8
				snr  int8
			}
			dataMap := make(map[int64]antData)
			foundRadioInfo := false

			for i := 0; i < n; i++ {
				// Decode Key: ((freq, mcs, bw), ant_id)
				var key []interface{}
				if err := dec.Decode(&key); err != nil {
					log.Printf("Error decoding ant key: %v", err)
					break
				}

				// Extract Radio Info from the first valid key
				// Key[0] is [freq, mcs, bw]
				if !foundRadioInfo && len(key) >= 1 {
					if radioInfo, ok := key[0].([]interface{}); ok && len(radioInfo) >= 3 {
						newStats.Frequency = uint32(convertToInt64(radioInfo[0]))
						newStats.McsIndex = int(convertToInt64(radioInfo[1]))
						newStats.Bandwidth = int(convertToInt64(radioInfo[2]))
						foundRadioInfo = true
					}
					// Check if msgpack decoded it as slice of correct types directly
					// Sometimes it might come as []int64 if generic decoding didn't happen?
					// But we are decoding into []interface{} which should work.
				}

				// Extract ant_id. Key should be [ [f,m,b], ant_id ]
				var antID int64
				if len(key) >= 2 {
					antID = convertToInt64(key[1])
				}

				// Decode Value: [pkt_s, rssi_min, rssi_avg, rssi_max, snr_min, snr_avg, snr_max]
				var val []int64
				if err := dec.Decode(&val); err != nil {
					log.Printf("Error decoding ant val: %v", err)
					break
				}

				if len(val) >= 6 {
					dataMap[antID] = antData{
						rssi: int8(val[2]),
						snr:  int8(val[5]),
					}
				}
			}

			// Sort by antenna ID
			var keys []int64
			for k := range dataMap {
				keys = append(keys, k)
			}
			// Bubble sort
			for i := 0; i < len(keys); i++ {
				for j := i + 1; j < len(keys); j++ {
					if keys[i] > keys[j] {
						keys[i], keys[j] = keys[j], keys[i]
					}
				}
			}

			for _, k := range keys {
				d := dataMap[k]
				newStats.Rssi = append(newStats.Rssi, d.rssi)
				newStats.Snr = append(newStats.Snr, d.snr)
			}
		} else {
			log.Printf("Error decoding map len: %v", err)
		}
	}

	s.currentStats = newStats
}

func convertToInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case uint64:
		return int64(val)
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case uint32:
		return int64(val)
	case int8:
		return int64(val)
	case uint8:
		return int64(val)
	case float64: // Msgpack sometimes decodes as float
		return int64(val)
	default:
		// Use reflect as last resort (e.g. strict types)
		rv := reflect.ValueOf(v)
		if rv.CanInt() {
			return rv.Int()
		}
		if rv.CanUint() {
			return int64(rv.Uint())
		}
		return 0
	}
}
