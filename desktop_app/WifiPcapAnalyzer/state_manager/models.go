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
	// Internal fields for metric calculation (not marshalled to JSON)
	lastCalcTime       time.Time
	totalAirtime       time.Duration // In a calculation window, contributed by this STA
	totalUplinkBytes   int64         // In a calculation window
	totalDownlinkBytes int64         // In a calculation window
}

// HT (High Throughput) Capabilities
type HTCapabilities struct {
	SupportedMCSSet   []int `json:"supported_mcs_set"`
	ShortGI20MHz      bool  `json:"short_gi_20mhz"`
	ShortGI40MHz      bool  `json:"short_gi_40mhz"`
	ChannelWidth40MHz bool  `json:"channel_width_40mhz"`
	// Add more fields as needed from HT Capabilities IE
}

// VHT (Very High Throughput) Capabilities
type VHTCapabilities struct {
	SupportedMCSSet         map[string][]int `json:"supported_mcs_set"` // e.g., "1x1": [0-9], "2x2": [0-9]
	ShortGI80MHz            bool             `json:"short_gi_80mhz"`
	ShortGI160MHz           bool             `json:"short_gi_160mhz"`
	ChannelWidth80MHz       bool             `json:"channel_width_80mhz"`
	ChannelWidth160MHz      bool             `json:"channel_width_160mhz"`
	ChannelWidth80Plus80MHz bool             `json:"channel_width_80plus80mhz"`
	// Add more fields as needed
}

// HE (High Efficiency) Capabilities
type HECapabilities struct {
	SupportedMCSSet map[string][]int `json:"supported_mcs_set"` // Similar to VHT
	// Add more HE specific fields
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
		MACAddress:   mac,
		LastSeen:     time.Now().UnixMilli(),
		lastCalcTime: time.Now(), // Initialize lastCalcTime
	}
}

// Snapshot represents a snapshot of all BSS and STA information
type Snapshot struct {
	BSSs []*BSSInfo `json:"bsss"`
	STAs []*STAInfo `json:"stas"`
}
