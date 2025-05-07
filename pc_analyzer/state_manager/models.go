package state_manager

import "time"

// BSS (Basic Service Set) information
type BSSInfo struct {
	BSSID          string    `json:"bssid"`
	SSID           string    `json:"ssid"`
	Channel        int       `json:"channel"`
	Bandwidth      string    `json:"bandwidth"`       // e.g., "20MHz", "40MHz", "80MHz"
	Security       string    `json:"security"`        // e.g., "Open", "WPA2-PSK", "WPA3-SAE"
	SignalStrength int       `json:"signal_strength"` // dBm
	LastSeen       time.Time `json:"last_seen"`
	// Capabilities (HT, VHT, HE, EHT) can be added as booleans or more detailed structs
	HTCapabilities  *HTCapabilities  `json:"ht_capabilities,omitempty"`
	VHTCapabilities *VHTCapabilities `json:"vht_capabilities,omitempty"`
	HECapabilities  *HECapabilities  `json:"he_capabilities,omitempty"`
	// EHTCapabilities *EHTCapabilities `json:"eht_capabilities,omitempty"` // If needed
	AssociatedSTAs map[string]*STAInfo `json:"associated_stas"` // Keyed by STA MAC
}

// STA (Station) information
type STAInfo struct {
	MACAddress      string    `json:"mac_address"`
	AssociatedBSSID string    `json:"associated_bssid,omitempty"` // BSSID of the AP this STA is associated with
	SignalStrength  int       `json:"signal_strength"`            // dBm, from STA's perspective if available, or AP's perspective
	LastSeen        time.Time `json:"last_seen"`
	// Capabilities can also be added here if specific to STA
	HTCapabilities  *HTCapabilities  `json:"ht_capabilities,omitempty"`
	VHTCapabilities *VHTCapabilities `json:"vht_capabilities,omitempty"`
	HECapabilities  *HECapabilities  `json:"he_capabilities,omitempty"`
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
		LastSeen:       time.Now(),
	}
}

// Helper function to create a new STAInfo
func NewSTAInfo(mac string) *STAInfo {
	return &STAInfo{
		MACAddress: mac,
		LastSeen:   time.Now(),
	}
}
