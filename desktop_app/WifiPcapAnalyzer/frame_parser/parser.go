package frame_parser

import (
	"WifiPcapAnalyzer/logger"
	"WifiPcapAnalyzer/utils"
	"strings"

	// "encoding/csv" // No longer needed after CSVParser removal
	// "encoding/hex" // No longer needed
	"fmt"
	// "io" // No longer needed
	"net"
	// "os/exec" // No longer needed after TSharkExecutor removal
	// "strconv" // No longer needed
	// "strings" // No longer needed
	"time"
	"unicode/utf8"

	"encoding/binary"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// GoPacketParser will use gopacket to parse 802.11 frames.
type GoPacketParser struct {
	// Potentially add fieldsNil here if needed later, e.g., options or a reusable packet source.
}

// HTCapabilityInfo stores parsed HT capabilities.
type HTCapabilityInfo struct {
	ChannelWidth40MHz      bool   `json:"channel_width_40mhz"`
	ShortGI20MHz           bool   `json:"short_gi_20mhz"`
	ShortGI40MHz           bool   `json:"short_gi_40mhz"`
	SupportedMCSSet        []byte `json:"supported_mcs_set"` // Raw 16 bytes
	PrimaryChannel         uint8  `json:"primary_channel"`
	SecondaryChannelOffset string `json:"secondary_channel_offset"`
	OperatingAt40MHz       bool   `json:"operating_at_40mhz"`
	// Additional fields from tshark
	LDPCCoding         bool   `json:"ldpc_coding"`
	FortyMhzIntolerant bool   `json:"40mhz_intolerant"`
	TxSTBC             bool   `json:"tx_stbc"`
	RxSTBC             uint8  `json:"rx_stbc"`
	MaxAMSDULength     uint16 `json:"max_amsdu_length"`
	DSSCck             bool   `json:"dsss_cck_mode_40mhz"`
	HTDelayedBlockAck  bool   `json:"delayed_block_ack"`
	MaxAMPDULength     uint32 `json:"max_ampdu_length"`
}

// VHTCapabilityInfo stores parsed VHT capabilities.
type VHTCapabilityInfo struct {
	MaxMPDULength            uint8  `json:"max_mpdu_length"`
	SupportedChannelWidthSet uint8  `json:"supported_channel_width_set"`
	ShortGI80MHz             bool   `json:"short_gi_80mhz"`
	ShortGI160MHz            bool   `json:"short_gi_160mhz"`
	SUBeamformerCapable      bool   `json:"su_beamformer_capable"`
	MUBeamformerCapable      bool   `json:"mu_beamformer_capable"`
	RxMCSMap                 uint16 `json:"rx_mcs_map"`
	RxHighestLongGIRate      uint16 `json:"rx_highest_long_gi_rate"`
	TxMCSMap                 uint16 `json:"tx_mcs_map"`
	TxHighestLongGIRate      uint16 `json:"tx_highest_long_gi_rate"`
	ChannelWidth             string `json:"channel_width"` // e.g., "20", "40", "80", "160", "80+80"
	ChannelCenter0           uint8  `json:"channel_center_0"`
	ChannelCenter1           uint8  `json:"channel_center_1"` // Added: channel_center_1
	// Additional fields from tshark
	RxLDPC               bool  `json:"rx_ldpc"`
	TxSTBC               bool  `json:"tx_stbc"`
	RxSTBC               uint8 `json:"rx_stbc"`
	SUBeamformee         bool  `json:"su_beamformee"`
	MUBeamformee         bool  `json:"mu_beamformee"`
	BeamformeeSTS        uint8 `json:"beamformee_sts"`
	SoundingDimensions   uint8 `json:"sounding_dimensions"`
	MaxAMPDULengthExp    uint8 `json:"max_ampdu_length_exp"`
	RxPatternConsistency bool  `json:"rx_pattern_consistency"`
	TxPatternConsistency bool  `json:"tx_pattern_consistency"`
	VHTHTCCapability     bool  `json:"vht_htc_capability"`     // 修正：添加vht_htc_capability字段
	VHTTXOPPSCapability  bool  `json:"vht_txop_ps_capability"` // 添加：txop省电能力
}

// HECapabilityInfo stores parsed HE capabilities.
type HECapabilityInfo struct {
	// 只保留JSON示例中确认存在的字段
	BSSColor string `json:"bss_color"`
	// MAC能力字段
	HTCHESupport        bool `json:"htc_he_support"`
	TwtRequesterSupport bool `json:"twt_requester_support"`
	TwtResponderSupport bool `json:"twt_responder_support"`

	// PHY能力字段
	SUBeamformer bool `json:"su_beamformer"`
	SUBeamformee bool `json:"su_beamformee"`

	// 支持的通道宽度能力 - 从示例JSON确认存在的字段
	ChannelWidth160MHz       bool `json:"channel_width_160mhz"`         // 对应 wlan.ext_tag.he_phy_cap.chan_width_set.160_in_5ghz
	ChannelWidth80Plus80MHz  bool `json:"channel_width_80plus80mhz"`    // 对应 wlan.ext_tag.he_phy_cap.chan_width_set.160_80_80_in_5ghz
	ChannelWidth40_80MHzIn5G bool `json:"channel_width_40_80mhz_in_5g"` // 对应 wlan.ext_tag.he_phy_cap.chan_width_set.40_80_in_5ghz

	// MCS相关字段
	MaxMCSForOneSS   uint8  `json:"max_mcs_for_1_ss"`
	MaxMCSForTwoSS   uint8  `json:"max_mcs_for_2_ss"`
	MaxMCSForThreeSS uint8  `json:"max_mcs_for_3_ss"`
	MaxMCSForFourSS  uint8  `json:"max_mcs_for_4_ss"`
	RxHEMCSMap       uint16 `json:"rx_he_mcs_map"`
	TxHEMCSMap       uint16 `json:"tx_he_mcs_map"`
}

// ParsedFrameInfo holds extracted information from a single 802.11 frame.
type ParsedFrameInfo struct {
	Timestamp              time.Time
	FrameType              string // e.g., "Beacon", "ProbeResp", "Data", "QoSData" (derived from wlan.fc.type_subtype)
	WlanFcType             uint8  // WLAN Frame Type (integer)
	WlanFcSubtype          uint8  // WLAN Frame Subtype (integer)
	BSSID                  net.HardwareAddr
	SA                     net.HardwareAddr
	DA                     net.HardwareAddr
	RA                     net.HardwareAddr
	TA                     net.HardwareAddr
	Channel                int      // Derived from radiotap.channel.freq or wlan.ds.current_channel
	Frequency              int      // radiotap.channel.freq
	SignalStrength         int      // radiotap.dbm_antsignal
	NoiseLevel             int      // radiotap.dbm_antnoise
	Bandwidth              string   // Derived from HT/VHT/HE capabilities
	SSID                   string   // wlan.ssid
	SupportedRates         []string // From relevant IEs, if parsed
	DSSetChannel           uint8    // wlan.ds.current_channel
	TIM                    []byte   // wlan.tim (raw bytes or parsed structure)
	RSNRaw                 []byte   // wlan.rsn.* (raw bytes or parsed structure)
	Security               string   // Security information
	IsQoSData              bool     // Derived from frame type/subtype
	ParsedHTCaps           *HTCapabilityInfo
	ParsedVHTCaps          *VHTCapabilityInfo
	ParsedHECaps           *HECapabilityInfo // New
	FrameLength            int               // frame.len (original frame length)
	FrameCapLength         int               // frame.cap_len (captured frame length)
	PHYRateMbps            float64           // Estimated PHY rate in Mbps
	IsShortPreamble        bool              // Potentially from radiotap flags (if available) or inferred
	IsShortGI              bool              // From Radiotap MCS/HT/VHT/HE flags or capabilities
	TransportPayloadLength int               // L4+ payload length (ip.len, ipv6.plen, tcp.len, udp.length)
	MACDurationID          uint16            // wlan.duration
	RetryFlag              bool              // wlan.flags.retry
	// Fields from radiotap.mcs.*, radiotap.vht.*, radiotap.he.* for PhyRateCalculator
	RadiotapDataRate   float64 // radiotap.datarate (legacy)
	RadiotapMCSIndex   uint8   // radiotap.mcs.index
	RadiotapMCSBw      uint8   // radiotap.mcs.bw (20, 40)
	RadiotapMCSGI      bool    // radiotap.mcs.gi (short GI)
	RadiotapVHTMCS     uint8   // radiotap.vht.mcs
	RadiotapVHTNSS     uint8   // radiotap.vht.nss
	RadiotapVHTBw      string  // radiotap.vht.bw (e.g., "20", "40", "80", "160", "80+80")
	RadiotapVHTShortGI bool    // radiotap.vht.gi
	RadiotapHEMCS      uint8   // radiotap.he.mcs
	RadiotapHENSS      uint8   // radiotap.he.nss
	RadiotapHEBw       string  // radiotap.he.bw (e.g., "20MHz", "40MHz", "80MHz", "HE_MU_80MHz")
	RadiotapHEGI       string  // radiotap.he.gi (e.g., "0.8us", "1.6us", "3.2us")
	BitRate            float64 // STA BitRate

	// Raw tshark fields for debugging or further processing if needed
	// This field might be removed or re-purposed if not used by gopacket direct parsing.
	RawFields map[string]string
}

// PacketInfoHandler is a function that processes parsed frame information.
type PacketInfoHandler func(info *ParsedFrameInfo)

// ProcessPacketSource is the main entry point for parsing pcap data
// from a gopacket.PacketDataSource.
func ProcessPacketSource(packetSource *gopacket.PacketSource, pktHandler PacketInfoHandler) error {
	parser := &GoPacketParser{}
	frameCount := 0
	errorCount := 0

	logger.Log.Info().Msg("INFO_PCAP_PROCESS: Starting packet processing from gopacket.PacketSource")

	for packet := range packetSource.Packets() {
		if packet == nil {
			logger.Log.Warn().Msg("WARN_PCAP_PROCESS: Nil packet received from source, stopping.")
			break // End of stream or error
		}
		frameCount++

		// Log basic packet metadata
		// logger.Log.Debug().
		// 	Int("frameNum", frameCount).
		// 	Time("timestamp", packet.Metadata().Timestamp).
		// 	Int("length", packet.Metadata().Length).
		// 	Msg("Processing packet")

		parsedInfo, err := parser.ParsePacket(packet)
		if err != nil {
			errorCount++
			// Log more detailed error, including packet dump if small enough or relevant parts
			// logger.Log.Warn().Err(err).Int("frameNum", frameCount).Msg("Error parsing packet")

			// Consider logging a snippet of the packet data for debugging difficult cases.
			// Example: logger.Log.Debug().Str("packet_data_snippet", hex.EncodeToString(packet.Data()[:min(32, len(packet.Data()))])).Msg("Packet data snippet on error")

			// Continue processing other packets
			continue
		}

		if parsedInfo != nil {
			// The RawFields map is not populated by GoPacketParser, initialize if nil to prevent panic
			if parsedInfo.RawFields == nil {
				parsedInfo.RawFields = make(map[string]string)
			}
			pktHandler(parsedInfo)
		}
	}

	logger.Log.Info().
		Int("totalFrames", frameCount).
		Int("errorCount", errorCount).
		Msg("INFO_PCAP_PROCESS: Finished processing packets from gopacket.PacketSource")

	if errorCount > 0 {
		return fmt.Errorf("encountered %d errors during packet parsing", errorCount)
	}
	return nil
}

// ProcessPcapFile processes a pcap file using gopacket.
func ProcessPcapFile(pcapFilePath string, _ string /* tsharkPath (unused) */, pktHandler PacketInfoHandler) error {
	logger.Log.Info().Str("filePath", pcapFilePath).Msg("INFO_PCAP_PROCESS: Opening pcap file for gopacket processing")
	handle, err := pcap.OpenOffline(pcapFilePath)
	if err != nil {
		logger.Log.Error().Err(err).Str("filePath", pcapFilePath).Msg("Error opening pcap file with gopacket")
		return fmt.Errorf("gopacket.OpenOffline failed: %w", err)
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	return ProcessPacketSource(packetSource, pktHandler)
}

// ProcessPcapStream processes a pcap stream using gopacket.
// The packetSource is now expected to be created by the caller (e.g., from pcapgo.NewReader).
func ProcessPcapStream(packetSource *gopacket.PacketSource, _ string /* tsharkPath (unused) */, pktHandler PacketInfoHandler) error {
	logger.Log.Info().Msg("INFO_PCAP_PROCESS: Starting pcap stream processing with gopacket")
	return ProcessPacketSource(packetSource, pktHandler)
}

// getPHYRateMbps estimates the PHY rate. This function is a placeholder and needs actual implementation
// based on radiotap fields if they are reliably available, or from parsed HT/VHT/HE capabilities.
// Note: This function is currently NOT CALLED by GoPacketParser.ParsePacket directly.
// The rate calculation is embedded within ParsePacket.
func getPHYRateMbps(info *ParsedFrameInfo) float64 {
	// This is a very basic placeholder.
	// A more accurate calculation would involve looking at Radiotap MCS/VHT/HE fields,
	// or HT/VHT/HE capabilities elements.
	// For now, if RadiotapDataRate is present, use it (it's often for legacy rates).
	if info.RadiotapDataRate > 0 {
		return info.RadiotapDataRate
	}
	// Fallback or more complex logic needed here.
	return 0.0 // Default if no rate info found
}

// CalculateFrameAirtime estimates the airtime for a given frame.
// This is a simplified version. Real airtime calculation is complex.
// NOTE: This function is currently NOT USED. Airtime calculation was removed from StateManager.
func CalculateFrameAirtime(frameLengthBytes int, phyRateMbps float64, isShortPreamble bool, isShortGI bool) time.Duration {
	if phyRateMbps <= 0 {
		return 0 // Avoid division by zero or negative rates
	}

	// Basic data transmission time
	// bits = bytes * 8
	// time_seconds = bits / (phyRateMbps * 1,000,000)
	dataTxTimeMicroseconds := float64(frameLengthBytes*8) / phyRateMbps

	// Simplified preamble and overhead considerations (these are very rough estimates)
	// Actual overhead depends on PHY type (802.11a/g/n/ac/ax), preambles, SIFS, DIFS, backoff, etc.
	var overheadMicroseconds float64 = 20 // Generic overhead for SIFS + ACK etc. (very rough)

	if isShortPreamble {
		overheadMicroseconds -= 5 // Arbitrary reduction for short preamble
	}
	if isShortGI {
		overheadMicroseconds -= 4 // Arbitrary reduction for short GI (especially at higher rates)
	}

	totalAirtimeMicroseconds := dataTxTimeMicroseconds + overheadMicroseconds
	if totalAirtimeMicroseconds < 0 {
		totalAirtimeMicroseconds = 0
	}

	return time.Duration(totalAirtimeMicroseconds * float64(time.Microsecond))
}

// ParsePacket uses gopacket to parse an 802.11 frame and extract information.
func (p *GoPacketParser) ParsePacket(packet gopacket.Packet) (*ParsedFrameInfo, error) {
	info := &ParsedFrameInfo{
		Timestamp:      packet.Metadata().Timestamp,
		FrameLength:    packet.Metadata().Length,
		FrameCapLength: packet.Metadata().CaptureLength,
		RawFields:      make(map[string]string),
	}

	if radiotapLayer := packet.Layer(layers.LayerTypeRadioTap); radiotapLayer != nil {
		rt, ok := radiotapLayer.(*layers.RadioTap)
		if !ok {
			return nil, fmt.Errorf("failed to assert RadioTap layer")
		}
		if rt.Present.DBMAntennaSignal() {
			info.SignalStrength = int(rt.DBMAntennaSignal)
		}
		if rt.Present.DBMAntennaNoise() {
			info.NoiseLevel = int(rt.DBMAntennaNoise)
		}
		if rt.Present.Channel() {
			info.Frequency = int(rt.ChannelFrequency)
			info.Channel = utils.FrequencyToChannel(info.Frequency)
		}
		if rt.Present.Rate() {
			info.RadiotapDataRate = float64(rt.Rate) * 0.5
		}
		if rt.Present.MCS() {
			info.RadiotapMCSIndex = rt.MCS.MCS
			flags := rt.MCS.Flags // This is layers.RadioTapMCSFlags
			if flags.ShortGI() {
				info.IsShortGI = true
			}

			// Populate RadiotapMCSBw based on MCS flags
			// info.RadiotapMCSBw: 0 for 20MHz, 1 for 40MHz
			// rt.MCS.Flags.Bandwidth() returns (flags & 0x03) which corresponds to:
			// 0 (00_bin): 20MHz (layers.RadioTapMCSBandwidth20)
			// 1 (01_bin): 40MHz (layers.RadioTapMCSBandwidth40)
			// 2 (10_bin): 20L MHz (VHT) (layers.RadioTapMCSBandwidth20L)
			// 3 (11_bin): 20U MHz (VHT) (layers.RadioTapMCSBandwidth20U)
			bwRawValue := flags.Bandwidth() // Type is layers.RadioTapMCSFlags, value is 0,1,2 or 3

			if bwRawValue == 1 { // Value for 40MHz
				info.RadiotapMCSBw = 1 // 40MHz
			} else { // Covers 0 (20MHz), 2 (20L), 3 (20U)
				info.RadiotapMCSBw = 0 // Defaulting to 20MHz for these cases, aligning with info.RadiotapMCSBw structure
			}
		}
		if rt.Present.VHT() {
			// VHT parsing
		}
	} else {
		logger.Log.Warn().Msg("No Radiotap layer found in packet")
	}

	dot11Layer := packet.Layer(layers.LayerTypeDot11)
	if dot11Layer == nil {
		return nil, fmt.Errorf("no Dot11 layer found")
	}
	dot11, ok := dot11Layer.(*layers.Dot11)
	if !ok {
		return nil, fmt.Errorf("failed to assert Dot11 layer")
	}

	mainType := dot11.Type.MainType()
	info.WlanFcType = uint8(mainType)
	info.WlanFcSubtype = uint8(dot11.Type)
	info.FrameType = dot11.Type.String()

	info.RetryFlag = dot11.Flags.Retry()
	info.MACDurationID = dot11.DurationID

	info.DA = dot11.Address1
	info.SA = dot11.Address2
	info.BSSID = dot11.Address3

	toDS := dot11.Flags.ToDS()
	fromDS := dot11.Flags.FromDS()

	switch {
	case !toDS && !fromDS:
		info.RA = dot11.Address1
		info.TA = dot11.Address2
	case !toDS && fromDS:
		info.RA = dot11.Address1
		info.TA = dot11.Address2
		info.SA = dot11.Address3
		info.BSSID = dot11.Address2
	case toDS && !fromDS:
		info.RA = dot11.Address3
		info.TA = dot11.Address2
		info.DA = dot11.Address1
		info.BSSID = dot11.Address1
	case toDS && fromDS:
		info.RA = dot11.Address1
		info.TA = dot11.Address2
		if dot11.Address4 != nil {
			info.DA = dot11.Address3
			info.SA = dot11.Address4
		} else {
			info.DA = dot11.Address3
		}
	}

	if dot11.Type.QOS() {
		info.IsQoSData = true
	}

	// --- IE Parsing for Management Frames (Manual from Payload) ---
	if dot11.Type.MainType() == layers.Dot11TypeMgmt {
		var iePayload []byte

		switch dot11.Type {
		case layers.Dot11TypeMgmtBeacon:
			if beaconLayer := packet.Layer(layers.LayerTypeDot11MgmtBeacon); beaconLayer != nil {
				if beacon, ok := beaconLayer.(*layers.Dot11MgmtBeacon); ok {
					iePayload = beacon.Payload
				} else {
					logger.Log.Debug().Msg("Failed to assert Dot11MgmtBeacon layer after finding it.")
				}
			} else {
				logger.Log.Debug().Msg("Dot11MgmtBeacon layer not found by direct type request.")
			}
		case layers.Dot11TypeMgmtProbeResp:
			if probeRespLayer := packet.Layer(layers.LayerTypeDot11MgmtProbeResp); probeRespLayer != nil {
				if probeResp, ok := probeRespLayer.(*layers.Dot11MgmtProbeResp); ok {
					iePayload = probeResp.Payload
				} else {
					logger.Log.Debug().Msg("Failed to assert Dot11MgmtProbeResp layer after finding it.")
				}
			} else {
				logger.Log.Debug().Msg("Dot11MgmtProbeResp layer not found by direct type request.")
			}
		case layers.Dot11TypeMgmtProbeReq:
			if probeReqLayer := packet.Layer(layers.LayerTypeDot11MgmtProbeReq); probeReqLayer != nil {
				if probeReq, ok := probeReqLayer.(*layers.Dot11MgmtProbeReq); ok {
					iePayload = probeReq.Payload
				} else {
					logger.Log.Debug().Msg("Failed to assert Dot11MgmtProbeReq layer after finding it.")
				}
			} else {
				logger.Log.Debug().Msg("Dot11MgmtProbeReq layer not found by direct type request.")
			}
		default:
			logger.Log.Debug().Stringer("mgmt_frame_type", dot11.Type).Msg("SSID parsing not specifically handled for this management frame subtype via specific layer.")
		}

		if iePayload != nil {
			currentIndex := 0
			for currentIndex < len(iePayload) {
				if currentIndex+2 > len(iePayload) { // Need at least ID and Length fields
					logger.Log.Warn().Int("offset", currentIndex).Int("payload_len", len(iePayload)).Msg("IE parsing stopped: not enough data for ID/Length.")
					break
				}
				ieID := layers.Dot11InformationElementID(iePayload[currentIndex])
				ieLength := int(iePayload[currentIndex+1])

				if currentIndex+2+ieLength > len(iePayload) { // Check if data for this IE is fully present
					logger.Log.Warn().Stringer("ie_id", ieID).Int("declared_len", ieLength).Int("remaining_payload", len(iePayload)-currentIndex-2).Msg("IE parsing stopped: declared length exceeds available payload.")
					break
				}
				ieData := iePayload[currentIndex+2 : currentIndex+2+ieLength]

				switch ieID {
				case layers.Dot11InformationElementIDSSID:
					if len(ieData) == 0 {
						info.SSID = "<empty>"
					} else {
						isHidden := true
						for _, b := range ieData {
							if b != 0 {
								isHidden = false
								break
							}
						}
						if isHidden {
							info.SSID = "<hidden>"
						} else {
							if utf8.Valid(ieData) {
								info.SSID = string(ieData)
							} else {
								hexSSID := fmt.Sprintf("%x", ieData)
								logger.Log.Warn().Str("bssid", info.BSSID.String()).Str("ssid_hex", hexSSID).Msg("Non-UTF-8 SSID encountered, displaying as hex.")
								info.SSID = fmt.Sprintf("<HEX:%s>", hexSSID)
							}
						}
					}
				case layers.Dot11InformationElementIDDSSet:
					if len(ieData) == 1 {
						info.DSSetChannel = ieData[0]
						if info.Channel == 0 && info.DSSetChannel > 0 {
							info.Channel = int(info.DSSetChannel)
						}
					}
				case layers.Dot11InformationElementIDTIM:
					info.TIM = make([]byte, len(ieData))
					copy(info.TIM, ieData)

				// Simplified stubs for capability IEs - assuming functions exist later
				case layers.Dot11InformationElementIDHTCapabilities:
					if info.ParsedHTCaps == nil {
						info.ParsedHTCaps = &HTCapabilityInfo{}
					} // Ensure not nil
					parseHTCapabilitiesIE(info, ieData)
				case layers.Dot11InformationElementIDRSNInfo:
					// info.RSNRaw is currently just storing the raw IE.
					// We will now call the dedicated parser instead/in addition.
					parseRSNIE(info, ieData)
					// Keep storing raw RSN for now, might be useful for debugging or if parsing fails
					info.RSNRaw = make([]byte, 2+len(ieData))
					info.RSNRaw[0] = byte(ieID)
					info.RSNRaw[1] = byte(ieLength)
					copy(info.RSNRaw[2:], ieData)
				case layers.Dot11InformationElementIDVHTCapabilities:
					if info.ParsedVHTCaps == nil {
						info.ParsedVHTCaps = &VHTCapabilityInfo{}
					}
					parseVHTCapabilitiesIE(info, ieData)
				case layers.Dot11InformationElementID(61): // HT Operation Element ID is 61
					parseHTOperationIE(info, ieData)
				case layers.Dot11InformationElementID(192): // VHT Operation Element ID is 192
					parseVHTOperationIE(info, ieData)
				case layers.Dot11InformationElementID(0xff): // Check for Extension (255), assuming layers.Dot11InformationElementIDExtension is not defined for linter
					if len(ieData) > 0 {
						extensionID := ieData[0]
						const heCapabilitiesExtID uint8 = 35
						if extensionID == heCapabilitiesExtID {
							if info.ParsedHECaps == nil {
								info.ParsedHECaps = &HECapabilityInfo{}
							}
							// extractHECapabilities(info, &layers.Dot11InformationElement{ID: ieID, Length: byte(ieLength), Info: ieData}, nil, nil)
						}
					}
				}
				currentIndex += 2 + ieLength
			}
			// Calls to parseSecurity and determineBandwidth would ideally be here, after loop
			// parseSecurity(packet, info)
			// determineBandwidth(info)
		}
	}

	if dot11.Type.MainType() == layers.Dot11TypeData {
		llcLayer := packet.Layer(layers.LayerTypeLLC)
		if llcLayer != nil {
			llc, _ := llcLayer.(*layers.LLC)
			if llc.DSAP == 0xAA && llc.SSAP == 0xAA && llc.Control == 0x03 {
				// SNAP packet
			}
		}
	}

	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ipv4, _ := ipLayer.(*layers.IPv4)
		info.TransportPayloadLength = int(ipv4.Length) - (int(ipv4.IHL) * 4)
	} else if ipLayer := packet.Layer(layers.LayerTypeIPv6); ipLayer != nil {
		ipv6, _ := ipLayer.(*layers.IPv6)
		info.TransportPayloadLength = int(ipv6.Length)
	}

	if info.RadiotapDataRate > 0 {
		info.BitRate = info.RadiotapDataRate
		info.PHYRateMbps = info.RadiotapDataRate
	}

	// --- Determine Bandwidth based on parsed IEs and Radiotap ---
	// This logic should be placed after all relevant IEs have been parsed.
	foundBandwidth := false
	// 1. VHT Operation IE (Most accurate if present)
	if info.ParsedVHTCaps != nil && info.ParsedVHTCaps.ChannelWidth != "" {
		if info.ParsedVHTCaps.ChannelWidth == "20_40" {
			// Fallback to HT Operation or capabilities for 20 vs 40 decision
			// Check HT Operation first if available
			if info.ParsedHTCaps != nil && info.ParsedHTCaps.PrimaryChannel != 0 { // HT Op was parsed
				if info.ParsedHTCaps.OperatingAt40MHz {
					info.Bandwidth = "40MHz"
				} else {
					info.Bandwidth = "20MHz"
				}
			} else if info.RadiotapMCSBw == 1 { // Fallback to Radiotap if no HT Op
				info.Bandwidth = "40MHz"
			} else {
				info.Bandwidth = "20MHz"
			}
		} else {
			info.Bandwidth = info.ParsedVHTCaps.ChannelWidth + "MHz"
		}
		foundBandwidth = true
	}

	// 2. HT Operation IE (If no VHT Operation)
	if !foundBandwidth && info.ParsedHTCaps != nil && info.ParsedHTCaps.PrimaryChannel != 0 { // PrimaryChannel check indicates HT Op was likely parsed
		if info.ParsedHTCaps.OperatingAt40MHz { // This flag is now set by parseHTOperationIE
			info.Bandwidth = "40MHz"
		} else {
			info.Bandwidth = "20MHz"
		}
		foundBandwidth = true
	}

	// 3. VHT Capabilities IE (If no Operation IEs)
	if !foundBandwidth && info.ParsedVHTCaps != nil {
		// SupportedChannelWidthSet from VHT Caps: 0 (20/40), 1 (80), 2 (160/80+80)
		// This is complex; for now, if VHT caps are present, rely on Radiotap or default to 20/40 from Radiotap.
		if info.RadiotapMCSBw == 1 { // Radiotap says 40MHz
			info.Bandwidth = "40MHz"
		} else {
			info.Bandwidth = "20MHz"
		}
		// A more detailed VHT Cap check would look at info.ParsedVHTCaps.SupportedChannelWidthSet
		foundBandwidth = true
	}

	// 4. HT Capabilities IE (If only HT Capabilities)
	if !foundBandwidth && info.ParsedHTCaps != nil {
		if info.ParsedHTCaps.ChannelWidth40MHz { // From HT Capabilities IE
			info.Bandwidth = "40MHz"
		} else {
			info.Bandwidth = "20MHz"
		}
		foundBandwidth = true
	}

	// 5. Radiotap Fallback (If no relevant IEs parsed or they don't specify width)
	if !foundBandwidth {
		if info.RadiotapMCSBw == 1 { // 1 for 40MHz from Radiotap MCS flags
			info.Bandwidth = "40MHz"
		} else {
			info.Bandwidth = "20MHz"
		}
		foundBandwidth = true // Or consider it a default rather than found
	}

	// 6. Default (Should ideally be covered by Radiotap fallback)
	// if !foundBandwidth { // Should not happen if Radiotap fallback is comprehensive
	// 	info.Bandwidth = "20MHz"
	// }

	if dot11.Flags.WEP() {
		info.Security = "WEP"
	} else if info.Security == "" {
		info.Security = "Open/Unknown"
	}

	return info, nil
}

// --- Information Element Parsers ---

// parseHTCapabilitiesIE parses the HT Capabilities information element.
// Reference: IEEE 802.11-2016, Section 9.4.2.56 (HT Capabilities element)
// Reference: gopacket-80211.md for field details
func parseHTCapabilitiesIE(info *ParsedFrameInfo, ieData []byte) {
	if info.ParsedHTCaps == nil {
		info.ParsedHTCaps = &HTCapabilityInfo{}
	}
	logger.Log.Debug().Int("ht_cap_ie_len", len(ieData)).Msg("Parsing HT Capabilities IE")

	if len(ieData) < 2 { // Minimum: HT Capabilities Information field (2 bytes)
		logger.Log.Warn().Msg("HT Capabilities IE too short for HT Capability Info field.")
		return
	}

	// HT Capabilities Information field (2 bytes, little-endian)
	htCapInfoField := binary.LittleEndian.Uint16(ieData[0:2])
	info.ParsedHTCaps.LDPCCoding = (htCapInfoField & (1 << 0)) != 0
	info.ParsedHTCaps.ChannelWidth40MHz = (htCapInfoField & (1 << 1)) != 0 // Bit 1: Supported Channel Width Set (0: 20MHz only, 1: 20/40MHz)
	// SMPS (Spatial Multiplexing Power Save) (bits 2-3)
	// info.ParsedHTCaps.SMPSMode = (htCapInfoField >> 2) & 0x03
	info.ParsedHTCaps.ShortGI20MHz = (htCapInfoField & (1 << 5)) != 0
	info.ParsedHTCaps.ShortGI40MHz = (htCapInfoField & (1 << 6)) != 0
	info.ParsedHTCaps.TxSTBC = (htCapInfoField & (1 << 7)) != 0
	info.ParsedHTCaps.RxSTBC = uint8((htCapInfoField >> 8) & 0x03) // Bits 8-9: Rx STBC
	info.ParsedHTCaps.HTDelayedBlockAck = (htCapInfoField & (1 << 10)) != 0
	// Max A-MSDU Length (bit 11): 0 for 3839 bytes, 1 for 7935 bytes
	if (htCapInfoField & (1 << 11)) != 0 {
		info.ParsedHTCaps.MaxAMSDULength = 7935
	} else {
		info.ParsedHTCaps.MaxAMSDULength = 3839
	}
	info.ParsedHTCaps.DSSCck = (htCapInfoField & (1 << 12)) != 0 // DSSS/CCK Mode in 40 MHz (Bit 12)
	// Bit 13 is reserved
	info.ParsedHTCaps.FortyMhzIntolerant = (htCapInfoField & (1 << 14)) != 0 // Bit 14: 40 MHz Intolerant
	// Bit 15: L-SIG TXOP Protection Support

	currentIndex := 2

	// A-MPDU Parameters (1 byte)
	if len(ieData) >= currentIndex+1 {
		ampduParams := ieData[currentIndex]
		// Max A-MPDU Length Exponent (bits 0-1) -> 2^(13 + exponent) - 1 bytes
		maxRxAMPDUExp := ampduParams & 0x03
		info.ParsedHTCaps.MaxAMPDULength = (1 << (13 + maxRxAMPDUExp)) - 1
		// MPDU Density (bits 2-4)
		// info.ParsedHTCaps.MPDUDensity = (ampduParams >> 2) & 0x07
		currentIndex++
	} else {
		logger.Log.Warn().Msg("HT Capabilities IE too short for A-MPDU Parameters.")
		// Not returning, as MCS set might still be parsable if length is non-standard
	}

	// Supported MCS Set (16 bytes)
	// Rx MCS Bitmask (78 bits = 10 bytes), Tx MCS Set Defined (1 bit), Tx Rx MCS Set Not Equal (1 bit), Max Spatial Streams Supported (2 bits), etc.
	if len(ieData) >= currentIndex+16 {
		info.ParsedHTCaps.SupportedMCSSet = make([]byte, 16)
		copy(info.ParsedHTCaps.SupportedMCSSet, ieData[currentIndex:currentIndex+16])
		currentIndex += 16
	} else {
		logger.Log.Warn().Msg("HT Capabilities IE too short for full Supported MCS Set.")
	}

	// HT Extended Capabilities (2 bytes) - if present
	if len(ieData) >= currentIndex+2 {
		// extHtCapInfo := binary.LittleEndian.Uint16(ieData[currentIndex:currentIndex+2])
		// PCO (Phased Coexistence Operation) (bit 0)
		// TDC (Transmit Diversity MCSs) (bit 1)
		// MCS Feedback (bits 8-9)
		// ... and others
		currentIndex += 2
	}

	// Transmit Beamforming Capabilities (4 bytes) - if present
	if len(ieData) >= currentIndex+4 {
		// txBFCap := binary.LittleEndian.Uint32(ieData[currentIndex:currentIndex+4])
		// ... many bitfields ...
		currentIndex += 4
	}

	// Antenna Selection Capabilities (1 byte) - if present
	if len(ieData) >= currentIndex+1 {
		// aselCap := ieData[currentIndex]
		// ... bitfields ...
		currentIndex++
	}

	logger.Log.Debug().Interface("parsed_ht_caps", info.ParsedHTCaps).Msg("HT Capabilities IE Parsed")
}

// parseVHTCapabilitiesIE parses the VHT Capabilities information element.
// Reference: IEEE 802.11-2016, Section 9.4.2.158 (VHT Capabilities element)
// Reference: gopacket-80211.md for field details
func parseVHTCapabilitiesIE(info *ParsedFrameInfo, ieData []byte) {
	if info.ParsedVHTCaps == nil {
		info.ParsedVHTCaps = &VHTCapabilityInfo{}
	}
	logger.Log.Debug().Int("vht_cap_ie_len", len(ieData)).Msg("Parsing VHT Capabilities IE")

	// IE length for VHT Capabilities is typically 12 bytes
	if len(ieData) < 12 { // VHT Capabilities Info (4 bytes) + Supported VHT-MCS and NSS Set (8 bytes)
		logger.Log.Warn().Msg("VHT Capabilities IE too short for mandatory fields.")
		return
	}

	// VHT Capabilities Info (4 bytes, little-endian)
	vhtCapInfo := binary.LittleEndian.Uint32(ieData[0:4])
	info.ParsedVHTCaps.MaxMPDULength = uint8(vhtCapInfo & 0x03)                   // Bits 0-1
	info.ParsedVHTCaps.SupportedChannelWidthSet = uint8((vhtCapInfo >> 2) & 0x03) // Bits 2-3. 0: 80MHz, 1: 160MHz, 2: 160MHz (80+80), 3: reserved
	info.ParsedVHTCaps.RxLDPC = (vhtCapInfo & (1 << 4)) != 0
	info.ParsedVHTCaps.ShortGI80MHz = (vhtCapInfo & (1 << 5)) != 0
	info.ParsedVHTCaps.ShortGI160MHz = (vhtCapInfo & (1 << 6)) != 0 // Also for 80+80 MHz
	info.ParsedVHTCaps.TxSTBC = (vhtCapInfo & (1 << 7)) != 0
	info.ParsedVHTCaps.RxSTBC = uint8((vhtCapInfo >> 8) & 0x07) // Bits 8-10
	info.ParsedVHTCaps.SUBeamformerCapable = (vhtCapInfo & (1 << 11)) != 0
	info.ParsedVHTCaps.SUBeamformee = (vhtCapInfo & (1 << 12)) != 0
	info.ParsedVHTCaps.BeamformeeSTS = uint8((vhtCapInfo >> 13) & 0x07)      // Bits 13-15: Beamformee STS Capability
	info.ParsedVHTCaps.SoundingDimensions = uint8((vhtCapInfo >> 16) & 0x07) // Bits 16-18: Number of Sounding Dimensions
	info.ParsedVHTCaps.MUBeamformerCapable = (vhtCapInfo & (1 << 19)) != 0
	info.ParsedVHTCaps.MUBeamformee = (vhtCapInfo & (1 << 20)) != 0
	info.ParsedVHTCaps.VHTTXOPPSCapability = (vhtCapInfo & (1 << 21)) != 0  // VHT TXOP PS
	info.ParsedVHTCaps.VHTHTCCapability = (vhtCapInfo & (1 << 22)) != 0     // +HTC-VHT Capable
	info.ParsedVHTCaps.MaxAMPDULengthExp = uint8((vhtCapInfo >> 23) & 0x07) // Bits 23-25: Max A-MPDU Length Exponent
	// Bits 26-27: VHT Link Adaptation Capable
	info.ParsedVHTCaps.RxPatternConsistency = (vhtCapInfo & (1 << 28)) != 0 // Rx Antenna Pattern Consistency
	info.ParsedVHTCaps.TxPatternConsistency = (vhtCapInfo & (1 << 29)) != 0 // Tx Antenna Pattern Consistency

	// Supported VHT-MCS and NSS Set (8 bytes)
	// Rx VHT-MCS Map (2 bytes)
	info.ParsedVHTCaps.RxMCSMap = binary.LittleEndian.Uint16(ieData[4:6])
	// Rx Highest VHT Data Rate (2 bytes, but only 13 bits used)
	info.ParsedVHTCaps.RxHighestLongGIRate = binary.LittleEndian.Uint16(ieData[6:8]) & 0x1FFF // Mask for 13 bits

	// Tx VHT-MCS Map (2 bytes)
	info.ParsedVHTCaps.TxMCSMap = binary.LittleEndian.Uint16(ieData[8:10])
	// Tx Highest VHT Data Rate (2 bytes, but only 13 bits used)
	info.ParsedVHTCaps.TxHighestLongGIRate = binary.LittleEndian.Uint16(ieData[10:12]) & 0x1FFF // Mask for 13 bits

	logger.Log.Debug().Interface("parsed_vht_caps", info.ParsedVHTCaps).Msg("VHT Capabilities IE Parsed")
}

// parseRSNIE parses the RSN (Robust Security Network) information element.
// Reference: IEEE 802.11-2016, Section 9.4.2.25 (RSN element)
// Reference: gopacket-80211.md for RSN structure and cipher/AKM suites
func parseRSNIE(info *ParsedFrameInfo, ieData []byte) {
	logger.Log.Debug().Int("ie_len", len(ieData)).Msg("Parsing RSN IE")
	if len(ieData) < 2 { // Version (2 bytes)
		logger.Log.Warn().Msg("RSN IE too short for Version.")
		return
	}
	version := binary.LittleEndian.Uint16(ieData[0:2])
	if version != 1 {
		logger.Log.Warn().Uint16("rsn_version", version).Msg("Unsupported RSN version.")
		return // Or handle differently if other versions become relevant
	}

	currentIndex := 2
	var groupCipherStr string
	pairwiseCiphers := []string{}
	akms := []string{}
	// var rsnCaps uint16 // For later if needed

	// Group Data Cipher Suite (4 bytes: OUI[3] + Suite Type[1])
	if currentIndex+4 <= len(ieData) {
		groupCipherStr = ouiAndCipherSuiteToString(ieData[currentIndex:currentIndex+3], ieData[currentIndex+3])
		currentIndex += 4
	} else {
		logger.Log.Warn().Msg("RSN IE too short for Group Cipher Suite.")
		return
	}

	// Pairwise Cipher Suite Count (2 bytes)
	if currentIndex+2 <= len(ieData) {
		pairwiseCipherCount := int(binary.LittleEndian.Uint16(ieData[currentIndex : currentIndex+2]))
		currentIndex += 2
		for i := 0; i < pairwiseCipherCount; i++ {
			if currentIndex+4 <= len(ieData) {
				cipherStr := ouiAndCipherSuiteToString(ieData[currentIndex:currentIndex+3], ieData[currentIndex+3])
				pairwiseCiphers = append(pairwiseCiphers, cipherStr)
				currentIndex += 4
			} else {
				logger.Log.Warn().Int("expected_pairwise_count", pairwiseCipherCount).Int("parsed_count", i).Msg("RSN IE ended prematurely while parsing Pairwise Cipher Suites.")
				break
			}
		}
	} else {
		logger.Log.Warn().Msg("RSN IE too short for Pairwise Cipher Suite Count.")
		return
	}

	// AKM Suite Count (2 bytes)
	if currentIndex+2 <= len(ieData) {
		akmSuiteCount := int(binary.LittleEndian.Uint16(ieData[currentIndex : currentIndex+2]))
		currentIndex += 2
		for i := 0; i < akmSuiteCount; i++ {
			if currentIndex+4 <= len(ieData) {
				akmStr := ouiAndAKMSuiteToString(ieData[currentIndex:currentIndex+3], ieData[currentIndex+3])
				akms = append(akms, akmStr)
				currentIndex += 4
			} else {
				logger.Log.Warn().Int("expected_akm_count", akmSuiteCount).Int("parsed_count", i).Msg("RSN IE ended prematurely while parsing AKM Suites.")
				break
			}
		}
	} else {
		logger.Log.Warn().Msg("RSN IE too short for AKM Suite Count.")
		return
	}

	// (Optional) RSN Capabilities (2 bytes) - parse if present and needed
	// if currentIndex+2 <= len(ieData) {
	// 	rsnCaps = binary.LittleEndian.Uint16(ieData[currentIndex : currentIndex+2])
	// 	currentIndex += 2
	// 	// Process rsnCaps, e.g., Pre-Auth (bit 0), No Pairwise (bit 1), PTKSA Replay Counter (bits 2-3), etc.
	// }

	// Determine overall security string
	// This is a simplified logic. Real determination can be more complex based on combinations.
	var finalAKM, finalPairwise string
	if len(akms) > 0 {
		// Prioritize known strong AKMs if multiple are present
		for _, akm := range akms {
			if akm == "SAE" || akm == "PSK" || akm == "802.1X" { // Add other relevant AKMs like FT-PSK, FT-802.1X
				finalAKM = akm
				break
			}
		}
		if finalAKM == "" {
			finalAKM = akms[0]
		} // Fallback to first listed
	}

	if len(pairwiseCiphers) > 0 {
		// Prioritize known strong ciphers
		for _, pc := range pairwiseCiphers {
			if pc == "CCMP-128" || pc == "GCMP-256" || pc == "CCMP-256" || pc == "GCMP-128" {
				finalPairwise = pc
				break
			}
		}
		if finalPairwise == "" {
			finalPairwise = pairwiseCiphers[0]
		} // Fallback to first listed
	}

	if finalAKM != "" && finalPairwise != "" {
		// Construct a common security string, e.g., WPA2-PSK-CCMP, WPA3-SAE-CCMP
		securityName := ""
		switch finalAKM {
		case "PSK":
			if finalPairwise == "CCMP-128" {
				securityName = "WPA2-PSK"
			} // Common assumption
			// Add WPA-PSK if TKIP is primary pairwise
		case "802.1X":
			if finalPairwise == "CCMP-128" {
				securityName = "WPA2-Enterprise"
			}
		case "SAE":
			securityName = "WPA3-Personal"
		// Add more cases for FT, OWE, etc.
		default:
			securityName = finalAKM // Use raw AKM if no specific WPAx name
		}
		if securityName != "" {
			info.Security = fmt.Sprintf("%s (%s)", securityName, finalPairwise)
		} else {
			info.Security = fmt.Sprintf("RSN AKM: %s, Pairwise: %s", strings.Join(akms, "/"), strings.Join(pairwiseCiphers, "/"))
		}
	} else if groupCipherStr != "" { // Fallback if somehow AKM/Pairwise are missing but group is there
		info.Security = fmt.Sprintf("RSN Group: %s", groupCipherStr)
	} else {
		info.Security = "RSN (Unknown)" // Should not happen if RSN IE is valid
	}
	logger.Log.Info().Str("parsed_security", info.Security).Msg("RSN IE Parsed")
}

// parseHTOperationIE parses the HT Operation information element.
// Reference: IEEE 802.11-2016, Section 9.4.2.57 (HT Operation element)
func parseHTOperationIE(info *ParsedFrameInfo, ieData []byte) {
	if info.ParsedHTCaps == nil { // HT Operation usually appears with HT Capabilities, but init just in case
		info.ParsedHTCaps = &HTCapabilityInfo{}
	}
	logger.Log.Debug().Int("ht_op_ie_len", len(ieData)).Msg("Parsing HT Operation IE")

	if len(ieData) < 1 { // Primary Channel (1 byte)
		logger.Log.Warn().Msg("HT Operation IE too short for Primary Channel.")
		return
	}
	info.ParsedHTCaps.PrimaryChannel = ieData[0]
	currentIndex := 1

	if len(ieData) < currentIndex+1 { // HT Operation Information (Byte 1 of 5-byte set)
		logger.Log.Warn().Msg("HT Operation IE too short for HT Operation Information Set 1.")
		return
	}
	htOpInfoSet1 := ieData[currentIndex]
	secondaryChanOffsetValue := htOpInfoSet1 & 0x03 // Bits 0-1: Secondary Channel Offset
	switch secondaryChanOffsetValue {
	case 0:
		info.ParsedHTCaps.SecondaryChannelOffset = "None"
	case 1:
		info.ParsedHTCaps.SecondaryChannelOffset = "Above"
	case 3:
		info.ParsedHTCaps.SecondaryChannelOffset = "Below"
	default:
		info.ParsedHTCaps.SecondaryChannelOffset = "Reserved"
	}

	staChannelWidthIsAny := (htOpInfoSet1 >> 2) & 0x01 // Bit 2: STA Channel Width. 0 = 20MHz, 1 = Any (20MHz or 40MHz).

	if staChannelWidthIsAny == 1 && (secondaryChanOffsetValue == 1 || secondaryChanOffsetValue == 3) {
		info.ParsedHTCaps.OperatingAt40MHz = true
	} else {
		info.ParsedHTCaps.OperatingAt40MHz = false
	}

	currentIndex++
	currentIndex += 4 // Skip over the rest of HT Operation Information (bytes 2,3,4,5 of the set)

	if len(ieData) >= currentIndex+16 {
		currentIndex += 16
	}

	logger.Log.Debug().Interface("parsed_ht_op_fields_in_ht_caps", info.ParsedHTCaps).Msg("HT Operation IE Parsed (relevant fields stored in ParsedHTCaps)")
}

// parseVHTOperationIE parses the VHT Operation information element.
// Reference: IEEE 802.11-2016, Section 9.4.2.159 (VHT Operation element)
func parseVHTOperationIE(info *ParsedFrameInfo, ieData []byte) {
	if info.ParsedVHTCaps == nil { // VHT Operation usually appears with VHT Capabilities, but init just in case
		info.ParsedVHTCaps = &VHTCapabilityInfo{}
	}
	logger.Log.Debug().Int("vht_op_ie_len", len(ieData)).Msg("Parsing VHT Operation IE")

	// VHT Operation IE has a fixed length of 5 bytes if a narrower VHT CBW is used (e.g. 20/40MHz)
	// or can be longer if wider CBWs are signaled with HE variants. Std length is 5.
	if len(ieData) < 1 { // VHT Channel Width (1 byte)
		logger.Log.Warn().Msg("VHT Operation IE too short for VHT Channel Width.")
		return
	}

	// VHT Operation Info (1st byte): Channel Width
	vhtOpChannelWidth := ieData[0]
	// This field directly sets the operating channel width for VHT.
	// It will be used by the main bandwidth determination logic.
	switch vhtOpChannelWidth {
	case 0: // Operates in 20MHz or 40MHz. Actual width determined by HT Operation element or other means.
		// Store a temporary indicator or let main bandwidth logic handle this based on HT Op info.
		info.ParsedVHTCaps.ChannelWidth = "20_40" // Special value to indicate it needs further check from HT Op
	case 1:
		info.ParsedVHTCaps.ChannelWidth = "80"
	case 2:
		info.ParsedVHTCaps.ChannelWidth = "160"
	case 3:
		info.ParsedVHTCaps.ChannelWidth = "80+80"
	default:
		if vhtOpChannelWidth >= 4 && vhtOpChannelWidth <= 255 { // As per IEEE 802.11-2016, Table 9-247
			// These are reserved or might be for future/HE-variant operations if IE structure is overloaded.
			// For pure VHT Operation IE (5 bytes), these are typically not used.
			info.ParsedVHTCaps.ChannelWidth = fmt.Sprintf("ReservedVHTBW-%d", vhtOpChannelWidth)
		} else {
			info.ParsedVHTCaps.ChannelWidth = "Unknown"
		}
	}
	currentIndex := 1

	// VHT Operation Info (2nd byte): Channel Center Frequency Segment 0
	if len(ieData) >= currentIndex+1 {
		info.ParsedVHTCaps.ChannelCenter0 = ieData[currentIndex]
		currentIndex++
	} else {
		logger.Log.Warn().Msg("VHT Operation IE too short for Channel Center Freq Seg 0.")
		// Allow partial parse if only width was present
		logger.Log.Debug().Interface("parsed_vht_op_fields_in_vht_caps", info.ParsedVHTCaps).Msg("VHT Operation IE Parsed (partially)")
		return
	}

	// VHT Operation Info (3rd byte): Channel Center Frequency Segment 1
	if len(ieData) >= currentIndex+1 {
		info.ParsedVHTCaps.ChannelCenter1 = ieData[currentIndex]
		currentIndex++
	} else {
		logger.Log.Warn().Msg("VHT Operation IE too short for Channel Center Freq Seg 1.")
		logger.Log.Debug().Interface("parsed_vht_op_fields_in_vht_caps", info.ParsedVHTCaps).Msg("VHT Operation IE Parsed (partially)")
		return
	}

	// Basic VHT-MCS and NSS Set (2 bytes) - if present (IE length would be 5)
	if len(ieData) >= currentIndex+2 {
		// basicVHTMCSNSS := binary.LittleEndian.Uint16(ieData[currentIndex : currentIndex+2])
		// This field indicates the set of VHT-MCSs and NSS values that all STAs in the BSS must support for the operating channel width.
		// Example: For NSS=1, MCS0-7 (bits 0-1), NSS=2, MCS0-7 (bits 2-3) ... NSS=8, MCS0-7 (bits 14-15)
		// We might not need to store this raw value directly in ParsedVHTCaps unless for very specific analysis.
		currentIndex += 2
	}

	logger.Log.Debug().Interface("parsed_vht_op_fields_in_vht_caps", info.ParsedVHTCaps).Msg("VHT Operation IE Parsed")
}

// (Optional) Placeholder for HE Capabilities and Operation parsing if needed later
// func parseHECapabilitiesIE(info *ParsedFrameInfo, ieData []byte) {}
// func parseHEOperationIE(info *ParsedFrameInfo, ieData []byte) {}

// Helper function for RSN IE parsing (Placeholder)
// ouiAndTypeToString needs to be split for Cipher Suites and AKM Suites for clarity and correctness
func ouiAndCipherSuiteToString(oui []byte, suiteType byte) string {
	if len(oui) != 3 {
		return "InvalidOUI"
	}
	// IEEE Std 802.11-2016, Table 9-131—Cipher suite selectors
	ouiStr := fmt.Sprintf("%02X-%02X-%02X", oui[0], oui[1], oui[2])
	if ouiStr == "00-0F-AC" { //检查OUI是否为IEEE分配的Cipher/AKM OUI
		switch suiteType {
		case 0:
			return "Use Group Cipher"
		case 1:
			return "WEP-40"
		case 2:
			return "TKIP"
		// case 3 is reserved
		case 4:
			return "CCMP-128" // AES-CCMP
		case 5:
			return "WEP-104"
		case 6:
			return "BIP-CMAC-128" // For Management Frames (MFP)
		case 7:
			return "Group Addressed Traffic Not Allowed"
		case 8:
			return "GCMP-128"
		case 9:
			return "GCMP-256"
		case 10:
			return "CCMP-256"
		case 11:
			return "BIP-GMAC-128"
		case 12:
			return "BIP-GMAC-256"
		case 13:
			return "BIP-CMAC-256"
		default:
			return fmt.Sprintf("Cipher-Unknown(%d)", suiteType)
		}
	}
	return fmt.Sprintf("%s:%d", ouiStr, suiteType) // Non-standard or vendor specific
}

func ouiAndAKMSuiteToString(oui []byte, suiteType byte) string {
	if len(oui) != 3 {
		return "InvalidOUI"
	}
	// IEEE Std 802.11-2016, Table 9-134—AKM suite selectors
	ouiStr := fmt.Sprintf("%02X-%02X-%02X", oui[0], oui[1], oui[2])
	if ouiStr == "00-0F-AC" { //检查OUI是否为IEEE分配的Cipher/AKM OUI
		switch suiteType {
		// case 0 is Reserved
		case 1:
			return "802.1X" // IEEE 802.1X AKM
		case 2:
			return "PSK" // PSK (Pre-Shared Key) AKM
		case 3:
			return "FT-802.1X"
		case 4:
			return "FT-PSK"
		case 5:
			return "WPA-SHA256-802.1X" // 802.1X with SHA256 KDF
		case 6:
			return "WPA-SHA256-PSK" // PSK with SHA256 KDF
		case 7:
			return "TDLS"
		case 8:
			return "SAE" // Simultaneous Authentication of Equals (WPA3)
		case 9:
			return "FT-SAE"
		case 11:
			return "OWE" // Opportunistic Wireless Encryption
		case 12:
			return "FT-SUITEB-SHA256"
		case 13:
			return "FT-SUITEB-SHA384"
		// Other types exist for FILS, etc.
		default:
			return fmt.Sprintf("AKM-Unknown(%d)", suiteType)
		}
	}
	// Add other OUIs if necessary, e.g., Microsoft WPA OUI 00-50-F2, Type 1
	if ouiStr == "00-50-F2" && suiteType == 1 {
		return "WPA" // For original WPA (uses TKIP typically, but AKM is WPA)
	}
	return fmt.Sprintf("%s:%d", ouiStr, suiteType)
}

// The existing ouiAndTypeToString can be removed or kept if used elsewhere for generic OUI:Type formatting
// func ouiAndTypeToString(oui []byte, suiteType byte) string { ... }
