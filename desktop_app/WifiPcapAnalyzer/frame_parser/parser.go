package frame_parser

import (
	"WifiPcapAnalyzer/utils" // For utility functions like FrequencyToChannel
	"fmt"
	"io"
	"log"
	"net"

	// "strings" // Import strings for LayerType names - Removed as it's unused after commenting logs
	"time"
	"unicode/utf8" // For SSID validation

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

// HTCapabilityInfo stores parsed HT capabilities.
type HTCapabilityInfo struct {
	ChannelWidth40MHz bool   `json:"channel_width_40mhz"`
	ShortGI20MHz      bool   `json:"short_gi_20mhz"`
	ShortGI40MHz      bool   `json:"short_gi_40mhz"`
	SupportedMCSSet   []byte `json:"supported_mcs_set"` // Raw 16 bytes
}

// VHTCapabilityInfo stores parsed VHT capabilities.
type VHTCapabilityInfo struct {
	// VHT Cap Info Field (first 4 bytes of VHT Capabilities IE)
	MaxMPDULength            uint8 // Bit 0-1
	SupportedChannelWidthSet uint8 // Bit 2-3 (0: 20/40, 1: 80, 2: 160/80+80, 3: 160)
	ShortGI80MHz             bool  // Bit 5
	ShortGI160MHz            bool  // Bit 6
	SUBeamformerCapable      bool  // Bit 8 (of byte 1 of VHT Cap Info)
	MUBeamformerCapable      bool  // Bit 11 (of byte 1 of VHT Cap Info)

	// VHT MCS and NSS Set Field (8 bytes)
	RxMCSMap            uint16 // NSS 1-8, 2 bits per NSS (0: MCS 0-7, 1: MCS 0-8, 2: MCS 0-9, 3: Not supported)
	RxHighestLongGIRate uint16 // Bits 10-12 of Rx MCS Map (byte 1, bits 2-4 of VHT MCS Set field)
	TxMCSMap            uint16 // Similar to RxMCSMap
	TxHighestLongGIRate uint16 // Similar to RxHighestLongGIRate
}

// ParsedFrameInfo holds extracted information from a single 802.11 frame.
type ParsedFrameInfo struct {
	Timestamp              time.Time
	FrameType              layers.Dot11Type
	BSSID                  net.HardwareAddr
	SA                     net.HardwareAddr
	DA                     net.HardwareAddr
	RA                     net.HardwareAddr
	TA                     net.HardwareAddr
	Channel                int
	Frequency              int
	SignalStrength         int
	NoiseLevel             int
	MCS                    *layers.RadioTapMCS
	Flags                  layers.RadioTapFlags
	Bandwidth              string
	SSID                   string
	SupportedRates         []byte
	ExtendedSuppRates      []byte
	DSSetChannel           uint8
	TIM                    []byte
	HTCapabilitiesRaw      []byte
	VHTCapabilitiesRaw     []byte
	HECapabilitiesRaw      []byte
	VHTOperationRaw        []byte // New field for VHT Operation IE
	RSNRaw                 []byte
	IsQoSData              bool
	ParsedHTCaps           *HTCapabilityInfo
	ParsedVHTCaps          *VHTCapabilityInfo
	FrameLength            int              // Length of the frame in bytes (MAC frame, not including FCS if possible, or physical frame)
	PHYRateMbps            float64          // Estimated PHY rate in Mbps
	IsShortPreamble        bool             // From Radiotap Flags or inferred
	IsShortGI              bool             // From Radiotap MCS flags or HT/VHT/HE capabilities
	RadioTapLayer          *layers.RadioTap // Store the parsed RadioTap layer
	Dot11Layer             *layers.Dot11    // Store the parsed Dot11 layer
	TransportPayloadLength int              // Length of the transport layer payload (TCP/UDP)
	MACDurationID          uint16           // Duration/ID field from MAC header
}

// PacketInfoHandler is a function that processes parsed frame information.
type PacketInfoHandler func(info *ParsedFrameInfo)

// ProcessPcapStream reads a pcap stream, parses individual packets, and calls the handler.
func ProcessPcapStream(pcapStream io.Reader, pktHandler PacketInfoHandler) {
	pcapReader, err := pcapgo.NewReader(pcapStream)
	if err != nil {
		log.Printf("Error creating pcap reader: %v", err)
		return
	}

	log.Println("PCAP Reader created, starting to read packets...")
	packetCount := 0
	for {
		data, ci, err := pcapReader.ReadPacketData()
		if err != nil {
			if err == io.EOF {
				log.Println("EOF reached in pcap stream.")
				break
			}
			log.Printf("Error reading packet data from pcap stream: %v", err)
			break
		}

		parsedInfo, err := parsePacketLayers(data, pcapReader.LinkType(), ci.Timestamp)
		if err != nil {
			log.Printf("Error parsing packet layers: %v. Packet data length: %d", err, len(data))
			snippetLen := 20
			if len(data) < snippetLen {
				snippetLen = len(data)
			}
			log.Printf("Problematic packet data snippet (first %d bytes): %x", snippetLen, data[:snippetLen])
			continue
		}

		if parsedInfo != nil {
			pktHandler(parsedInfo)
		}
		packetCount++
		if packetCount%100 == 0 {
			log.Printf("Processed %d packets from pcap stream...", packetCount)
		}
	}
	log.Printf("Finished processing pcap stream. Total packets processed: %d", packetCount)
}

// getPHYRateMbps estimates the PHY rate in Mbps based on Radiotap and other info.
// This is a simplified version and might need significant enhancements for accuracy across all 802.11 standards.
func getPHYRateMbps(rt *layers.RadioTap, dot11 *layers.Dot11, htCaps *HTCapabilityInfo, vhtCaps *VHTCapabilityInfo) float64 {
	if rt == nil {
		if dot11 != nil && dot11.Type.MainType() == layers.Dot11TypeMgmt {
			return 6.0 // Default fallback for Mgmt frames without Radiotap
		}
		return 1.0 // Cannot determine
	}

	// Check Radiotap Rate field first (legacy)
	if (rt.Present & layers.RadioTapPresentRate) != 0 {
		return float64(rt.Rate) * 0.5 // Rate is in 500 Kbps units
	}

	// Check Radiotap MCS field (HT/VHT/HE)
	if (rt.Present & layers.RadioTapPresentMCS) != 0 {
		mcs := rt.MCS.MCS // Corrected access
		if mcs <= 7 {
			return float64(mcs+1) * 6.5 // Extremely rough estimate
		}
		return 54.0 // Default fallback if MCS is higher
	}

	// Check Radiotap VHT field
	if (rt.Present & layers.RadioTapPresentVHT) != 0 {
		if len(rt.VHT.MCSNSS) > 0 {
			vhtUser1MCSNSS := rt.VHT.MCSNSS[0]
			nss := (vhtUser1MCSNSS >> 4) & 0x0F
			mcs := vhtUser1MCSNSS & 0x0F
			if nss == 0 {
				nss = 1
			}
			baseRate := 65.0
			switch rt.VHT.Bandwidth {
			case 1: // 40MHz
				baseRate *= 2
			case 2: // 80MHz
				baseRate *= 4
			case 3: // 160MHz or 80+80MHz
				baseRate *= 8
			}
			// Check Short GI using the correct flag bit 0x04
			if (rt.VHT.Flags & 0x04) != 0 { // 0x04 corresponds to Short GI bit
				baseRate *= 1.1
			}
			return baseRate * float64(nss) * (float64(mcs+1) / 8.0) // Highly simplified
		}
	}

	// Fallback to a common rate if nothing specific found
	return 6.0 // Default to 6 Mbps
}

func parsePacketLayers(rawData []byte, linkType layers.LinkType, captureTimestamp time.Time) (*ParsedFrameInfo, error) {
	info := &ParsedFrameInfo{
		Timestamp:   captureTimestamp,
		FrameLength: len(rawData), // Store the raw frame length
	}

	packet := gopacket.NewPacket(rawData, linkType, gopacket.Lazy)

	// var layerTypes []string
	// for _, layer := range packet.Layers() {
	// 	layerTypes = append(layerTypes, layer.LayerType().String())
	// }
	// // log.Printf("DEBUG_PACKET_LAYERS: All layers found by gopacket.NewPacket: [%s]. LinkType used: %s. Raw data length: %d", strings.Join(layerTypes, ", "), linkType.String(), len(rawData))

	// Check for decoding errors
	if errLayer := packet.ErrorLayer(); errLayer != nil {
		// If gopacket itself reports a decoding error, it's often severe enough to prevent Dot11 parsing.
		// Log the error and return, preventing this packet from being processed further.
		log.Printf("ERROR_DECODE_FAILURE: gopacket.NewPacket encountered an error: %v. Problematic data snippet (first 20 bytes of rawData): %x", errLayer.Error(), rawData[:20])
		return nil, fmt.Errorf("gopacket decoding error: %w", errLayer.Error())
	}

	radiotapLayer := packet.Layer(layers.LayerTypeRadioTap)
	if radiotapLayer == nil {
		dot11LayerCheck := packet.Layer(layers.LayerTypeDot11)
		if dot11LayerCheck != nil {
			log.Println("Warning: Radiotap layer not found, but Dot11 layer exists. Parsing will lack some radio details.")
		} else {
			snippetLen := 20
			if len(rawData) < snippetLen {
				snippetLen = len(rawData)
			}
			log.Printf("ERROR_NO_RADIOTAP_OR_DOT11: Radiotap layer not found and no Dot11 layer either. Raw data snippet (first %d bytes): %x", snippetLen, rawData[:snippetLen])
			return nil, fmt.Errorf("radiotap layer not found and no Dot11 layer either")
		}
	} else {
		rt, ok := radiotapLayer.(*layers.RadioTap)
		if !ok {
			return nil, fmt.Errorf("could not assert RadioTap layer")
		}
		if (rt.Present & layers.RadioTapPresentChannel) != 0 {
			info.Frequency = int(rt.ChannelFrequency)
			info.Channel = utils.FrequencyToChannel(info.Frequency)
		}
		if (rt.Present & layers.RadioTapPresentDBMAntennaSignal) != 0 {
			info.SignalStrength = int(rt.DBMAntennaSignal)
		}
		if (rt.Present & layers.RadioTapPresentDBMAntennaNoise) != 0 {
			info.NoiseLevel = int(rt.DBMAntennaNoise)
		}
		if (rt.Present & layers.RadioTapPresentMCS) != 0 {
			info.MCS = &rt.MCS
			// Check for Short GI from MCS flags if available (placeholder)
			// if rt.MCS.Flags indicates Short GI { info.IsShortGI = true }
		}
		if (rt.Present & layers.RadioTapPresentFlags) != 0 {
			info.Flags = rt.Flags
			if (rt.Flags & layers.RadioTapFlagsShortPreamble) != 0 {
				info.IsShortPreamble = true
			}
			if (rt.Flags & layers.RadioTapFlagsShortGI) != 0 {
				info.IsShortGI = true
			}
		}
		// Attempt to get PHY rate after parsing Radiotap
		info.PHYRateMbps = getPHYRateMbps(rt, nil, nil, nil)
	}

	if rtLog, okLog := radiotapLayer.(*layers.RadioTap); okLog && rtLog != nil {
		// log.Printf("DEBUG_RADIOTAP_INFO: Radiotap Version: %d, Radiotap Length Field (rt.Length): %d, Present Flags: %#v", rtLog.Version, rtLog.Length, rtLog.Present)
		// log.Printf("DEBUG_RADIOTAP_INFO: Radiotap Calculated Header Length (len(rtLog.Contents)): %d", len(rtLog.Contents))
		// log.Printf("DEBUG_RADIOTAP_INFO: Length of Radiotap's gopacket payload (len(rtLog.Payload)): %d", len(rtLog.Payload))
		// if len(rtLog.Payload) > 0 {
		// 	snippetLen := 30
		// 	if len(rtLog.Payload) < snippetLen {
		// 		snippetLen = len(rtLog.Payload)
		// 	}
		// 	log.Printf("DEBUG_RADIOTAP_INFO: rt.Payload snippet (first %d bytes): %x", snippetLen, rtLog.Payload[:snippetLen])
		// }
	} else if radiotapLayer != nil {
		// log.Printf("DEBUG_RADIOTAP_INFO: Radiotap layer present but not assertable to *layers.RadioTap, or rtLog is nil.")
	} else {
		// log.Printf("DEBUG_RADIOTAP_INFO: Radiotap layer is nil.")
	}

	dot11Layer := packet.Layer(layers.LayerTypeDot11)
	if dot11Layer == nil {
		log.Printf("ERROR_NO_DOT11_LAYER: Dot11 layer is nil. Radiotap present: %t. Raw data length: %d", radiotapLayer != nil, len(rawData))
		// Regardless of Radiotap presence, if Dot11 layer is nil, we cannot proceed with meaningful 802.11 analysis.
		// Return an error to ensure this packet is skipped by the caller.
		return nil, fmt.Errorf("critical failure: dot11 layer not found after gopacket processing. Radiotap present: %t", radiotapLayer != nil)
	}
	dot11, ok := dot11Layer.(*layers.Dot11)
	if !ok {
		return nil, fmt.Errorf("could not assert Dot11 layer")
	}

	// if dot11 != nil {
	// log.Printf("DEBUG_DOT11_INFO: Dot11 Type: %s", dot11.Type.String())
	// log.Printf("DEBUG_DOT11_INFO: Dot11 MAC Header Length (from gopacket Contents - len(dot11.Contents)): %d", len(dot11.Contents))
	// log.Printf("DEBUG_DOT11_INFO: Dot11 Payload Length (len(dot11.Payload)): %d", len(dot11.Payload))
	// if len(dot11.Payload) > 0 {
	// 	snippetLen := 60
	// 	if len(dot11.Payload) < snippetLen {
	// 		snippetLen = len(dot11.Payload)
	// 	}
	// 	log.Printf("DEBUG_DOT11_INFO: dot11.Payload snippet (first %d bytes): %x", snippetLen, dot11.Payload[:snippetLen])
	// } else if len(dot11.Payload) == 0 {
	// 	log.Printf("DEBUG_DOT11_INFO: dot11.Payload is EMPTY for FrameType: %s.", dot11.Type.String())
	// }
	// }

	info.FrameType = dot11.Type
	info.MACDurationID = dot11.DurationID

	toDS := dot11.Flags.ToDS()
	fromDS := dot11.Flags.FromDS()

	switch {
	case !toDS && !fromDS:
		info.DA = dot11.Address1
		info.SA = dot11.Address2
		info.BSSID = dot11.Address3
		info.RA = info.DA
		info.TA = info.SA
	case !toDS && fromDS:
		info.DA = dot11.Address1
		info.BSSID = dot11.Address2
		info.SA = dot11.Address3
		info.RA = info.DA
		info.TA = info.BSSID
	case toDS && !fromDS:
		info.BSSID = dot11.Address1
		info.SA = dot11.Address2
		info.DA = dot11.Address3
		info.RA = info.BSSID
		info.TA = info.SA
	case toDS && fromDS:
		info.RA = dot11.Address1
		info.TA = dot11.Address2
		info.DA = dot11.Address3
		if len(dot11.Address4) > 0 {
			info.SA = dot11.Address4
		}
		// log.Printf("DEBUG_WDS_FRAME: RA:%s, TA:%s, DA:%s, SA:%s", info.RA, info.TA, info.DA, info.SA)
	}

	if dot11.Type.MainType() == layers.Dot11TypeMgmt {
		var iePayload []byte
		originalPayload := dot11.Payload
		offsetApplied := 0

		bssidForLog := "N/A"
		if info.BSSID != nil {
			bssidForLog = info.BSSID.String()
		}

		switch dot11.Type {
		case layers.Dot11TypeMgmtBeacon, layers.Dot11TypeMgmtProbeResp:
			const fixedHeaderLen = 12
			offsetApplied = fixedHeaderLen
			if len(originalPayload) >= fixedHeaderLen {
				iePayload = originalPayload[fixedHeaderLen:]
				// log.Printf("DEBUG_MGMT_PAYLOAD_OFFSET: FrameType: %s, BSSID: %s, Applied %d-byte offset. OriginalPayloadLen: %d, EffectiveIEPayloadLen: %d", dot11.Type.String(), bssidForLog, fixedHeaderLen, len(originalPayload), len(iePayload))
			} else {
				log.Printf("WARN_MGMT_PAYLOAD_OFFSET: FrameType: %s, BSSID: %s, Payload too short for fixed header (expected %d, got %d). No IEs will be parsed. Frame will be skipped.", dot11.Type.String(), bssidForLog, fixedHeaderLen, len(originalPayload))
				return nil, fmt.Errorf("payload too short for fixed header (expected %d, got %d) for %s", fixedHeaderLen, len(originalPayload), dot11.Type.String())
			}
		case layers.Dot11TypeMgmtAssociationReq:
			const fixedHeaderLen = 4
			offsetApplied = fixedHeaderLen
			if len(originalPayload) >= fixedHeaderLen {
				iePayload = originalPayload[fixedHeaderLen:]
				// log.Printf("DEBUG_MGMT_PAYLOAD_OFFSET: FrameType: %s, BSSID: %s, Applied %d-byte offset. OriginalPayloadLen: %d, EffectiveIEPayloadLen: %d", dot11.Type.String(), bssidForLog, fixedHeaderLen, len(originalPayload), len(iePayload))
			} else {
				log.Printf("WARN_MGMT_PAYLOAD_OFFSET: FrameType: %s, BSSID: %s, Payload too short for fixed header (expected %d, got %d). No IEs will be parsed.", dot11.Type.String(), bssidForLog, fixedHeaderLen, len(originalPayload))
				return info, nil
			}
		case layers.Dot11TypeMgmtReassociationReq:
			// Fixed fields before IEs in payload: CapabilityInfo (2) + ListenInterval (2) = 4 bytes.
			// Current AP Address (6 bytes) is part of the MAC header (dot11.Address3), not the payload being offset here.
			const fixedHeaderLenReassoc = 4
			offsetApplied = fixedHeaderLenReassoc
			if len(originalPayload) >= fixedHeaderLenReassoc {
				iePayload = originalPayload[fixedHeaderLenReassoc:]
				// log.Printf("DEBUG_MGMT_PAYLOAD_OFFSET: FrameType: %s, BSSID: %s, Applied %d-byte offset. OriginalPayloadLen: %d, EffectiveIEPayloadLen: %d", dot11.Type.String(), bssidForLog, fixedHeaderLenReassoc, len(originalPayload), len(iePayload))
			} else {
				log.Printf("WARN_MGMT_PAYLOAD_OFFSET: FrameType: %s, BSSID: %s, Payload too short for fixed header (expected %d, got %d). No IEs will be parsed.", dot11.Type.String(), bssidForLog, fixedHeaderLenReassoc, len(originalPayload))
				return info, nil
			}
		case layers.Dot11TypeMgmtMeasurementPilot:
			// MgmtMeasurementPilot fixed header: Category (1) + Action (1) + Dialog Token (1) = 3 bytes
			const fixedHeaderLenPilot = 3
			offsetApplied = fixedHeaderLenPilot
			if len(originalPayload) >= fixedHeaderLenPilot {
				iePayload = originalPayload[fixedHeaderLenPilot:]
				// log.Printf("DEBUG_MGMT_PAYLOAD_OFFSET: FrameType: %s, BSSID: %s, Applied %d-byte offset. OriginalPayloadLen: %d, EffectiveIEPayloadLen: %d", dot11.Type.String(), bssidForLog, fixedHeaderLenPilot, len(originalPayload), len(iePayload))
			} else {
				log.Printf("WARN_MGMT_PAYLOAD_OFFSET: FrameType: %s, BSSID: %s, Payload too short for fixed header (expected %d, got %d). No IEs will be parsed.", dot11.Type.String(), bssidForLog, fixedHeaderLenPilot, len(originalPayload))
				return info, nil
			}
		case layers.Dot11TypeMgmtAction, layers.Dot11TypeMgmtActionNoAck:
			const fixedHeaderLen = 2
			offsetApplied = fixedHeaderLen
			if len(originalPayload) >= fixedHeaderLen {
				iePayload = originalPayload[fixedHeaderLen:]
				// log.Printf("DEBUG_MGMT_PAYLOAD_OFFSET: FrameType: %s, BSSID: %s, Applied %d-byte offset. OriginalPayloadLen: %d, EffectiveIEPayloadLen: %d", dot11.Type.String(), bssidForLog, fixedHeaderLen, len(originalPayload), len(iePayload))
			} else {
				log.Printf("WARN_MGMT_PAYLOAD_OFFSET: FrameType: %s, BSSID: %s, Payload too short for fixed header (expected %d, got %d). No IEs will be parsed.", dot11.Type.String(), bssidForLog, fixedHeaderLen, len(originalPayload))
				return info, nil
			}
		case layers.Dot11TypeMgmtProbeReq:
			iePayload = originalPayload
			offsetApplied = 0
			// log.Printf("DEBUG_MGMT_PAYLOAD_OFFSET: FrameType: %s, BSSID: %s, No offset applied. Using original payload. OriginalPayloadLen: %d", dot11.Type.String(), bssidForLog, len(originalPayload))
		default:
			iePayload = originalPayload
			offsetApplied = 0
			// log.Printf("DEBUG_MGMT_PAYLOAD_OFFSET: FrameType: %s, BSSID: %s, No specific offset applied (default case). Using original payload. OriginalPayloadLen: %d", dot11.Type.String(), bssidForLog, len(originalPayload))
		}
		_ = offsetApplied

		info.SSID = ""
		info.SupportedRates = nil
		info.DSSetChannel = 0
		info.TIM = nil
		info.HTCapabilitiesRaw = nil
		info.VHTCapabilitiesRaw = nil
		info.HECapabilitiesRaw = nil
		info.VHTOperationRaw = nil
		info.RSNRaw = nil
		info.ParsedHTCaps = nil
		info.ParsedVHTCaps = nil

		// log.Printf("DEBUG_MGMT_PAYLOAD_PARSE: FrameType: %s, BSSID: %s, Effective IE Payload Length for parsing: %d", dot11.Type.String(), bssidForLog, len(iePayload))

		currentIEPayload := iePayload
		for len(currentIEPayload) > 0 {
			if len(currentIEPayload) < 2 {
				if len(currentIEPayload) > 0 {
					log.Printf("WARN_IE_PARSE: Trailing data too short for full IE header (ID+Length). FrameType: %s, BSSID: %s. Length: %d. Data: %x.", dot11.Type.String(), bssidForLog, len(currentIEPayload), currentIEPayload)
				}
				break
			}

			ieID := layers.Dot11InformationElementID(currentIEPayload[0])
			ieLength := int(currentIEPayload[1])

			if ieLength < 0 {
				log.Printf("WARN_IE_PARSE: Invalid IE length %d for IE ID %d (Name: %s). FrameType: %s, BSSID: %s. Stopping IE parse for this frame.", ieLength, ieID, ieID.String(), dot11.Type.String(), bssidForLog)
				break
			}

			availableDataForIEContent := len(currentIEPayload) - 2
			if availableDataForIEContent < ieLength {
				log.Printf("WARN_IE_PARSE: Declared IE length (%d) for IE ID %d (Name: %s) exceeds available data for content (%d). FrameType: %s, BSSID: %s. Stopping IE parse for this frame.", ieLength, ieID, ieID.String(), availableDataForIEContent, dot11.Type.String(), bssidForLog)
				break
			}

			ieInfo := currentIEPayload[2 : 2+ieLength]

			// log.Printf("DEBUG_IE_ITERATION: IE ID: %d (Name: %s), Declared Length: %d. FrameType: %s, BSSID: %s", ieID, ieID.String(), ieLength, dot11.Type.String(), bssidForLog)

			switch ieID {
			case layers.Dot11InformationElementIDSSID:
				var ssidContent string
				if ieLength == 0 {
					ssidContent = "<Hidden SSID>"
				} else {
					if utf8.Valid(ieInfo) {
						ssidContent = string(ieInfo)
					} else {
						ssidContent = "<Invalid/Undecodable SSID>" // Mark as invalid if not UTF-8
						log.Printf("WARN_SSID_PARSE: Invalid UTF-8 encoding for SSID IE. BSSID: %s, Length: %d, Hex: %x", bssidForLog, ieLength, ieInfo)
					}
				}
				info.SSID = ssidContent
				// log.Printf("DEBUG_SSID_PARSE: Found SSID IE for BSSID %s. Length: %d, SSID: [%s], Hex: %x", bssidForLog, ieLength, ssidContent, ieInfo) // Log the final result

			case layers.Dot11InformationElementIDRates:
				info.SupportedRates = make([]byte, ieLength)
				copy(info.SupportedRates, ieInfo)

			case layers.Dot11InformationElementIDDSSet:
				if len(ieInfo) > 0 {
					channelVal := ieInfo[0]
					info.DSSetChannel = channelVal
					if info.Channel == 0 && channelVal >= 1 && channelVal <= 14 { // Basic 2.4GHz channel check
						info.Channel = int(channelVal)
					}
				}

			case layers.Dot11InformationElementIDTIM:
				info.TIM = make([]byte, ieLength)
				copy(info.TIM, ieInfo)

			case layers.Dot11InformationElementIDHTCapabilities:
				info.HTCapabilitiesRaw = make([]byte, ieLength)
				copy(info.HTCapabilitiesRaw, ieInfo)

			case layers.Dot11InformationElementIDVHTCapabilities:
				info.VHTCapabilitiesRaw = make([]byte, ieLength)
				copy(info.VHTCapabilitiesRaw, ieInfo)

			case layers.Dot11InformationElementIDVHTOperation:
				info.VHTOperationRaw = make([]byte, ieLength)
				copy(info.VHTOperationRaw, ieInfo)

			// case layers.Dot11InformationElementIDHECapabilities: // Constant might be missing
			// 	info.HECapabilitiesRaw = make([]byte, ieLength)
			// 	copy(info.HECapabilitiesRaw, ieInfo)

			case layers.Dot11InformationElementIDRSNInfo:
				info.RSNRaw = make([]byte, ieLength)
				copy(info.RSNRaw, ieInfo)
			}
			currentIEPayload = currentIEPayload[2+ieLength:]
		}

		// Parse HT Capabilities
		if len(info.HTCapabilitiesRaw) >= 2 { // Minimum length for HT Capabilities Info field
			info.ParsedHTCaps = &HTCapabilityInfo{}
			htCapInfoField := uint16(info.HTCapabilitiesRaw[0]) | (uint16(info.HTCapabilitiesRaw[1]) << 8)
			info.ParsedHTCaps.ChannelWidth40MHz = (htCapInfoField & 0x0002) != 0 // Bit 1
			info.ParsedHTCaps.ShortGI20MHz = (htCapInfoField & 0x0020) != 0      // Bit 5
			info.ParsedHTCaps.ShortGI40MHz = (htCapInfoField & 0x0040) != 0      // Bit 6

			if len(info.HTCapabilitiesRaw) >= 18 { // MCS set is 16 bytes, starting at offset 2
				info.ParsedHTCaps.SupportedMCSSet = make([]byte, 16)
				copy(info.ParsedHTCaps.SupportedMCSSet, info.HTCapabilitiesRaw[2:18])
			}

			if info.ParsedHTCaps.ChannelWidth40MHz {
				info.Bandwidth = "40MHz"
			} else {
				info.Bandwidth = "20MHz"
			}
			if info.ParsedHTCaps.ShortGI20MHz || info.ParsedHTCaps.ShortGI40MHz {
				info.IsShortGI = true // Set from HT Caps if not already set by Radiotap
			}
		}

		// Parse VHT Operation to determine bandwidth (overrides HT if present)
		if len(info.VHTOperationRaw) >= 1 {
			vhtOpChannelWidth := info.VHTOperationRaw[0]
			switch vhtOpChannelWidth {
			case 0: // 20 or 40 MHz
				if info.ParsedHTCaps != nil && info.ParsedHTCaps.ChannelWidth40MHz {
					info.Bandwidth = "40MHz"
				} else {
					info.Bandwidth = "20MHz" // Default if no HT 40MHz
				}
			case 1:
				info.Bandwidth = "80MHz"
			case 2:
				info.Bandwidth = "160MHz"
			case 3:
				info.Bandwidth = "80+80MHz"
			default:
				log.Printf("WARN_VHT_OP: Unknown VHT Operation Channel Width: %d", vhtOpChannelWidth)
				// Keep existing bandwidth or default to 20MHz if nothing else set
				if info.Bandwidth == "" {
					info.Bandwidth = "20MHz"
				}
			}
		}

		// Parse VHT Capabilities
		if len(info.VHTCapabilitiesRaw) >= 12 { // Minimum length for VHT Capabilities IE
			info.ParsedVHTCaps = &VHTCapabilityInfo{}
			// VHT Capability Info field (first 4 bytes)
			vhtCapInfoByte0 := info.VHTCapabilitiesRaw[0]
			vhtCapInfoByte1 := info.VHTCapabilitiesRaw[1]

			info.ParsedVHTCaps.MaxMPDULength = vhtCapInfoByte0 & 0x03                   // Bits 0-1
			info.ParsedVHTCaps.SupportedChannelWidthSet = (vhtCapInfoByte0 & 0x0C) >> 2 // Bits 2-3
			info.ParsedVHTCaps.ShortGI80MHz = (vhtCapInfoByte0 & 0x20) != 0             // Bit 5
			info.ParsedVHTCaps.ShortGI160MHz = (vhtCapInfoByte0 & 0x40) != 0            // Bit 6

			info.ParsedVHTCaps.SUBeamformerCapable = (vhtCapInfoByte1 & 0x01) != 0 // Bit 8 (Byte 1, Bit 0)
			info.ParsedVHTCaps.MUBeamformerCapable = (vhtCapInfoByte1 & 0x08) != 0 // Bit 11 (Byte 1, Bit 3)

			// VHT MCS and NSS Set field (next 8 bytes, offset 4 from start of IE)
			info.ParsedVHTCaps.RxMCSMap = uint16(info.VHTCapabilitiesRaw[4]) | (uint16(info.VHTCapabilitiesRaw[5]) << 8)
			// RxHighestLongGIRate: bits 10-12 of RxMCSMap (which is byte 5, bits 2-4 of VHT MCS Set)
			info.ParsedVHTCaps.RxHighestLongGIRate = (info.ParsedVHTCaps.RxMCSMap >> 10) & 0x0007

			info.ParsedVHTCaps.TxMCSMap = uint16(info.VHTCapabilitiesRaw[8]) | (uint16(info.VHTCapabilitiesRaw[9]) << 8)
			// TxHighestLongGIRate: bits 10-12 of TxMCSMap (byte 9, bits 2-4 of VHT MCS Set)
			info.ParsedVHTCaps.TxHighestLongGIRate = (info.ParsedVHTCaps.TxMCSMap >> 10) & 0x0007

			// If bandwidth wasn't set by VHT Operation, try to infer from VHT Capabilities
			if info.Bandwidth == "" || info.Bandwidth == "20MHz" || info.Bandwidth == "40MHz" { // Only upgrade if not already set to higher by VHT Op
				switch info.ParsedVHTCaps.SupportedChannelWidthSet {
				case 1: // 80MHz
					info.Bandwidth = "80MHz"
				case 2: // 160MHz or 80+80MHz
					info.Bandwidth = "160MHz" // Default to 160 for simplicity
				}
			}
			if info.ParsedVHTCaps.ShortGI80MHz || info.ParsedVHTCaps.ShortGI160MHz {
				info.IsShortGI = true // Set from VHT Caps if not already set
			}
		}
		// Re-evaluate PHY rate after parsing Dot11 and IEs if needed
		// info.PHYRateMbps = getPHYRateMbps(rt, dot11, info.ParsedHTCaps, info.ParsedVHTCaps)

	} else if dot11.Type.MainType() == layers.Dot11TypeData {
		switch dot11.Type {
		case layers.Dot11TypeDataQOSData,
			layers.Dot11TypeDataQOSDataCFAck,
			layers.Dot11TypeDataQOSDataCFPoll,
			// layers.Dot11TypeDataQOSDataCFAckCFPoll, // Constant might be missing
			layers.Dot11TypeDataQOSNull,
			layers.Dot11TypeDataQOSCFPollNoData:
			// layers.Dot11TypeDataQOSCFAckCFPollNoData: // Constant might be missing
			info.IsQoSData = true
		default:
			info.IsQoSData = false
		}
	}

	// Attempt to parse transport layer for payload length
	info.TransportPayloadLength = -1 // Default to -1 (unavailable)
	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			if tcp != nil && tcp.Payload != nil {
				info.TransportPayloadLength = len(tcp.Payload)
			}
		} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			if udp != nil && udp.Payload != nil {
				info.TransportPayloadLength = len(udp.Payload)
			}
		}
	} else if ip6Layer := packet.Layer(layers.LayerTypeIPv6); ip6Layer != nil {
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			if tcp != nil && tcp.Payload != nil {
				info.TransportPayloadLength = len(tcp.Payload)
			}
		} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			if udp != nil && udp.Payload != nil {
				info.TransportPayloadLength = len(udp.Payload)
			}
		}
	}

	// // frameTypeStr := info.FrameType.String()
	// // saStr := "N/A"
	// // if info.SA != nil {
	// // 	saStr = info.SA.String()
	// // }
	// // daStr := "N/A"
	// // if info.DA != nil {
	// // 	daStr = info.DA.String()
	// // }
	// // finalBssidStr := "N/A"
	// // if info.BSSID != nil {
	// // 	finalBssidStr = info.BSSID.String()
	// // }
	// // ssidStr := info.SSID
	// // if ssidStr == "" {
	// // 	ssidStr = "N/A"
	// // }

	// // Log MACDurationID
	// // log.Printf("DEBUG_FRAME_PARSER_MAC_DURATION: FrameType: %s, BSSID: %s, SA: %s, DA: %s, MACDurationID: %d",
	// // 	frameTypeStr, finalBssidStr, saStr, daStr, info.MACDurationID)

	// // Log Transport Layer Info and Payload Length
	// // hasIPv4 := packet.Layer(layers.LayerTypeIPv4) != nil
	// // hasIPv6 := packet.Layer(layers.LayerTypeIPv6) != nil
	// // hasTCP := packet.Layer(layers.LayerTypeTCP) != nil
	// // hasUDP := packet.Layer(layers.LayerTypeUDP) != nil
	// // log.Printf("DEBUG_FRAME_PARSER_TRANSPORT_INFO: FrameType: %s, BSSID: %s, SA: %s, DA: %s, HasIPv4: %t, HasIPv6: %t, HasTCP: %t, HasUDP: %t, TransportPayloadLength: %d",
	// // 	frameTypeStr, finalBssidStr, saStr, daStr, hasIPv4, hasIPv6, hasTCP, hasUDP, info.TransportPayloadLength)

	// // log.Printf("DEBUG_FRAME_PARSER_SUMMARY: Frame Type: %s, BSSID: %s, SA: %s, DA: %s, SSID: [%s], Channel: %d, Signal: %d dBm, Bandwidth: %s",
	// // 	frameTypeStr, finalBssidStr, saStr, daStr, ssidStr, info.Channel, info.SignalStrength, info.Bandwidth)

	return info, nil
}

// CalculateFrameAirtime estimates the airtime of a given 802.11 frame.
// This is a simplified model. A more accurate calculation would consider
// preamble, PLCP header, MAC header, FCS, and any ACK/BlockAck exchanges.
func CalculateFrameAirtime(frameLengthBytes int, phyRateMbps float64, isShortPreamble bool, isShortGI bool) time.Duration {
	if phyRateMbps <= 0 {
		return 0 // Avoid division by zero or negative rates
	}

	// Basic data transmission time
	// Frame length in bits / PHY rate in Mbps = time in microseconds
	dataTxTimeMicroseconds := float64(frameLengthBytes*8) / phyRateMbps

	// Simplified Preamble + PLCP Header time (microseconds)
	// Long Preamble: 192us (at 1 Mbps)
	// Short Preamble: 96us (at 1 Mbps) - Note: gopacket uses 120us for short? Let's stick to standard values.
	preamblePlcpTimeMicroseconds := 192.0
	if isShortPreamble {
		preamblePlcpTimeMicroseconds = 96.0
	}

	// Simplified SIFS (Short Interframe Space)
	sifsMicroseconds := 10.0 // Typical for OFDM PHYs (a, g, n, ac, ax)

	// Guard Interval (GI) impact (simplified)
	giFactor := 1.0
	if isShortGI {
		giFactor = 0.9 // Rough approximation (e.g., 4us vs 3.6us symbol time)
	}

	// Total estimated airtime
	// Model: Preamble/PLCP + DataTxTime + SIFS (for potential ACK)
	// This is highly simplified. Ignores DIFS, Backoff, ACK frame duration, etc.
	totalMicroseconds := (preamblePlcpTimeMicroseconds + dataTxTimeMicroseconds*giFactor) + sifsMicroseconds

	return time.Duration(totalMicroseconds * float64(time.Microsecond))
}
