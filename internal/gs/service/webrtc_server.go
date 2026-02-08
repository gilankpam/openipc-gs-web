package service

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

// StreamServer handles WebRTC streaming of RTP H265 video
type StreamServer struct {
	rtpPort    int
	cmd        *exec.Cmd
	peers      map[string]*webrtc.PeerConnection
	peersMu    sync.RWMutex
	videoTrack *webrtc.TrackLocalStaticSample
	running    bool
	stopCh     chan struct{}
}

// NewStreamServer creates a new streaming server
func NewStreamServer(rtpPort int) *StreamServer {
	return &StreamServer{
		rtpPort: rtpPort,
		peers:   make(map[string]*webrtc.PeerConnection),
		stopCh:  make(chan struct{}),
	}
}

// Start begins listening for RTP packets via GStreamer and serving WebRTC
func (s *StreamServer) Start() error {
	// Create a video track for H265 samples
	var err error
	s.videoTrack, err = webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH265},
		"video",
		"wfb-stream",
	)
	if err != nil {
		return err
	}

	s.running = true
	log.Printf("Streaming server starting GStreamer pipeline on UDP port %d", s.rtpPort)

	// Start GStreamer reader goroutine
	go s.readFromGStreamer()

	return nil
}

// Stop shuts down the streaming server
func (s *StreamServer) Stop() {
	if !s.running {
		return
	}
	s.running = false
	close(s.stopCh)

	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}

	// Close all peer connections
	s.peersMu.Lock()
	for id, pc := range s.peers {
		pc.Close()
		delete(s.peers, id)
	}
	s.peersMu.Unlock()
}

// readFromGStreamer spawns GStreamer to depayload RTP and reads H265 byte-stream
func (s *StreamServer) readFromGStreamer() {
	for s.running {
		// GStreamer pipeline: receive RTP H265, depayload, output byte-stream to stdout
		s.cmd = exec.Command("gst-launch-1.0", "-q",
			"udpsrc", "port=5601",
			"caps=application/x-rtp, media=(string)video, encoding-name=(string)H265, clock-rate=(int)90000",
			"!", "rtph265depay",
			"!", "h265parse", "config-interval=-1",
			"!", "video/x-h265, stream-format=byte-stream",
			"!", "fdsink", "fd=1",
		)

		stdout, err := s.cmd.StdoutPipe()
		if err != nil {
			log.Printf("Failed to get GStreamer stdout: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Capture stderr for debugging
		stderr, _ := s.cmd.StderrPipe()
		go func() {
			buf := make([]byte, 4096)
			for {
				n, err := stderr.Read(buf)
				if n > 0 {
					log.Printf("GStreamer stderr: %s", string(buf[:n]))
				}
				if err != nil {
					return
				}
			}
		}()

		if err := s.cmd.Start(); err != nil {
			log.Printf("Failed to start GStreamer: %v", err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("GStreamer pipeline started (PID: %d)", s.cmd.Process.Pid)

		// Read H265 NAL units from stdout
		buf := make([]byte, 65535)
		for s.running {
			n, err := stdout.Read(buf)
			if err != nil {
				if err != io.EOF && s.running {
					log.Printf("GStreamer read error: %v", err)
				}
				break
			}

			if n > 0 {
				// Write H265 data as a sample
				sample := media.Sample{
					Data:     buf[:n],
					Duration: time.Millisecond * 33, // ~30fps
				}
				if err := s.videoTrack.WriteSample(sample); err != nil {
					log.Printf("Sample write error: %v", err)
				}
			}
		}

		s.cmd.Wait()
		if s.running {
			log.Printf("GStreamer pipeline ended, restarting...")
			time.Sleep(time.Second)
		}
	}
}

// SignalingRequest represents a WebRTC offer from the browser
type SignalingRequest struct {
	Offer webrtc.SessionDescription `json:"offer"`
}

// SignalingResponse represents the WebRTC answer to send to the browser
type SignalingResponse struct {
	Answer webrtc.SessionDescription `json:"answer"`
}

// HandleSignaling handles WebRTC signaling HTTP requests
func (s *StreamServer) HandleSignaling(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the offer
	var req SignalingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create a new peer connection
	peerConnection, err := s.createPeerConnection()
	if err != nil {
		log.Printf("Failed to create peer connection: %v", err)
		http.Error(w, "Failed to create connection", http.StatusInternalServerError)
		return
	}

	// Set the remote description (the browser's offer)
	if err = peerConnection.SetRemoteDescription(req.Offer); err != nil {
		log.Printf("Failed to set remote description: %v", err)
		peerConnection.Close()
		http.Error(w, "Failed to process offer", http.StatusInternalServerError)
		return
	}

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Printf("Failed to create answer: %v", err)
		peerConnection.Close()
		http.Error(w, "Failed to create answer", http.StatusInternalServerError)
		return
	}

	// Create channel to wait for ICE gathering completion
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Set the local description
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		log.Printf("Failed to set local description: %v", err)
		peerConnection.Close()
		http.Error(w, "Failed to set local description", http.StatusInternalServerError)
		return
	}

	// Wait for ICE gathering to complete
	<-gatherComplete

	// Send the answer back
	resp := SignalingResponse{
		Answer: *peerConnection.LocalDescription(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// createPeerConnection creates and configures a new WebRTC peer connection
func (s *StreamServer) createPeerConnection() (*webrtc.PeerConnection, error) {
	// Configure WebRTC
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	// Add the video track
	rtpSender, err := peerConnection.AddTrack(s.videoTrack)
	if err != nil {
		peerConnection.Close()
		return nil, err
	}

	// Read incoming RTCP packets (required for NACK processing)
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// Handle connection state changes
	peerID := peerConnection.GetConfiguration().PeerIdentity
	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("Peer %s connection state: %s", peerID, state.String())

		if state == webrtc.ICEConnectionStateFailed ||
			state == webrtc.ICEConnectionStateClosed ||
			state == webrtc.ICEConnectionStateDisconnected {
			peerConnection.Close()
			s.peersMu.Lock()
			delete(s.peers, peerID)
			s.peersMu.Unlock()
		}
	})

	// Store peer connection
	s.peersMu.Lock()
	s.peers[peerID] = peerConnection
	s.peersMu.Unlock()

	return peerConnection, nil
}
