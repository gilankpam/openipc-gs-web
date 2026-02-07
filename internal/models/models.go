package models

// WFBConfig represents the structure of /etc/wfb.yaml
type WFBConfig struct {
	Wireless  WirelessConfig  `yaml:"wireless"`
	Broadcast BroadcastConfig `yaml:"broadcast"`
	Telemetry TelemetryConfig `yaml:"telemetry"`
}

type WirelessConfig struct {
	TxPower     int    `yaml:"txpower"` // Note: sample says txpower, not tx_power
	Channel     int    `yaml:"channel"`
	Width       int    `yaml:"width"`
	Mlink       int    `yaml:"mlink"`
	WlanAdapter string `yaml:"wlan_adapter"`
	LinkControl string `yaml:"link_control"`
}

type BroadcastConfig struct {
	McsIndex int `yaml:"mcs_index"`
	TunIndex int `yaml:"tun_index"`
	FecK     int `yaml:"fec_k"`
	FecN     int `yaml:"fec_n"`
	Stbc     int `yaml:"stbc"`
	Ldpc     int `yaml:"ldpc"`
	LinkId   int `yaml:"link_id"`
}

type TelemetryConfig struct {
	Router string `yaml:"router"`
	Serial string `yaml:"serial"`
	OsdFps int    `yaml:"osd_fps"`
}

// MajesticConfig represents the structure of /etc/majestic.yaml
type MajesticConfig struct {
	System       SystemConfig       `yaml:"system"`
	Isp          IspConfig          `yaml:"isp"`
	Image        ImageConfig        `yaml:"image"`
	Video0       VideoConfig        `yaml:"video0"`
	Video1       VideoConfig        `yaml:"video1"`
	Jpeg         JpegConfig         `yaml:"jpeg"`
	Osd          OsdConfig          `yaml:"osd"`
	Audio        AudioConfig        `yaml:"audio"`
	Rtsp         RtspConfig         `yaml:"rtsp"`
	NightMode    NightModeConfig    `yaml:"nightMode"`
	MotionDetect MotionDetectConfig `yaml:"motionDetect"`
	Records      RecordsConfig      `yaml:"records"`
	Outgoing     OutgoingConfig     `yaml:"outgoing"`
	Watchdog     WatchdogConfig     `yaml:"watchdog"`
	Hls          HlsConfig          `yaml:"hls"`
	Fpv          FpvConfig          `yaml:"fpv"`
}

type SystemConfig struct {
	WebPort   int    `yaml:"webPort"`
	HttpsPort int    `yaml:"httpsPort"`
	LogLevel  string `yaml:"logLevel"`
}

type IspConfig struct {
	AntiFlicker  string `yaml:"antiFlicker"`
	SensorConfig string `yaml:"sensorConfig"`
	Exposure     int    `yaml:"exposure"`
}

type ImageConfig struct {
	Mirror     bool `yaml:"mirror"`
	Flip       bool `yaml:"flip"`
	Rotate     int  `yaml:"rotate"`
	Contrast   int  `yaml:"contrast"`
	Hue        int  `yaml:"hue"`
	Saturation int  `yaml:"saturation"`
	Luminance  int  `yaml:"luminance"`
}

type VideoConfig struct {
	Enabled bool   `yaml:"enabled"`
	Codec   string `yaml:"codec"`
	Fps     int    `yaml:"fps"`
	Bitrate int    `yaml:"bitrate,omitempty"`
	RcMode  string `yaml:"rcMode,omitempty"`
	GopSize int    `yaml:"gopSize,omitempty"`
	Size    string `yaml:"size"`
}

type JpegConfig struct {
	Enabled bool `yaml:"enabled"`
	Qfactor int  `yaml:"qfactor"`
	Fps     int  `yaml:"fps"`
}

type OsdConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Font     string `yaml:"font"`
	Template string `yaml:"template"`
	PosX     int    `yaml:"posX"`
	PosY     int    `yaml:"posY"`
}

type AudioConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Volume       int    `yaml:"volume"`
	Srate        int    `yaml:"srate"`
	Codec        string `yaml:"codec"`
	OutputEnabled bool  `yaml:"outputEnabled"`
	OutputVolume int    `yaml:"outputVolume"`
}

type RtspConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type NightModeConfig struct {
	ColorToGray       bool `yaml:"colorToGray"`
	IrCutSingleInvert bool `yaml:"irCutSingleInvert"`
	LightMonitor      bool `yaml:"lightMonitor"`
	LightSensorInvert bool `yaml:"lightSensorInvert"`
}

type MotionDetectConfig struct {
	Enabled   bool `yaml:"enabled"`
	Visualize bool `yaml:"visualize"`
	Debug     bool `yaml:"debug"`
}

type RecordsConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Path     string `yaml:"path"`
	Split    int    `yaml:"split"`
	MaxUsage int    `yaml:"maxUsage"`
	NoTime   bool   `yaml:"notime"`
}

type OutgoingConfig struct {
	Enabled  bool `yaml:"enabled"`
	NaluSize int  `yaml:"naluSize"`
	Wfb      bool `yaml:"wfb"`
}

type WatchdogConfig struct {
	Enabled bool `yaml:"enabled"`
	Timeout int  `yaml:"timeout"`
}

type HlsConfig struct {
	Enabled bool `yaml:"enabled"`
}

type FpvConfig struct {
	Enabled    bool `yaml:"enabled"`
	NoiseLevel int  `yaml:"noiseLevel"`
}

// AlinkConfig represents the structure of /etc/alink.conf
type AlinkConfig struct {
	AllowSetPower            bool    `conf:"allow_set_power"`
	Use0To4TxPower           bool    `conf:"use_0_to_4_txpower"`
	PowerLevel0To4           int     `conf:"power_level_0_to_4"`
	GetCardInfoFromYaml      bool    `conf:"get_card_info_from_yaml"`
	RssiWeight               float64 `conf:"rssi_weight"`
	SnrWeight                float64 `conf:"snr_weight"`
	FallbackMs               int     `conf:"fallback_ms"`
	HoldFallbackModeS        int     `conf:"hold_fallback_mode_s"`
	MinBetweenChangesMs      int     `conf:"min_between_changes_ms"`
	HoldModesDownS           int     `conf:"hold_modes_down_s"`
	HysteresisPercent        int     `conf:"hysteresis_percent"`
	HysteresisPercentDown    int     `conf:"hysteresis_percent_down"`
	ExpSmoothingFactor       float64 `conf:"exp_smoothing_factor"`
	ExpSmoothingFactorDown   float64 `conf:"exp_smoothing_factor_down"`
	AllowRequestKeyframe     bool    `conf:"allow_request_keyframe"`
	AllowRqKfByTxD           bool    `conf:"allow_rq_kf_by_tx_d"`
	CheckXtxPeriodMs         int     `conf:"check_xtx_period_ms"`
	RequestKeyframeIntervalMs int    `conf:"request_keyframe_interval_ms"`
	IdrEveryChange           bool    `conf:"idr_every_change"`
	RoiFocusMode             int     `conf:"roi_focus_mode"`
	AllowDynamicFec          bool    `conf:"allow_dynamic_fec"`
	FecKAdjust               int     `conf:"fec_k_adjust"`
	SpikeFixDynamicFec       bool    `conf:"spike_fix_dynamic_fec"`
	AllowSpikeFixFps         bool    `conf:"allow_spike_fix_fps"`
	AllowXtxReduceBitrate    bool    `conf:"allow_xtx_reduce_bitrate"`
	XtxReduceBitrateFactor   float64 `conf:"xtx_reduce_bitrate_factor"`
	OsdLevel                 int     `conf:"osd_level"`
	MultiplyFontSizeBy       int     `conf:"multiply_font_size_by"`
	
	// Command templates (strings)
	PowerCommandTemplate     string `conf:"powerCommandTemplate"`
	FpsCommandTemplate       string `conf:"fpsCommandTemplate"`
	QpDeltaCommandTemplate   string `conf:"qpDeltaCommandTemplate"`
	McsCommandTemplate       string `conf:"mcsCommandTemplate"`
	BitrateCommandTemplate   string `conf:"bitrateCommandTemplate"`
	GopCommandTemplate       string `conf:"gopCommandTemplate"`
	FecCommandTemplate       string `conf:"fecCommandTemplate"`
	RoiCommandTemplate       string `conf:"roiCommandTemplate"`
	IdrCommandTemplate       string `conf:"idrCommandTemplate"`
	CustomOSD                string `conf:"customOSD"`
}

// API Request/Response Models

type RadioSettings struct {
	Channel   *int `json:"channel"`
	Bandwidth *int `json:"bandwidth"`
	TxPower   *int `json:"tx_power"`
	McsIndex  *int `json:"mcs_index"`
	FecK      *int `json:"fec_k"`
	FecN      *int `json:"fec_n"`
}

type VideoSettings struct {
	Resolution *string `json:"resolution"`
	Fps        *int    `json:"fps"`
	Codec      *string `json:"codec"`
	Bitrate    *int    `json:"bitrate"`
	GopSize    *int    `json:"gop_size"`
}

type CameraSettings struct {
	Exposure   *string `json:"exposure"` // simplified
	Contrast   *int    `json:"contrast"`
	Saturation *int    `json:"saturation"`
	Flip       *bool   `json:"flip"`
	Mirror     *bool   `json:"mirror"`
	Rotate     *int    `json:"rotate"`
}

type TelemetrySettings struct {
	SerialPort *string `json:"serial_port"`
	Router     *string `json:"router"`
	BaudRate   *int    `json:"baud_rate"`
}

type AdaptiveLinkSettings struct {
	Enabled          *bool `json:"enabled"`
	AllowSetPower    *bool `json:"allow_set_power"`
	Use0To4TxPower   *bool `json:"use_0_to_4_txpower"`
	PowerLevel0To4   *int  `json:"power_level_0_to_4"`
	AllowSpikeFixFps *bool `json:"allow_spike_fix_fps"`
	OsdLevel         *int  `json:"osd_level"`
}
