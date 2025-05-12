package state_manager

import "time"

// BSS (Basic Service Set) information
type BSSInfo struct {
	BSSID          string `json:"bssid"`
	SSID           string `json:"ssid"`
	Channel        int    `json:"channel"`
	Bandwidth      string `json:"bandwidth"`       // e.g., "20MHz", "40MHz", "80MHz"
	Security       string `json:"security"`        // e.g., "Open", "WPA2-PSK", "WPA3-SAE"
	SignalStrength int    `json:"signal_strength"` // dBm
	LastSeen       int64  `json:"last_seen"`       // Unix milliseconds
	// Capabilities (HT, VHT, HE, EHT) can be added as booleans or more detailed structs
	HTCapabilities  *HTCapabilities  `json:"ht_capabilities,omitempty"`
	VHTCapabilities *VHTCapabilities `json:"vht_capabilities,omitempty"`
	HECapabilities  *HECapabilities  `json:"he_capabilities,omitempty"`
	// EHTCapabilities *EHTCapabilities `json:"eht_capabilities,omitempty"` // If needed
	AssociatedSTAs map[string]*STAInfo `json:"associated_stas"` // Keyed by STA MAC

	// New metrics for channel utilization and throughput
	ChannelUtilization           float64   `json:"channel_utilization"`            // Current channel utilization percentage (0.0 - 100.0)
	Throughput                   int64     `json:"throughput"`                     // Current throughput in bps
	HistoricalChannelUtilization []float64 `json:"historical_channel_utilization"` // Historical channel utilization data
	HistoricalThroughput         []int64   `json:"historical_throughput"`          // Historical throughput data
	// Internal fields for metric calculation (not marshalled to JSON)
	lastCalcTime               time.Time
	totalAirtime               time.Duration // In a calculation window
	totalTxBytes               int64         // In a calculation window
	AccumulatedNavMicroseconds uint64        // Accumulated NAV duration in microseconds
	Util                       float64       `json:"util"`  // Current channel utilization percentage (0.0 - 100.0)
	Thrpt                      int64         `json:"thrpt"` // Current throughput in bps
}

// STA (Station) information
type STAInfo struct {
	MACAddress      string `json:"mac_address"`
	AssociatedBSSID string `json:"associated_bssid,omitempty"` // BSSID of the AP this STA is associated with
	SignalStrength  int    `json:"signal_strength"`            // dBm, from STA's perspective if available, or AP's perspective
	LastSeen        int64  `json:"last_seen"`                  // Unix milliseconds
	// Capabilities can also be added here if specific to STA
	HTCapabilities  *HTCapabilities  `json:"ht_capabilities,omitempty"`
	VHTCapabilities *VHTCapabilities `json:"vht_capabilities,omitempty"`
	HECapabilities  *HECapabilities  `json:"he_capabilities,omitempty"`

	// New metrics for channel utilization and throughput
	ChannelUtilization           float64   `json:"channel_utilization"`            // Current channel utilization percentage (0.0 - 100.0) by this STA
	UplinkThroughput             int64     `json:"uplink_throughput"`              // Current uplink throughput in bps
	DownlinkThroughput           int64     `json:"downlink_throughput"`            // Current downlink throughput in bps
	HistoricalChannelUtilization []float64 `json:"historical_channel_utilization"` // Historical channel utilization data for this STA
	HistoricalUplinkThroughput   []int64   `json:"historical_uplink_throughput"`   // Historical uplink throughput data for this STA
	HistoricalDownlinkThroughput []int64   `json:"historical_downlink_throughput"` // Historical downlink throughput data for this STA

	// 添加累积流量统计字段
	RxBytes   int64 `json:"rx_bytes"`   // 累积接收字节数
	TxBytes   int64 `json:"tx_bytes"`   // 累积发送字节数
	RxPackets int64 `json:"rx_packets"` // 累积接收包数
	TxPackets int64 `json:"tx_packets"` // 累积发送包数
	RxRetries int64 `json:"rx_retries"` // 累积接收重试次数
	TxRetries int64 `json:"tx_retries"` // 累积发送重试次数

	// Internal fields for metric calculation (not marshalled to JSON)
	lastCalcTime               time.Time
	totalAirtime               time.Duration // In a calculation window, contributed by this STA
	totalUplinkBytes           int64         // In a calculation window
	totalDownlinkBytes         int64         // In a calculation window
	AccumulatedNavMicroseconds uint64        // Accumulated NAV duration in microseconds (for NAV-based channel utilization)
	Util                       float64       `json:"util"`    // Current channel utilization percentage (0.0 - 100.0)
	Thrpt                      int64         `json:"thrpt"`   // Current throughput in bps
	BitRate                    float64       `json:"bitrate"` // Current BitRate in Mbps
}

// HT (High Throughput) Capabilities
type HTCapabilities struct {
	SupportedMCSSet   []int `json:"supported_mcs_set"`
	ShortGI20MHz      bool  `json:"short_gi_20mhz"`
	ShortGI40MHz      bool  `json:"short_gi_40mhz"`
	ChannelWidth40MHz bool  `json:"channel_width_40mhz"`
	// Additional fields to match the parser
	LDPCCoding         bool   `json:"ldpc_coding"`
	FortyMhzIntolerant bool   `json:"40mhz_intolerant"`
	TxSTBC             bool   `json:"tx_stbc"`
	RxSTBC             uint8  `json:"rx_stbc"`
	MaxAMSDULength     uint16 `json:"max_amsdu_length"`
	DSSCck             bool   `json:"dsss_cck_mode_40mhz"`
	HTDelayedBlockAck  bool   `json:"delayed_block_ack"`
	MaxAMPDULength     uint32 `json:"max_ampdu_length"`
	PrimaryChannel     uint8  `json:"primary_channel"`
}

// VHT (Very High Throughput) Capabilities
type VHTCapabilities struct {
	SupportedMCSSet         map[string][]int `json:"supported_mcs_set"` // e.g., "1x1": [0-9], "2x2": [0-9]
	ShortGI80MHz            bool             `json:"short_gi_80mhz"`
	ShortGI160MHz           bool             `json:"short_gi_160mhz"`
	ChannelWidth80MHz       bool             `json:"channel_width_80mhz"`
	ChannelWidth160MHz      bool             `json:"channel_width_160mhz"`
	ChannelWidth80Plus80MHz bool             `json:"channel_width_80plus80mhz"`
	// Additional fields to match the parser
	MaxMPDULength        uint8  `json:"max_mpdu_length"`
	RxLDPC               bool   `json:"rx_ldpc"`
	TxSTBC               bool   `json:"tx_stbc"`
	RxSTBC               uint8  `json:"rx_stbc"`
	SUBeamformerCapable  bool   `json:"su_beamformer_capable"`
	SUBeamformeeCapable  bool   `json:"su_beamformee_capable"`
	MUBeamformerCapable  bool   `json:"mu_beamformer_capable"`
	MUBeamformeeCapable  bool   `json:"mu_beamformee_capable"`
	BeamformeeSTS        uint8  `json:"beamformee_sts"`
	SoundingDimensions   uint8  `json:"sounding_dimensions"`
	MaxAMPDULengthExp    uint8  `json:"max_ampdu_length_exp"`
	RxPatternConsistency bool   `json:"rx_pattern_consistency"`
	TxPatternConsistency bool   `json:"tx_pattern_consistency"`
	RxMCSMap             uint16 `json:"rx_mcs_map"`
	TxMCSMap             uint16 `json:"tx_mcs_map"`
	RxHighestLongGIRate  uint16 `json:"rx_highest_long_gi_rate"`
	TxHighestLongGIRate  uint16 `json:"tx_highest_long_gi_rate"`
	VHTHTCCapability     bool   `json:"vht_htc_capability"`     // 添加: HTC功能
	VHTTXOPPSCapability  bool   `json:"vht_txop_ps_capability"` // 添加: TXOP省电功能
	ChannelCenter0       uint8  `json:"channel_center_0"`       // 添加: 主通道中心
	ChannelCenter1       uint8  `json:"channel_center_1"`       // 添加: 次通道中心
	// 带宽设置直接来自 SupportedChannelWidthSet 值
	SupportedChannelWidthSet uint8 `json:"supported_channel_width_set"`
}

// HE (High Efficiency) Capabilities
type HECapabilities struct {
	// 只保留与HECapabilityInfo中匹配的字段
	SupportedMCSSet map[string][]int `json:"supported_mcs_set"` // Similar to VHT
	BSSColor        string           `json:"bss_color"`

	// MAC能力字段
	HTCHESupport        bool `json:"htc_he_support"`
	TwtRequesterSupport bool `json:"twt_requester_support"`
	TwtResponderSupport bool `json:"twt_responder_support"`

	// PHY能力字段
	SUBeamformer bool `json:"su_beamformer"`
	SUBeamformee bool `json:"su_beamformee"`

	// 通道宽度相关字段
	ChannelWidth160MHz       bool `json:"channel_width_160mhz"`
	ChannelWidth80Plus80MHz  bool `json:"channel_width_80plus80mhz"`
	ChannelWidth40_80MHzIn5G bool `json:"channel_width_40_80mhz_in_5g"`

	// MCS相关字段
	MaxMCSForOneSS   uint8  `json:"max_mcs_for_1_ss"`
	MaxMCSForTwoSS   uint8  `json:"max_mcs_for_2_ss"`
	MaxMCSForThreeSS uint8  `json:"max_mcs_for_3_ss"`
	MaxMCSForFourSS  uint8  `json:"max_mcs_for_4_ss"`
	RxHEMCSMap       uint16 `json:"rx_he_mcs_map"`
	TxHEMCSMap       uint16 `json:"tx_he_mcs_map"`
}

// Helper function to create a new BSSInfo
func NewBSSInfo(bssid string) *BSSInfo {
	return &BSSInfo{
		BSSID:          bssid,
		AssociatedSTAs: make(map[string]*STAInfo),
		LastSeen:       time.Now().UnixMilli(),
		lastCalcTime:   time.Now(), // Initialize lastCalcTime
	}
}

// Helper function to create a new STAInfo
func NewSTAInfo(mac string) *STAInfo {
	return &STAInfo{
		MACAddress:                 mac,
		LastSeen:                   time.Now().UnixMilli(),
		lastCalcTime:               time.Now(), // Initialize lastCalcTime
		AccumulatedNavMicroseconds: 0,          // Initialize AccumulatedNavMicroseconds
		// 初始化累积统计字段
		RxBytes:   0,
		TxBytes:   0,
		RxPackets: 0,
		TxPackets: 0,
		RxRetries: 0,
		TxRetries: 0,
	}
}

// Snapshot represents a snapshot of all BSS and STA information
type Snapshot struct {
	BSSs []*BSSInfo `json:"bsss"`
	STAs []*STAInfo `json:"stas"`
}
