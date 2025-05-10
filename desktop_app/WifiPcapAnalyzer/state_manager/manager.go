package state_manager

import (
	"WifiPcapAnalyzer/config"       // Import for config.GlobalConfig
	"WifiPcapAnalyzer/frame_parser" // Import for ParsedFrameInfo
	"log"
	"net"
	"sync"
	"time"
	// Import for layers.Dot11Type constants
)

// StateManager holds the current state of all observed BSSs and STAs.
type StateManager struct {
	bssInfos map[string]*BSSInfo // Keyed by BSSID string
	staInfos map[string]*STAInfo // Keyed by STA MAC string
	mutex    sync.RWMutex

	// Pending entries waiting for confirmation (seen once)
	pendingBSSInfos map[string]time.Time // Key: BSSID, Value: First seen time
	pendingSTAInfos map[string]time.Time // Key: STA MAC, Value: First seen time

	// Metrics calculation parameters
	metricsCalcInterval time.Duration // How often to calculate metrics
	maxHistoryPoints    int           // Max number of historical data points
}

// NewStateManager creates a new StateManager.
func NewStateManager(metricsInterval time.Duration, historyPoints int) *StateManager {
	if historyPoints <= 0 {
		historyPoints = 60 // Default to 60 points (e.g., 1 minute if interval is 1s)
	}
	return &StateManager{
		bssInfos:            make(map[string]*BSSInfo),
		staInfos:            make(map[string]*STAInfo),
		pendingBSSInfos:     make(map[string]time.Time),
		pendingSTAInfos:     make(map[string]time.Time),
		metricsCalcInterval: metricsInterval,
		maxHistoryPoints:    historyPoints,
	}
}

// ProcessParsedFrame is the main entry point for updating state based on a parsed frame.
func (sm *StateManager) ProcessParsedFrame(parsedInfo *frame_parser.ParsedFrameInfo) {
	if parsedInfo == nil {
		return
	}
	log.Printf("DEBUG_SM_PROCESS_FRAME_INPUT: Processing ParsedFrameInfo: BSSID=%s, SA=%s, DA=%s, SSID=%s, Type=%s, Signal=%d",
		parsedInfo.BSSID, parsedInfo.SA, parsedInfo.DA, parsedInfo.SSID, parsedInfo.FrameType, parsedInfo.SignalStrength)

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	nowMilli := now.UnixMilli()
	confirmationWindow := 1 * time.Minute // 1 minute confirmation window

	// Calculate data length for metrics (airtime calculation is removed)
	// frameAirtime := frame_parser.CalculateFrameAirtime(parsedInfo.FrameLength, parsedInfo.PHYRateMbps, parsedInfo.IsShortPreamble, parsedInfo.IsShortGI)
	frameDataLength := 0
	// Check WlanFcType for Data type (2)
	if parsedInfo.WlanFcType == 2 { // 2 corresponds to Dot11TypeData
		if parsedInfo.TransportPayloadLength > 0 {
			frameDataLength = parsedInfo.TransportPayloadLength
		} else {
			// Option A: Ignore frame for throughput if TransportPayloadLength is not valid.
			// No bytes are added to frameDataLength, so it remains 0.
		}
	}

	// --- Helper function to handle STA update/creation ---
	handleSTA := func(mac net.HardwareAddr, isSource bool) {
		if mac == nil || !isUnicastMAC(mac) {
			return // Ignore invalid or non-unicast MACs
		}
		macStr := mac.String()

		sta, exists := sm.staInfos[macStr]
		if exists {
			// STA already confirmed, just update LastSeen and potentially signal
			sta.LastSeen = nowMilli
			if parsedInfo.SignalStrength != 0 {
				// Update signal based on source/transmitter context if needed
				// For simplicity, let's update if non-zero, maybe prioritize SA later
				sta.SignalStrength = parsedInfo.SignalStrength
			}
			// Update capabilities if needed (logic omitted for brevity, similar to below)
			log.Printf("DEBUG_SM_UPDATE: STA %s updated. LastSeen: %v, Signal: %d", macStr, time.UnixMilli(sta.LastSeen), sta.SignalStrength)
		} else {
			// STA not confirmed, check pending list
			firstSeenTime, pendingExists := sm.pendingSTAInfos[macStr]
			if pendingExists {
				// Exists in pending list, check time window
				if now.Sub(firstSeenTime) < confirmationWindow {
					// Seen again within the window, confirm it!
					log.Printf("DEBUG_SM_UPDATE: Confirming STA %s (seen again within %v)", macStr, confirmationWindow)
					delete(sm.pendingSTAInfos, macStr) // Remove from pending
					sta = NewSTAInfo(macStr)           // Create new STA
					sta.LastSeen = nowMilli
					if parsedInfo.SignalStrength != 0 {
						sta.SignalStrength = parsedInfo.SignalStrength
					}
					// Update capabilities for the newly created STA
					if parsedInfo.ParsedHTCaps != nil {
						if sta.HTCapabilities == nil {
							sta.HTCapabilities = &HTCapabilities{}
						}
						sta.HTCapabilities.ChannelWidth40MHz = parsedInfo.ParsedHTCaps.ChannelWidth40MHz
						sta.HTCapabilities.ShortGI20MHz = parsedInfo.ParsedHTCaps.ShortGI20MHz
						sta.HTCapabilities.ShortGI40MHz = parsedInfo.ParsedHTCaps.ShortGI40MHz
					}
					if parsedInfo.ParsedVHTCaps != nil {
						if sta.VHTCapabilities == nil {
							sta.VHTCapabilities = &VHTCapabilities{}
						}
						sta.VHTCapabilities.ShortGI80MHz = parsedInfo.ParsedVHTCaps.ShortGI80MHz
						sta.VHTCapabilities.ShortGI160MHz = parsedInfo.ParsedVHTCaps.ShortGI160MHz
						sta.VHTCapabilities.ChannelWidth80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 1)
						sta.VHTCapabilities.ChannelWidth160MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 2)
						sta.VHTCapabilities.ChannelWidth80Plus80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet == 3)
					}
					sm.staInfos[macStr] = sta // Add to confirmed list
					log.Printf("DEBUG_SM_UPDATE: STA %s created/confirmed. Associated BSSID: %s, Signal: %d", sta.MACAddress, sta.AssociatedBSSID, sta.SignalStrength)
				} else {
					// Seen again, but outside the window. Reset the timer.
					log.Printf("DEBUG_SM_UPDATE: Re-pending STA %s (seen again after %v)", macStr, now.Sub(firstSeenTime))
					sm.pendingSTAInfos[macStr] = now // Update timestamp
				}
			} else {
				// First time seeing this STA, add to pending list
				log.Printf("DEBUG_SM_UPDATE: Pending STA %s (first seen)", macStr)
				sm.pendingSTAInfos[macStr] = now
			}
		}
	}
	// --- End Helper function ---

	// Process SA (Source Address)
	handleSTA(parsedInfo.SA, true)

	// Process TA (Transmitter Address) - often the same as SA for STAs, but can differ
	// Avoid processing twice if SA == TA
	if parsedInfo.TA != nil && (parsedInfo.SA == nil || parsedInfo.SA.String() != parsedInfo.TA.String()) {
		handleSTA(parsedInfo.TA, false)
	}

	// --- Original STA logic commented out ---
	/*
		// Update STA info based on SA (Source Address)
		if parsedInfo.SA != nil {
			saStr := parsedInfo.SA.String()
			// Ignore broadcast, multicast, or zero MAC addresses as STA source
			if isUnicastMAC(parsedInfo.SA) {
				sta, exists := sm.staInfos[saStr]
				if !exists {
					sta = NewSTAInfo(saStr)
					sm.staInfos[saStr] = sta
				}
				sta.LastSeen = nowMilli
				// Only update signal if it's a plausible non-zero value
				if parsedInfo.SignalStrength != 0 {
					sta.SignalStrength = parsedInfo.SignalStrength
				}
				// Update STA capabilities
				if parsedInfo.ParsedHTCaps != nil {
					// Create a new HTCapabilities struct for the STA or update existing one
					if sta.HTCapabilities == nil {
						sta.HTCapabilities = &HTCapabilities{}
					}
					sta.HTCapabilities.ChannelWidth40MHz = parsedInfo.ParsedHTCaps.ChannelWidth40MHz
					sta.HTCapabilities.ShortGI20MHz = parsedInfo.ParsedHTCaps.ShortGI20MHz
					sta.HTCapabilities.ShortGI40MHz = parsedInfo.ParsedHTCaps.ShortGI40MHz
					if parsedInfo.ParsedHTCaps.SupportedMCSSet != nil {
						sta.HTCapabilities.SupportedMCSSet = make([]int, len(parsedInfo.ParsedHTCaps.SupportedMCSSet))
						// This assumes ParsedHTCaps.SupportedMCSSet is []byte, needs conversion if it's []int
						// For now, let's assume it's already []int or handle conversion if needed.
						// If it's []byte, we'd iterate and convert. Given the model, it's likely []byte.
						// The model HTCapabilities has []int, ParsedHTCaps has []byte.
						// This part needs careful handling or model alignment.
						// For now, skipping direct copy of MCS set to avoid type mismatch issues without further clarification.
						// log.Printf("DEBUG_STA_HT_CAPS: STA %s HT MCS set to be processed.", saStr)
					}
				}
				if parsedInfo.ParsedVHTCaps != nil {
					if sta.VHTCapabilities == nil {
						sta.VHTCapabilities = &VHTCapabilities{}
					}
					sta.VHTCapabilities.ShortGI80MHz = parsedInfo.ParsedVHTCaps.ShortGI80MHz
					sta.VHTCapabilities.ShortGI160MHz = parsedInfo.ParsedVHTCaps.ShortGI160MHz
					// Map SupportedChannelWidthSet to boolean fields
					sta.VHTCapabilities.ChannelWidth80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 1)
					sta.VHTCapabilities.ChannelWidth160MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 2)
					sta.VHTCapabilities.ChannelWidth80Plus80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet == 3)
					// Copy other VHT fields as needed, e.g., MCS maps, SU/MU beamformer
					// sta.VHTCapabilities.SUBeamformerCapable = parsedInfo.ParsedVHTCaps.SUBeamformerCapable (if field exists in models.VHTCapabilities)
					// sta.VHTCapabilities.MUBeamformerCapable = parsedInfo.ParsedVHTCaps.MUBeamformerCapable (if field exists in models.VHTCapabilities)
				}
			}
		}

		// Update STA info based on TA (Transmitter Address), only if different from SA
		// TA is often the actual device MAC, even in AP-to-STA frames where SA might be the BSSID
		if parsedInfo.TA != nil && (parsedInfo.SA == nil || parsedInfo.TA.String() != parsedInfo.SA.String()) {
			taStr := parsedInfo.TA.String()
			// Ignore broadcast, multicast, or zero MAC addresses as STA transmitter
			if isUnicastMAC(parsedInfo.TA) {
				sta, exists := sm.staInfos[taStr]
				if !exists {
					sta = NewSTAInfo(taStr)
					sm.staInfos[taStr] = sta
				}
				sta.LastSeen = nowMilli
				// Signal strength from TA might be more relevant in some cases (e.g., STA sending)
				// Let's update signal from TA as well if SA didn't provide it or if TA is different
				if parsedInfo.SignalStrength != 0 && sta.SignalStrength == 0 { // Prioritize SA signal if available
					sta.SignalStrength = parsedInfo.SignalStrength
				}
				// Update STA capabilities from TA (less common, but possible if TA is the STA)
				if parsedInfo.ParsedHTCaps != nil {
					if sta.HTCapabilities == nil {
						sta.HTCapabilities = &HTCapabilities{}
					}
					sta.HTCapabilities.ChannelWidth40MHz = parsedInfo.ParsedHTCaps.ChannelWidth40MHz
					sta.HTCapabilities.ShortGI20MHz = parsedInfo.ParsedHTCaps.ShortGI20MHz
					sta.HTCapabilities.ShortGI40MHz = parsedInfo.ParsedHTCaps.ShortGI40MHz
					// Similar MCS set handling as above for SA
				}
				if parsedInfo.ParsedVHTCaps != nil {
					if sta.VHTCapabilities == nil {
						sta.VHTCapabilities = &VHTCapabilities{}
					}
					sta.VHTCapabilities.ShortGI80MHz = parsedInfo.ParsedVHTCaps.ShortGI80MHz
					sta.VHTCapabilities.ShortGI160MHz = parsedInfo.ParsedVHTCaps.ShortGI160MHz
					sta.VHTCapabilities.ChannelWidth80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 1)
					sta.VHTCapabilities.ChannelWidth160MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 2)
					sta.VHTCapabilities.ChannelWidth80Plus80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet == 3)
					// ... copy other VHT fields
				// ... (capability update logic was here) ...
			}
		}

		// Update STA info based on TA (Transmitter Address), only if different from SA
		// TA is often the actual device MAC, even in AP-to-STA frames where SA might be the BSSID
		if parsedInfo.TA != nil && (parsedInfo.SA == nil || parsedInfo.TA.String() != parsedInfo.SA.String()) {
			taStr := parsedInfo.TA.String()
			// Ignore broadcast, multicast, or zero MAC addresses as STA transmitter
			if isUnicastMAC(parsedInfo.TA) {
				sta, exists := sm.staInfos[taStr]
				if !exists {
					sta = NewSTAInfo(taStr)
					sm.staInfos[taStr] = sta
				}
				sta.LastSeen = nowMilli
				// Signal strength from TA might be more relevant in some cases (e.g., STA sending)
				// Let's update signal from TA as well if SA didn't provide it or if TA is different
				if parsedInfo.SignalStrength != 0 && sta.SignalStrength == 0 { // Prioritize SA signal if available
					sta.SignalStrength = parsedInfo.SignalStrength
				}
				// Update STA capabilities from TA (less common, but possible if TA is the STA)
				// ... (capability update logic was here) ...
			}
		}
	*/
	// --- End Original STA logic ---

	// --- BSS Processing ---
	// Check WlanFcType for Management type (0)
	log.Printf("DEBUG_SM_BSS_CHECK_TYPE: Frame BSSID: %s, SA: %s, DA: %s, FrameType: %s, WlanFcType: %d", parsedInfo.BSSID, parsedInfo.SA, parsedInfo.DA, parsedInfo.FrameType, parsedInfo.WlanFcType)
	if parsedInfo.WlanFcType == 0 { // 0 corresponds to Dot11TypeMgmt
		bssidMAC := parsedInfo.BSSID
		log.Printf("DEBUG_SM_BSS_CHECK_MAC: Frame BSSID: %s, bssidMAC: %v, isUnicast: %t", parsedInfo.BSSID, bssidMAC, isUnicastMAC(bssidMAC))
		if bssidMAC != nil && isUnicastMAC(bssidMAC) { // Ensure BSSID is valid and unicast
			bssidStr := bssidMAC.String()

			// Critical check: Do not process or create BSSInfo for broadcast BSSID
			if bssidStr == "ff:ff:ff:ff:ff:ff" {
				// log.Printf("DEBUG_STATE_MANAGER: Ignoring Mgmt frame with broadcast BSSID: %s. SA: %s, DA: %s, FrameType: %s", bssidStr, parsedInfo.SA, parsedInfo.DA, parsedInfo.FrameType.String())
				return // Exit early, do not create/update BSS for ff:ff:ff:ff:ff:ff
			} // Removed broadcast check here as it's covered by isUnicastMAC

			bss, bssExists := sm.bssInfos[bssidStr]

			// --- BSS Creation/Update Logic with Confirmation ---
			// Compare with string representations or WlanFcSubtype for specific Mgmt frames
			isBeaconOrProbeResp := parsedInfo.FrameType == "MgmtBeacon" || parsedInfo.FrameType == "MgmtProbeResp"
			log.Printf("DEBUG_SM_BSS_CHECK_BEACON_PROBE: Frame BSSID: %s, FrameType: %s, isBeaconOrProbeResp: %t", parsedInfo.BSSID, parsedInfo.FrameType, isBeaconOrProbeResp)

			if bssExists {
				// BSS already confirmed, update LastSeen and details if Beacon/ProbeResp
				bss.LastSeen = nowMilli
				if isBeaconOrProbeResp {
					// Update attributes (Signal, SSID, Channel, BW, Security, Caps)
					if parsedInfo.SignalStrength != 0 {
						bss.SignalStrength = parsedInfo.SignalStrength
					}
					if parsedInfo.SSID != "" && parsedInfo.SSID != "[N/A]" {
						if parsedInfo.SSID == "<Hidden SSID>" {
							if bss.SSID == "" {
								bss.SSID = parsedInfo.SSID
							}
						} else {
							bss.SSID = parsedInfo.SSID
						}
					}
					if parsedInfo.Channel != 0 {
						bss.Channel = parsedInfo.Channel
					}
					if parsedInfo.Bandwidth != "" {
						bss.Bandwidth = parsedInfo.Bandwidth
					}
					// Update Capabilities
					if parsedInfo.ParsedHTCaps != nil {
						if bss.HTCapabilities == nil {
							bss.HTCapabilities = &HTCapabilities{}
						}
						bss.HTCapabilities.ChannelWidth40MHz = parsedInfo.ParsedHTCaps.ChannelWidth40MHz
						bss.HTCapabilities.ShortGI20MHz = parsedInfo.ParsedHTCaps.ShortGI20MHz
						bss.HTCapabilities.ShortGI40MHz = parsedInfo.ParsedHTCaps.ShortGI40MHz
					}
					if parsedInfo.ParsedVHTCaps != nil {
						if bss.VHTCapabilities == nil {
							bss.VHTCapabilities = &VHTCapabilities{}
						}
						bss.VHTCapabilities.ShortGI80MHz = parsedInfo.ParsedVHTCaps.ShortGI80MHz
						bss.VHTCapabilities.ShortGI160MHz = parsedInfo.ParsedVHTCaps.ShortGI160MHz
						bss.VHTCapabilities.ChannelWidth80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 1)
						bss.VHTCapabilities.ChannelWidth160MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 2)
						bss.VHTCapabilities.ChannelWidth80Plus80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet == 3)
					}
					// Update Security
					if len(parsedInfo.RSNRaw) > 0 {
						isWPA := false
						for _, rsnElem := range parsedInfo.RSNRaw {
							if rsnElem > 0 {
								isWPA = true
								break
							}
						}
						if isWPA {
							bss.Security = "RSN/WPA2/WPA3"
						}
					} else {
						if bss.Security == "" {
							bss.Security = "Open"
						}
					}
				}
				log.Printf("DEBUG_SM_UPDATE: BSS %s updated. LastSeen: %v, SSID: %s, Signal: %d", bssidStr, time.UnixMilli(bss.LastSeen), bss.SSID, bss.SignalStrength)
				log.Printf("DEBUG_SM_BSS_UPDATE: BSS %s created/updated. SSID: '%s', Channel: %d, Signal: %d, LastSeen: %v", bss.BSSID, bss.SSID, bss.Channel, bss.SignalStrength, time.UnixMilli(bss.LastSeen))
			} else {
				// BSS not confirmed, check pending list (only for Beacon/ProbeResp)
				if isBeaconOrProbeResp {
					firstSeenTime, pendingExists := sm.pendingBSSInfos[bssidStr]
					if pendingExists {
						// Exists in pending, check time window
						if now.Sub(firstSeenTime) < confirmationWindow {
							// Seen again within window, try to confirm
							log.Printf("DEBUG_SM_UPDATE: Confirming BSS %s (seen again within %v)", bssidStr, confirmationWindow)
							delete(sm.pendingBSSInfos, bssidStr) // Remove from pending

							// --- Apply RSSI and Completeness Filters before confirming ---
							minRSSI := config.GlobalConfig.MinBSSCreationRSSI
							log.Printf("DEBUG_SM_BSS_FILTER_RSSI: BSSID: %s, Signal: %d, MinRSSI: %d, Pass: %t", bssidStr, parsedInfo.SignalStrength, minRSSI, parsedInfo.SignalStrength >= minRSSI)
							if parsedInfo.SignalStrength < minRSSI {
								// log.Printf("DEBUG_STATE_MANAGER: Confirmation failed for BSS %s. Signal %d dBm < threshold %d dBm.", bssidStr, parsedInfo.SignalStrength, minRSSI)
							} else {
								isSsidMissing := (parsedInfo.SSID == "" || parsedInfo.SSID == "[N/A]" || parsedInfo.SSID == "<Hidden SSID>" || parsedInfo.SSID == "<Invalid SSID Encoding>")
								isSecurityMissing := len(parsedInfo.RSNRaw) == 0
								areCapsMissing := parsedInfo.ParsedHTCaps == nil && parsedInfo.ParsedVHTCaps == nil
								passCompleteness := !(isSsidMissing && isSecurityMissing && areCapsMissing)
								log.Printf("DEBUG_SM_BSS_FILTER_COMPLETE: BSSID: %s, SSIDMissing: %t, SecurityMissing: %t, CapsMissing: %t, Pass: %t", bssidStr, isSsidMissing, isSecurityMissing, areCapsMissing, passCompleteness)
								if !passCompleteness {
									log.Printf("WARN_STATE_MANAGER: Confirmation failed for BSS %s due to severely incomplete info.", bssidStr)
								} else {
									// Passed filters, create and add to confirmed list
									bss = NewBSSInfo(bssidStr)
									bss.LastSeen = nowMilli
									// Populate initial data (same logic as update block above)
									if parsedInfo.SignalStrength != 0 {
										bss.SignalStrength = parsedInfo.SignalStrength
									}
									if parsedInfo.SSID != "" && parsedInfo.SSID != "[N/A]" {
										if parsedInfo.SSID == "<Hidden SSID>" {
											if bss.SSID == "" {
												bss.SSID = parsedInfo.SSID
											}
										} else {
											bss.SSID = parsedInfo.SSID
										}
									}
									if parsedInfo.Channel != 0 {
										bss.Channel = parsedInfo.Channel
									}
									if parsedInfo.Bandwidth != "" {
										bss.Bandwidth = parsedInfo.Bandwidth
									}
									if parsedInfo.ParsedHTCaps != nil {
										if bss.HTCapabilities == nil {
											bss.HTCapabilities = &HTCapabilities{}
										}
										bss.HTCapabilities.ChannelWidth40MHz = parsedInfo.ParsedHTCaps.ChannelWidth40MHz
										bss.HTCapabilities.ShortGI20MHz = parsedInfo.ParsedHTCaps.ShortGI20MHz
										bss.HTCapabilities.ShortGI40MHz = parsedInfo.ParsedHTCaps.ShortGI40MHz
									}
									if parsedInfo.ParsedVHTCaps != nil {
										if bss.VHTCapabilities == nil {
											bss.VHTCapabilities = &VHTCapabilities{}
										}
										bss.VHTCapabilities.ShortGI80MHz = parsedInfo.ParsedVHTCaps.ShortGI80MHz
										bss.VHTCapabilities.ShortGI160MHz = parsedInfo.ParsedVHTCaps.ShortGI160MHz
										bss.VHTCapabilities.ChannelWidth80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 1)
										bss.VHTCapabilities.ChannelWidth160MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 2)
										bss.VHTCapabilities.ChannelWidth80Plus80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet == 3)
									}
									if len(parsedInfo.RSNRaw) > 0 {
										isWPA := false
										for _, rsnElem := range parsedInfo.RSNRaw {
											if rsnElem > 0 {
												isWPA = true
												break
											}
										}
										if isWPA {
											bss.Security = "RSN/WPA2/WPA3"
										}
									} else {
										if bss.Security == "" {
											bss.Security = "Open"
										}
									}
									sm.bssInfos[bssidStr] = bss // Add to confirmed map
									log.Printf("DEBUG_SM_UPDATE: BSS %s created/confirmed. SSID: %s, Channel: %d, Signal: %d", bss.BSSID, bss.SSID, bss.Channel, bss.SignalStrength)
									log.Printf("DEBUG_SM_BSS_UPDATE: BSS %s created/updated. SSID: '%s', Channel: %d, Signal: %d, LastSeen: %v", bss.BSSID, bss.SSID, bss.Channel, bss.SignalStrength, time.UnixMilli(bss.LastSeen))
									// log.Printf("DEBUG_SM_BSS_UPDATE: BSS %s created/updated. SSID: %s, Channel: %d, Signal: %d", bss.BSSID, bss.SSID, bss.Channel, bss.SignalStrength) // Original log, will be replaced by the more detailed one below
								}
							}
						} else {
							// Seen again, but outside window. Reset timer.
							log.Printf("DEBUG_SM_UPDATE: Re-pending BSS %s (seen again after %v)", bssidStr, now.Sub(firstSeenTime))
							sm.pendingBSSInfos[bssidStr] = now // Update timestamp
						}
					} else {
						// First time seeing this BSS, add to pending list
						log.Printf("DEBUG_SM_UPDATE: Pending BSS %s (first seen)", bssidStr)
						sm.pendingBSSInfos[bssidStr] = now
					}
				}
			} // End if !bssExists

			// --- Association logic (needs to check confirmed BSS/STA) ---
			// Re-fetch bss in case it was just confirmed above
			bss, bssExists = sm.bssInfos[bssidStr] // Check confirmed list now

			if bssExists { // Proceed with association logic only if BSS is confirmed
				switch parsedInfo.FrameType { // Compare with string representations
				case "MgmtAssocReq", "MgmtReassocReq":
					staMAC := parsedInfo.SA.String()
					// Associate only if STA is also confirmed
					if sta, staExists := sm.staInfos[staMAC]; staExists {
						sta.AssociatedBSSID = bssidStr
						bss.AssociatedSTAs[staMAC] = sta
					}
				case "MgmtAssocResp", "MgmtReassocResp":
					staMAC := parsedInfo.DA.String()
					// Associate only if STA is also confirmed
					if sta, staExists := sm.staInfos[staMAC]; staExists {
						sta.AssociatedBSSID = bssidStr // BSSID is the SA in Resp frames
						bss.AssociatedSTAs[staMAC] = sta
					}
				case "MgmtDisassoc", "MgmtDeauth":
					if parsedInfo.SA != nil && parsedInfo.DA != nil {
						saStr := parsedInfo.SA.String()
						daStr := parsedInfo.DA.String()
						// Check if SA is the STA (must be confirmed)
						if sta, staExists := sm.staInfos[saStr]; staExists && daStr == bssidStr {
							if sta.AssociatedBSSID == bssidStr {
								delete(bss.AssociatedSTAs, saStr)
								sta.AssociatedBSSID = ""
							}
							// Check if DA is the STA (must be confirmed)
						} else if sta, staExists := sm.staInfos[daStr]; staExists && saStr == bssidStr {
							if sta.AssociatedBSSID == bssidStr {
								delete(bss.AssociatedSTAs, daStr)
								sta.AssociatedBSSID = ""
							}
						}
					}
				}
			} // End if bssExists (for association logic)
		} // End if bssidMAC != nil && isUnicastMAC
	} // End if Mgmt frame

	// --- Data Frame Association Logic (needs confirmed BSS/STA) ---
	if parsedInfo.WlanFcType == 2 { // 2 corresponds to Dot11TypeData
		staMAC := ""
		apMAC := ""

		// Infer STA/AP based on TA/RA and confirmed BSS list
		if parsedInfo.TA != nil && isUnicastMAC(parsedInfo.TA) {
			taStr := parsedInfo.TA.String()
			if _, isBSS := sm.bssInfos[taStr]; isBSS { // TA is a known AP
				apMAC = taStr
				if parsedInfo.RA != nil && isUnicastMAC(parsedInfo.RA) {
					staMAC = parsedInfo.RA.String()
				}
			} else { // Assume TA is the STA
				staMAC = taStr
				if parsedInfo.RA != nil && isUnicastMAC(parsedInfo.RA) {
					if _, isBSS_RA := sm.bssInfos[parsedInfo.RA.String()]; isBSS_RA {
						apMAC = parsedInfo.RA.String()
					}
				}
			}
		}

		// Proceed only if both STA and AP are inferred and are confirmed
		if staMAC != "" && apMAC != "" {
			sta, staExists := sm.staInfos[staMAC]
			bss, bssExists := sm.bssInfos[apMAC]

			if staExists && bssExists {
				// Both are confirmed, update association and LastSeen
				sta.LastSeen = nowMilli
				bss.LastSeen = nowMilli
				if sta.AssociatedBSSID != apMAC {
					// Remove from old BSS if necessary
					if sta.AssociatedBSSID != "" {
						if oldBss, oldBssExists := sm.bssInfos[sta.AssociatedBSSID]; oldBssExists {
							delete(oldBss.AssociatedSTAs, staMAC)
						}
					}
					sta.AssociatedBSSID = apMAC
					bss.AssociatedSTAs[staMAC] = sta
					// log.Printf("DEBUG_STATE_MANAGER: Associated confirmed STA %s with confirmed BSS %s based on data frame.", staMAC, apMAC)
				} else {
					// Already associated, just ensure STA is in the map (should be)
					if _, ok := bss.AssociatedSTAs[staMAC]; !ok {
						bss.AssociatedSTAs[staMAC] = sta
					}
				}
				// Update STA signal from data frame
				if parsedInfo.SignalStrength != 0 {
					sta.SignalStrength = parsedInfo.SignalStrength
				}
			} else {
				// Log if association cannot be made because one/both are not confirmed
				// log.Printf("DEBUG_STATE_MANAGER: Ignored data frame association. STA Confirmed: %v, BSS Confirmed: %v", staExists, bssExists)
			}
		} else if staMAC != "" {
			// Update last seen for confirmed STA even if AP is unknown/invalid
			if sta, staExists := sm.staInfos[staMAC]; staExists {
				sta.LastSeen = nowMilli
			}
		}
	} // End if Data frame

	// --- Original Data Frame Logic Commented Out ---
	/*
		if parsedInfo.FrameType.MainType() == layers.Dot11TypeData {
			staMAC := ""
			apMAC := ""

			if parsedInfo.TA != nil {
				taStr := parsedInfo.TA.String()
				if _, isBSS := sm.bssInfos[taStr]; isBSS {
					apMAC = taStr
					if parsedInfo.RA != nil {
						if isUnicastMAC(parsedInfo.RA) {
							staMAC = parsedInfo.RA.String()
						} else {
							log.Printf("DEBUG_STATE_MANAGER: Data frame RA %s is not unicast, not considered for STA.", parsedInfo.RA.String())
							// staMAC will remain empty or its previous value,
							// and will likely be filtered by subsequent staMAC != "" checks.
						}
					}
				} else {
					staMAC = taStr
					if parsedInfo.RA != nil {
						if _, isBSS_RA := sm.bssInfos[parsedInfo.RA.String()]; isBSS_RA {
							apMAC = parsedInfo.RA.String()
						}
					}
				}
			}

			// Ensure both inferred STA and AP MACs are valid and not broadcast/zero
			if staMAC != "" && apMAC != "" && staMAC != "ff:ff:ff:ff:ff:ff" && staMAC != "00:00:00:00:00:00" && apMAC != "ff:ff:ff:ff:ff:ff" && apMAC != "00:00:00:00:00:00" {
				sta, staExists := sm.staInfos[staMAC]
				if !staExists {
					sta = NewSTAInfo(staMAC)
					sm.staInfos[staMAC] = sta
				}
				sta.LastSeen = nowMilli
				// Update signal strength if available from the data frame context
				if parsedInfo.SignalStrength != 0 {
					sta.SignalStrength = parsedInfo.SignalStrength
				}

				// Only update association if it's different or not set, to avoid flapping
				if sta.AssociatedBSSID != apMAC {
					// Remove from old BSS if previously associated to a different BSS
					if sta.AssociatedBSSID != "" && sta.AssociatedBSSID != apMAC {
						if oldBss, bssExists := sm.bssInfos[sta.AssociatedBSSID]; bssExists {
							delete(oldBss.AssociatedSTAs, staMAC)
						}
					}
					sta.AssociatedBSSID = apMAC
				}

				bss, bssExists := sm.bssInfos[apMAC]
				// --- Modification: Do NOT create new BSS from data frames ---
				if bssExists {
					// Only update existing BSS and association if BSS is already known
					bss.LastSeen = nowMilli
					if _, ok := bss.AssociatedSTAs[staMAC]; !ok {
						bss.AssociatedSTAs[staMAC] = sta
						log.Printf("DEBUG_STATE_MANAGER: Associated STA %s with existing BSS %s based on data frame.", staMAC, apMAC)
					}
				} else {
					log.Printf("DEBUG_STATE_MANAGER: Ignored potential STA %s association with unknown BSS %s based on data frame.", staMAC, apMAC)
				}
				// --- End Modification ---
			} else if staMAC != "" && staMAC != "ff:ff:ff:ff:ff:ff" && staMAC != "00:00:00:00:00:00" {
				// Update last seen for STA even if AP is unknown/invalid
				sta, staExists := sm.staInfos[staMAC]
				if !staExists {
					// Optionally create STA here? Or only create STAs when seen with a known BSS or via Mgmt frames?
					// Let's be strict for now: only update LastSeen if STA already exists.
					// sta = NewSTAInfo(staMAC)
					// sm.staInfos[staMAC] = sta
				} else {
					sta.LastSeen = nowMilli
				}
			}
		}
	*/
	// --- End Original Data Frame Logic ---

	// Accumulate metrics for confirmed BSS and STA
	if parsedInfo.BSSID != nil {
		bssidStr := parsedInfo.BSSID.String()
		if bss, exists := sm.bssInfos[bssidStr]; exists && bss != nil {
			// bss.totalAirtime += frameAirtime // Airtime calculation removed
			if frameDataLength > 0 {
				bss.totalTxBytes += int64(frameDataLength)
				// log.Printf("DEBUG_METRIC_ACCUM: BSSID: %s, Added Bytes: %d, Total Bytes: %d for Throughput", bssidStr, frameDataLength, bss.totalTxBytes)
			}

			// Accumulate MACDurationID
			// Check if the frame is a Control frame and PS-Poll subtype
			isCtrlPSPoll := false
			// Check WlanFcType for Control (1) and WlanFcSubtype for PS-Poll (10)
			if parsedInfo.WlanFcType == 1 && parsedInfo.WlanFcSubtype == 10 { // 10 corresponds to SubtypeCtrlPSPoll
				isCtrlPSPoll = true
				// log.Printf("DEBUG_NAV_SKIP: Skipping NAV accumulation for PS-Poll frame. BSSID: %s, SA: %s, DurationID: %d", bssidStr, parsedInfo.SA, parsedInfo.MACDurationID)
			}

			if !isCtrlPSPoll {
				bss.AccumulatedNavMicroseconds += uint64(parsedInfo.MACDurationID)
				// log.Printf("DEBUG_METRIC_ACCUM: BSSID: %s, Added NAV Microseconds: %d, Total NAV Microseconds: %d for Channel Utilization", bssidStr, parsedInfo.MACDurationID, bss.AccumulatedNavMicroseconds)
			}
		}
	}

	// Accumulate for specific STA if SA is confirmed
	if parsedInfo.SA != nil && isUnicastMAC(parsedInfo.SA) {
		saStr := parsedInfo.SA.String()
		if sta, exists := sm.staInfos[saStr]; exists && sta != nil {
			// sta.totalAirtime += frameAirtime // Airtime calculation removed
			if frameDataLength > 0 {
				originalUplink := sta.totalUplinkBytes
				originalDownlink := sta.totalDownlinkBytes
				// Determine uplink/downlink based on confirmed association
				if sta.AssociatedBSSID != "" {
					if parsedInfo.DA != nil && parsedInfo.DA.String() == sta.AssociatedBSSID {
						sta.totalUplinkBytes += int64(frameDataLength)
						// log.Printf("DEBUG_METRIC_ACCUM: STA: %s (to BSS: %s), Added Uplink Bytes: %d, Total Uplink: %d", saStr, sta.AssociatedBSSID, frameDataLength, sta.totalUplinkBytes)
					} else if parsedInfo.SA.String() == sta.AssociatedBSSID { // Frame from AP to STA
						sta.totalDownlinkBytes += int64(frameDataLength)
						// log.Printf("DEBUG_METRIC_ACCUM: STA: %s (from BSS: %s), Added Downlink Bytes: %d, Total Downlink: %d", saStr, sta.AssociatedBSSID, frameDataLength, sta.totalDownlinkBytes)
					} else {
						// Could be broadcast/multicast from STA, count as uplink? Or ignore?
						// For now, let's count as uplink if DA is not the associated AP.
						sta.totalUplinkBytes += int64(frameDataLength)
						// log.Printf("DEBUG_METRIC_ACCUM: STA: %s (DA not BSS), Added Uplink Bytes: %d, Total Uplink: %d", saStr, frameDataLength, sta.totalUplinkBytes)
					}
				} else {
					// Unassociated STA sending data, count as uplink
					sta.totalUplinkBytes += int64(frameDataLength)
					// log.Printf("DEBUG_METRIC_ACCUM: STA: %s (Unassociated), Added Uplink Bytes: %d, Total Uplink: %d", saStr, frameDataLength, sta.totalUplinkBytes)
				}
				if sta.totalUplinkBytes != originalUplink || sta.totalDownlinkBytes != originalDownlink {
					// This log is redundant if the specific UL/DL logs above are active.
					// log.Printf("DEBUG_METRIC_ACCUM_STA_BYTES: STA: %s, Added Bytes: %d. Uplink: %d -> %d, Downlink: %d -> %d", saStr, frameDataLength, originalUplink, sta.totalUplinkBytes, originalDownlink, sta.totalDownlinkBytes)
				}
			}
		}
	}
	// Accumulate airtime for TA if different and confirmed (Airtime calculation removed)
	/*
		if parsedInfo.TA != nil && isUnicastMAC(parsedInfo.TA) && (parsedInfo.SA == nil || parsedInfo.TA.String() != parsedInfo.SA.String()) {
			taStr := parsedInfo.TA.String()
			if sta, exists := sm.staInfos[taStr]; exists && sta != nil {
				// sta.totalAirtime += frameAirtime // TA also contributes to airtime
			}
		}
	*/

	// log.Printf("DEBUG_STATE_MANAGER: BSS Count: %d, STA Count: %d, Pending BSS: %d, Pending STA: %d",
	// 	len(sm.bssInfos), len(sm.staInfos), len(sm.pendingBSSInfos), len(sm.pendingSTAInfos))
}

// PeriodicallyCalculateMetrics calculates and updates metrics for all confirmed BSSs and STAs.
func (sm *StateManager) PeriodicallyCalculateMetrics() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	// log.Printf("DEBUG_METRIC_CALC_PERIODIC_START: Initiating periodic metrics calculation. CurrentTime: %s, LastCalcInterval: %v", now.Format(time.RFC3339), sm.metricsCalcInterval)
	calculationWindowSeconds := sm.metricsCalcInterval.Seconds()
	if calculationWindowSeconds <= 0 {
		log.Printf("WARN_METRIC_CALC_PERIODIC: calculationWindowSeconds is <= 0 (%f), defaulting to 1.0s", calculationWindowSeconds)
		calculationWindowSeconds = 1.0 // Avoid division by zero, default to 1 second
	}

	for bssID, bss := range sm.bssInfos {
		if bss == nil {
			log.Printf("WARN_METRIC_CALC_PERIODIC: Encountered nil BSS for BSSID: %s, skipping.", bssID)
			continue
		}
		// log.Printf("DEBUG_METRIC_CALC_BSS_PRE: BSSID: %s, LastCalcTime: %s, AccumulatedNavMicroseconds: %d, TotalTxBytes: %d", bssID, bss.lastCalcTime.Format(time.RFC3339), bss.AccumulatedNavMicroseconds, bss.totalTxBytes)

		// originalChannelUtilization := bss.ChannelUtilization
		// originalThroughput := bss.Throughput // Commented out as it's unused after commenting logs

		if bss.lastCalcTime.IsZero() {
			// log.Printf("DEBUG_METRIC_CALC_BSS_INIT: BSSID: %s, First calculation cycle (lastCalcTime is zero). Setting metrics to 0.", bssID)
			bss.ChannelUtilization = 0
			bss.Throughput = 0
		} else {
			elapsed := now.Sub(bss.lastCalcTime).Seconds()
			// log.Printf("DEBUG_METRIC_CALC_BSS_ELAPSED: BSSID: %s, Time since last calculation: %.2fs", bssID, elapsed)
			if elapsed > 0 { // Technically, calculationWindowSeconds is used, but elapsed confirms activity.
				totalWindowMicroseconds := calculationWindowSeconds * 1_000_000
				if totalWindowMicroseconds > 0 {
					bss.ChannelUtilization = (float64(bss.AccumulatedNavMicroseconds) / totalWindowMicroseconds) * 100
				} else {
					bss.ChannelUtilization = 0
				}

				if bss.ChannelUtilization < 0 {
					bss.ChannelUtilization = 0
				}
				if bss.ChannelUtilization > 100.0 {
					bss.ChannelUtilization = 100.0
				}
				bss.Throughput = int64(float64(bss.totalTxBytes*8) / calculationWindowSeconds)
			} else {
				// log.Printf("DEBUG_METRIC_CALC_BSS_NO_ELAPSED: BSSID: %s, Elapsed time is not positive (%.2fs). Setting metrics to 0 for this cycle.", bssID, elapsed)
				bss.ChannelUtilization = 0
				bss.Throughput = 0
			}
		}
		// log.Printf("DEBUG_METRIC_CALC_BSS_POST: BSSID: %s, Calculated ChannelUtilization: %.2f%% (was %.2f%%), Throughput: %d bps (was %d bps)", bssID, bss.ChannelUtilization, originalChannelUtilization, bss.Throughput, originalThroughput)
		// log.Printf("DEBUG_METRIC_UPDATE_BSS: Updating BSS %s: ChannelUtil=%.2f, Throughput=%d", bssID, bss.ChannelUtilization, bss.Throughput)

		bss.HistoricalChannelUtilization = append(bss.HistoricalChannelUtilization, bss.ChannelUtilization)
		if len(bss.HistoricalChannelUtilization) > sm.maxHistoryPoints {
			bss.HistoricalChannelUtilization = bss.HistoricalChannelUtilization[1:]
		}
		bss.HistoricalThroughput = append(bss.HistoricalThroughput, bss.Throughput)
		if len(bss.HistoricalThroughput) > sm.maxHistoryPoints {
			bss.HistoricalThroughput = bss.HistoricalThroughput[1:]
		}

		bss.totalAirtime = 0
		bss.totalTxBytes = 0
		bss.AccumulatedNavMicroseconds = 0
		bss.lastCalcTime = now
	}

	for staMAC, sta := range sm.staInfos {
		if sta == nil {
			log.Printf("WARN_METRIC_CALC_PERIODIC: Encountered nil STA for MAC: %s, skipping.", staMAC)
			continue
		}
		// log.Printf("DEBUG_METRIC_CALC_STA_PRE: STA: %s, LastCalcTime: %s, TotalAirtime: %v, TotalUplinkBytes: %d, TotalDownlinkBytes: %d", staMAC, sta.lastCalcTime.Format(time.RFC3339), sta.totalAirtime, sta.totalUplinkBytes, sta.totalDownlinkBytes)

		// originalSTAChannelUtilization := sta.ChannelUtilization
		// originalSTAUplinkThroughput := sta.UplinkThroughput // Commented out as it's unused after commenting logs
		// originalSTADownlinkThroughput := sta.DownlinkThroughput // Commented out as it's unused after commenting logs

		if sta.lastCalcTime.IsZero() {
			// log.Printf("DEBUG_METRIC_CALC_STA_INIT: STA: %s, First calculation cycle. Setting metrics to 0.", staMAC)
			sta.ChannelUtilization = 0
			sta.UplinkThroughput = 0
			sta.DownlinkThroughput = 0
		} else {
			elapsed := now.Sub(sta.lastCalcTime).Seconds()
			// log.Printf("DEBUG_METRIC_CALC_STA_ELAPSED: STA: %s, Time since last calculation: %.2fs", staMAC, elapsed)
			if elapsed > 0 {
				// STA Channel Utilization - still using old totalAirtime. Consider changing to MACDurationID if applicable for STAs.
				// For now, log the existing calculation.
				sta.ChannelUtilization = (sta.totalAirtime.Seconds() / calculationWindowSeconds) * 100
				if sta.ChannelUtilization < 0 {
					sta.ChannelUtilization = 0
				}
				if sta.ChannelUtilization > 100.0 {
					sta.ChannelUtilization = 100.0
				}
				sta.UplinkThroughput = int64(float64(sta.totalUplinkBytes*8) / calculationWindowSeconds)
				sta.DownlinkThroughput = int64(float64(sta.totalDownlinkBytes*8) / calculationWindowSeconds)
			} else {
				// log.Printf("DEBUG_METRIC_CALC_STA_NO_ELAPSED: STA: %s, Elapsed time is not positive (%.2fs). Setting metrics to 0.", staMAC, elapsed)
				sta.ChannelUtilization = 0
				sta.UplinkThroughput = 0
				sta.DownlinkThroughput = 0
			}
		}
		// log.Printf("DEBUG_METRIC_CALC_STA_POST: STA: %s, Calculated CU: %.2f%% (was %.2f%%), UL: %d bps (was %d), DL: %d bps (was %d)", staMAC, sta.ChannelUtilization, originalSTAChannelUtilization, sta.UplinkThroughput, originalSTAUplinkThroughput, sta.DownlinkThroughput, originalSTADownlinkThroughput)
		// log.Printf("DEBUG_METRIC_UPDATE_STA: Updating STA %s: ChannelUtil=%.2f, UplinkTput=%d, DownlinkTput=%d", staMAC, sta.ChannelUtilization, sta.UplinkThroughput, sta.DownlinkThroughput)

		sta.HistoricalChannelUtilization = append(sta.HistoricalChannelUtilization, sta.ChannelUtilization)
		if len(sta.HistoricalChannelUtilization) > sm.maxHistoryPoints {
			sta.HistoricalChannelUtilization = sta.HistoricalChannelUtilization[1:]
		}
		sta.HistoricalUplinkThroughput = append(sta.HistoricalUplinkThroughput, sta.UplinkThroughput)
		if len(sta.HistoricalUplinkThroughput) > sm.maxHistoryPoints {
			sta.HistoricalUplinkThroughput = sta.HistoricalUplinkThroughput[1:]
		}
		sta.HistoricalDownlinkThroughput = append(sta.HistoricalDownlinkThroughput, sta.DownlinkThroughput)
		if len(sta.HistoricalDownlinkThroughput) > sm.maxHistoryPoints {
			sta.HistoricalDownlinkThroughput = sta.HistoricalDownlinkThroughput[1:]
		}

		sta.totalAirtime = 0
		sta.totalUplinkBytes = 0
		sta.totalDownlinkBytes = 0
		sta.lastCalcTime = now
	}
	// log.Printf("DEBUG_METRIC_CALC_PERIODIC_END: Metrics calculation finished for %d BSSs and %d STAs.", len(sm.bssInfos), len(sm.staInfos))
}

// GetSnapshot returns a deep copy of the current BSS and STA information.
// Only includes confirmed entries.
func (sm *StateManager) GetSnapshot() Snapshot {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	bssList := make([]*BSSInfo, 0, len(sm.bssInfos))
	for bssidKey, bssOriginal := range sm.bssInfos {
		if bssOriginal == nil {
			log.Printf("WARN_SNAPSHOT: Skipping nil BSS in main map for BSSID: %s", bssidKey)
			continue
		}
		bssCopy := *bssOriginal
		bssCopy.AssociatedSTAs = make(map[string]*STAInfo)
		bssCopy.HistoricalChannelUtilization = append([]float64(nil), bssOriginal.HistoricalChannelUtilization...)
		bssCopy.HistoricalThroughput = append([]int64(nil), bssOriginal.HistoricalThroughput...)

		// log.Printf("DEBUG_SNAPSHOT_BSS: BSSID: %s, SSID: %s, ChannelUtil: %.2f, Throughput: %d, NumAssocSTAsInOrig: %d",
		// 	bssCopy.BSSID, bssCopy.SSID, bssCopy.ChannelUtilization, bssCopy.Throughput, len(bssOriginal.AssociatedSTAs))

		for staMAC, _ := range bssOriginal.AssociatedSTAs {
			if mainSta, mainStaExists := sm.staInfos[staMAC]; mainStaExists && mainSta != nil {
				staCopyForBss := *mainSta
				staCopyForBss.HistoricalChannelUtilization = append([]float64(nil), mainSta.HistoricalChannelUtilization...)
				staCopyForBss.HistoricalUplinkThroughput = append([]int64(nil), mainSta.HistoricalUplinkThroughput...)
				staCopyForBss.HistoricalDownlinkThroughput = append([]int64(nil), mainSta.HistoricalDownlinkThroughput...)
				if _, bssStillExists := sm.bssInfos[staCopyForBss.AssociatedBSSID]; !bssStillExists && staCopyForBss.AssociatedBSSID != "" {
					staCopyForBss.AssociatedBSSID = ""
				}
				bssCopy.AssociatedSTAs[staMAC] = &staCopyForBss
				// log.Printf("DEBUG_SNAPSHOT_BSS_STA: Associated STA %s to BSS %s in snapshot. STA CU: %.2f, UL: %d, DL: %d", staMAC, bssCopy.BSSID, staCopyForBss.ChannelUtilization, staCopyForBss.UplinkThroughput, staCopyForBss.DownlinkThroughput)
			} else {
				log.Printf("WARN_SNAPSHOT: STA %s associated with BSS %s not found in main STA list or is nil.", staMAC, bssOriginal.BSSID)
			}
		}
		bssList = append(bssList, &bssCopy)
	}

	staList := make([]*STAInfo, 0, len(sm.staInfos))
	for staMAC, staOriginal := range sm.staInfos {
		if staOriginal == nil {
			log.Printf("WARN_SNAPSHOT: Skipping nil STA in main map for MAC: %s", staMAC)
			continue
		}
		staCopy := *staOriginal
		staCopy.HistoricalChannelUtilization = append([]float64(nil), staOriginal.HistoricalChannelUtilization...)
		staCopy.HistoricalUplinkThroughput = append([]int64(nil), staOriginal.HistoricalUplinkThroughput...)
		staCopy.HistoricalDownlinkThroughput = append([]int64(nil), staOriginal.HistoricalDownlinkThroughput...)

		if staCopy.AssociatedBSSID != "" {
			if _, bssExists := sm.bssInfos[staCopy.AssociatedBSSID]; !bssExists {
				staCopy.AssociatedBSSID = ""
			}
		}
		staList = append(staList, &staCopy)
		// log.Printf("DEBUG_SNAPSHOT_STA: STA: %s, AssociatedBSSID: %s, CU: %.2f, UL: %d, DL: %d",
		// 	staCopy.MACAddress, staCopy.AssociatedBSSID, staCopy.ChannelUtilization, staCopy.UplinkThroughput, staCopy.DownlinkThroughput)
	}
	log.Printf("DEBUG_SM_EVENT_EMIT: Preparing state snapshot. BSS count: %d, STA count: %d", len(bssList), len(staList))
	return Snapshot{BSSs: bssList, STAs: staList}
}

func (sm *StateManager) PruneOldEntries(timeout time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	nowMilli := now.UnixMilli()
	confirmationWindow := 1 * time.Minute                         // Use the same window for pruning pending
	pendingPruneTimeout := confirmationWindow + (1 * time.Minute) // Prune pending if not seen again after 2 mins total

	// Prune confirmed BSSs
	for bssidStr, bss := range sm.bssInfos {
		if (nowMilli - bss.LastSeen) > timeout.Milliseconds() {
			log.Printf("Pruning old BSS: %s (Last seen: %v)", bssidStr, time.UnixMilli(bss.LastSeen))
			// Remove BSS from associated STAs
			for staMAC := range bss.AssociatedSTAs {
				if sta, exists := sm.staInfos[staMAC]; exists {
					if sta.AssociatedBSSID == bssidStr {
						sta.AssociatedBSSID = ""
						log.Printf("STA %s unassociated due to BSS %s pruning.", staMAC, bssidStr)
					}
				}
			}
			delete(sm.bssInfos, bssidStr)
		}
	}

	// Prune confirmed STAs
	for staMAC, sta := range sm.staInfos {
		if (nowMilli - sta.LastSeen) > timeout.Milliseconds() {
			log.Printf("Pruning old STA: %s (Last seen: %v)", staMAC, time.UnixMilli(sta.LastSeen))
			// Remove STA from associated BSS
			if sta.AssociatedBSSID != "" {
				if bss, exists := sm.bssInfos[sta.AssociatedBSSID]; exists {
					delete(bss.AssociatedSTAs, staMAC)
					log.Printf("STA %s removed from BSS %s's association list due to STA pruning.", staMAC, sta.AssociatedBSSID)
				}
			}
			delete(sm.staInfos, staMAC)
		}
	}

	// Prune pending BSSs that haven't been confirmed
	for bssidStr, firstSeenTime := range sm.pendingBSSInfos {
		if now.Sub(firstSeenTime) > pendingPruneTimeout { // Prune if pending for too long
			log.Printf("Pruning pending BSS %s (First seen: %v, timed out)", bssidStr, firstSeenTime)
			delete(sm.pendingBSSInfos, bssidStr)
		}
	}

	// Prune pending STAs that haven't been confirmed
	for staMAC, firstSeenTime := range sm.pendingSTAInfos {
		if now.Sub(firstSeenTime) > pendingPruneTimeout { // Prune if pending for too long
			log.Printf("Pruning pending STA %s (First seen: %v, timed out)", staMAC, firstSeenTime)
			delete(sm.pendingSTAInfos, staMAC)
		}
	}
}

/*
				if isBeaconOrProbeResp {
					if parsedInfo.SignalStrength != 0 {
						bss.SignalStrength = parsedInfo.SignalStrength
					}
					// Handle SSID update carefully
					if parsedInfo.SSID != "" && parsedInfo.SSID != "[N/A]" {
						if parsedInfo.SSID == "\u003cHidden SSID\u003e" {
							if bss.SSID == "" { // Only set to <Hidden SSID> if we don't know the real one yet
								bss.SSID = parsedInfo.SSID
							}
						} else { // It's a real SSID name, update/overwrite
							bss.SSID = parsedInfo.SSID
						}
					} // Don't clear known SSID if current frame has no SSID IE

					if parsedInfo.Channel != 0 {
						bss.Channel = parsedInfo.Channel
					}
					if parsedInfo.Bandwidth != "" {
						bss.Bandwidth = parsedInfo.Bandwidth
					}

					// Update capabilities from parsed structures only from Beacon/ProbeResp
					if parsedInfo.ParsedHTCaps != nil {
						if bss.HTCapabilities == nil {
							bss.HTCapabilities = &HTCapabilities{}
						}
						bss.HTCapabilities.ChannelWidth40MHz = parsedInfo.ParsedHTCaps.ChannelWidth40MHz
						bss.HTCapabilities.ShortGI20MHz = parsedInfo.ParsedHTCaps.ShortGI20MHz
						bss.HTCapabilities.ShortGI40MHz = parsedInfo.ParsedHTCaps.ShortGI40MHz
					} // Keep old HT caps if not present

					if parsedInfo.ParsedVHTCaps != nil {
						if bss.VHTCapabilities == nil {
							bss.VHTCapabilities = &VHTCapabilities{}
						}
						bss.VHTCapabilities.ShortGI80MHz = parsedInfo.ParsedVHTCaps.ShortGI80MHz
						bss.VHTCapabilities.ShortGI160MHz = parsedInfo.ParsedVHTCaps.ShortGI160MHz
						bss.VHTCapabilities.ChannelWidth80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 1)
						bss.VHTCapabilities.ChannelWidth160MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 2)
						bss.VHTCapabilities.ChannelWidth80Plus80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet == 3)
					} // Keep old VHT caps if not present

					// Update Security only from Beacon/ProbeResp
					if len(parsedInfo.RSNRaw) > 0 {
						isWPA := false
						for _, rsnElem := range parsedInfo.RSNRaw {
							if rsnElem > 0 {
								isWPA = true
								break
							}
						}
						if isWPA {
							bss.Security = "RSN/WPA2/WPA3"
						} // Overwrite based on RSN presence
					} else {
						// Only set to Open if security isn't already known (e.g. from previous RSN)
						if bss.Security == "" {
							bss.Security = "Open"
						}
					}
				} // End of Beacon/ProbeResp specific attribute updates
			} // End if bss != nil

			// --- Association logic (can be triggered by other Mgmt frames, but only if BSS is known) ---
			// Re-fetch bss in case it was set to nil above because it didn't exist and wasn't Beacon/ProbeResp
			// This ensures association logic only runs if the BSS is actually in our map.
			bss, bssExists = sm.bssInfos[bssidStr]

			if bssExists && bss != nil { // Proceed with association logic only if BSS is known
				switch parsedInfo.FrameType {
				case layers.Dot11TypeMgmtAssociationReq:
					staMAC := parsedInfo.SA.String()
					if sta, staExists := sm.staInfos[staMAC]; staExists {
						sta.AssociatedBSSID = bssidStr
						if bss, bssOk := sm.bssInfos[bssidStr]; bssOk {
							bss.AssociatedSTAs[staMAC] = sta
						}
					}
				case layers.Dot11TypeMgmtReassociationReq: // Corrected constant
					staMAC := parsedInfo.SA.String()
					if sta, staExists := sm.staInfos[staMAC]; staExists {
						sta.AssociatedBSSID = bssidStr
						if bss, bssOk := sm.bssInfos[bssidStr]; bssOk {
							bss.AssociatedSTAs[staMAC] = sta
						}
					}
				case layers.Dot11TypeMgmtAssociationResp: // Corrected constant
					staMAC := parsedInfo.DA.String()
					if sta, staExists := sm.staInfos[staMAC]; staExists {
						sta.AssociatedBSSID = parsedInfo.SA.String()
						if bss, bssOk := sm.bssInfos[parsedInfo.SA.String()]; bssOk {
							bss.AssociatedSTAs[staMAC] = sta
						}
					}
				case layers.Dot11TypeMgmtReassociationResp: // Corrected constant
					staMAC := parsedInfo.DA.String()
					if sta, staExists := sm.staInfos[staMAC]; staExists {
						sta.AssociatedBSSID = parsedInfo.SA.String()
						if bss, bssOk := sm.bssInfos[parsedInfo.SA.String()]; bssOk {
							bss.AssociatedSTAs[staMAC] = sta
						}
					}
				case layers.Dot11TypeMgmtDisassociation: // Corrected constant
					if parsedInfo.SA != nil && parsedInfo.DA != nil {
						saStr := parsedInfo.SA.String()
						daStr := parsedInfo.DA.String()
						if sta, staExists := sm.staInfos[saStr]; staExists { // STA sent disassoc
							if sta.AssociatedBSSID == daStr {
								if bss, bssOk := sm.bssInfos[daStr]; bssOk {
									delete(bss.AssociatedSTAs, saStr)
								}
								sta.AssociatedBSSID = ""
							}
						} else if _, bssExists := sm.bssInfos[saStr]; bssExists { // BSS sent disassoc
							if sta, staExists := sm.staInfos[daStr]; staExists {
								if sta.AssociatedBSSID == saStr {
									sta.AssociatedBSSID = ""
								}
							}
							if bss, bssOk := sm.bssInfos[saStr]; bssOk {
								delete(bss.AssociatedSTAs, daStr)
							}
						}
					}
				case layers.Dot11TypeMgmtDeauthentication: // Corrected constant
					if parsedInfo.SA != nil && parsedInfo.DA != nil {
						saStr := parsedInfo.SA.String()
						daStr := parsedInfo.DA.String()
						if sta, staExists := sm.staInfos[saStr]; staExists { // STA sent deauth
							if sta.AssociatedBSSID == daStr {
								if bss, bssOk := sm.bssInfos[daStr]; bssOk {
									delete(bss.AssociatedSTAs, saStr)
								}
								sta.AssociatedBSSID = ""
							}
						} else if _, bssExists := sm.bssInfos[saStr]; bssExists { // BSS sent deauth
							if sta, staExists := sm.staInfos[daStr]; staExists {
								if sta.AssociatedBSSID == saStr {
									sta.AssociatedBSSID = ""
								}
							}
							if bss, bssOk := sm.bssInfos[saStr]; bssOk {
								delete(bss.AssociatedSTAs, daStr)
							}
						}
					}
				}
			}
		}
	}

	if parsedInfo.FrameType.MainType() == layers.Dot11TypeData {
		staMAC := ""
		apMAC := ""

		if parsedInfo.TA != nil {
			taStr := parsedInfo.TA.String()
			if _, isBSS := sm.bssInfos[taStr]; isBSS {
				apMAC = taStr
				if parsedInfo.RA != nil {
					if isUnicastMAC(parsedInfo.RA) {
						staMAC = parsedInfo.RA.String()
					} else {
						log.Printf("DEBUG_STATE_MANAGER: Data frame RA %s is not unicast, not considered for STA.", parsedInfo.RA.String())
						// staMAC will remain empty or its previous value,
						// and will likely be filtered by subsequent staMAC != "" checks.
					}
				}
			} else {
				staMAC = taStr
				if parsedInfo.RA != nil {
					if _, isBSS_RA := sm.bssInfos[parsedInfo.RA.String()]; isBSS_RA {
						apMAC = parsedInfo.RA.String()
					}
				}
			}
		}

		// Ensure both inferred STA and AP MACs are valid and not broadcast/zero
		if staMAC != "" && apMAC != "" && staMAC != "ff:ff:ff:ff:ff:ff" && staMAC != "00:00:00:00:00:00" && apMAC != "ff:ff:ff:ff:ff:ff" && apMAC != "00:00:00:00:00:00" {
			sta, staExists := sm.staInfos[staMAC]
			if !staExists {
				sta = NewSTAInfo(staMAC)
				sm.staInfos[staMAC] = sta
			}
			sta.LastSeen = nowMilli
			// Update signal strength if available from the data frame context
			if parsedInfo.SignalStrength != 0 {
				sta.SignalStrength = parsedInfo.SignalStrength
			}

			// Only update association if it's different or not set, to avoid flapping
			if sta.AssociatedBSSID != apMAC {
				// Remove from old BSS if previously associated to a different BSS
				if sta.AssociatedBSSID != "" && sta.AssociatedBSSID != apMAC {
					if oldBss, bssExists := sm.bssInfos[sta.AssociatedBSSID]; bssExists {
						delete(oldBss.AssociatedSTAs, staMAC)
					}
				}
				sta.AssociatedBSSID = apMAC
			}

			bss, bssExists := sm.bssInfos[apMAC]
			// --- Modification: Do NOT create new BSS from data frames ---
			if bssExists {
				// Only update existing BSS and association if BSS is already known
				bss.LastSeen = nowMilli
				if _, ok := bss.AssociatedSTAs[staMAC]; !ok {
					bss.AssociatedSTAs[staMAC] = sta
					log.Printf("DEBUG_STATE_MANAGER: Associated STA %s with existing BSS %s based on data frame.", staMAC, apMAC)
				}
			} else {
				log.Printf("DEBUG_STATE_MANAGER: Ignored potential STA %s association with unknown BSS %s based on data frame.", staMAC, apMAC)
			}
			// --- End Modification ---
		} else if staMAC != "" && staMAC != "ff:ff:ff:ff:ff:ff" && staMAC != "00:00:00:00:00:00" {
			// Update last seen for STA even if AP is unknown/invalid
			sta, staExists := sm.staInfos[staMAC]
			if !staExists {
				// Optionally create STA here? Or only create STAs when seen with a known BSS or via Mgmt frames?
				// Let's be strict for now: only update LastSeen if STA already exists.
				// sta = NewSTAInfo(staMAC)
				// sm.staInfos[staMAC] = sta
			} else {
				sta.LastSeen = nowMilli
			}
		}
	}

	// DEBUG_STATE_MANAGER log
	// Note: This log is outside the lock, which is fine for a read-only operation like len().
	// If we were accessing individual elements, we'd need to be more careful or use sm.GetSnapshot().
	// For a simple count, this should be acceptable for debugging.
	// However, to be absolutely safe and get the count from within the locked section,
	// we could log just before the defer sm.mutex.Unlock() or re-lock for this log.
	// For now, let's log it here. A more robust way would be to get counts *before* unlock.
	// Re-evaluating: It's better to log *before* unlock to ensure consistent view.
	// The defer will unlock it anyway. So, the log should be just before the function ends,
	// but *after* all modifications and *before* the unlock.
	// The `defer sm.mutex.Unlock()` is at the top. So we log before the function truly exits.
	// The current position is fine as the lock is still held due to defer.
	// Let's re-check the defer logic. Defer executes when the surrounding function returns.
	// So, if we put the log here, the lock is still held.

	// Correct placement for the log, ensuring it's within the lock's scope implicitly
	// because defer sm.mutex.Unlock() is at the top of the function.
	// However, to be explicit and clear, let's get the counts just before the unlock would happen.
	// The defer statement means sm.mutex.Unlock() will be called when ProcessParsedFrame returns.
	// So, any statement before the function's closing brace `}` is effectively within the lock.
	// log.Printf("DEBUG_STATE_MANAGER: BSS Count: %d, STA Count: %d", len(sm.bssInfos), len(sm.staInfos))
}

// GetSnapshot returns a deep copy of the current BSS and STA information.
func (sm *StateManager) GetSnapshot() Snapshot {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	bssList := make([]*BSSInfo, 0, len(sm.bssInfos))
	for bssidKey, bss := range sm.bssInfos {
		// Filter out invalid BSSIDs before creating a copy for the snapshot
		if bssidKey != "ff:ff:ff:ff:ff:ff" && bssidKey != "00:00:00:00:00:00" && bss != nil {
			bssCopy := *bss
			bssCopy.AssociatedSTAs = make(map[string]*STAInfo) // Initialize fresh for the copy
			for staMAC, staOriginal := range bss.AssociatedSTAs {
				// Filter out invalid STAs from association list and ensure original STA is valid
				if staMAC != "ff:ff:ff:ff:ff:ff" && staMAC != "00:00:00:00:00:00" && staOriginal != nil {
					// Ensure the STA being pointed to in the main staInfos map is also not invalid
					if mainSta, mainStaExists := sm.staInfos[staMAC]; mainStaExists && mainSta != nil &&
						mainSta.MACAddress != "ff:ff:ff:ff:ff:ff" && mainSta.MACAddress != "00:00:00:00:00:00" {
						staCopy := *mainSta // Copy from the main STA map to ensure consistency
						// Clear invalid association on the copy if it points to a broadcast/zero BSSID
						if staCopy.AssociatedBSSID == "ff:ff:ff:ff:ff:ff" || staCopy.AssociatedBSSID == "00:00:00:00:00:00" {
							staCopy.AssociatedBSSID = ""
						}
						bssCopy.AssociatedSTAs[staMAC] = &staCopy
					}
				}
			}
			bssList = append(bssList, &bssCopy)
		}
	}

	staList := make([]*STAInfo, 0, len(sm.staInfos))
	for staKey, sta := range sm.staInfos {
		// Filter out invalid STAs from the main STA list
		if staKey != "ff:ff:ff:ff:ff:ff" && staKey != "00:00:00:00:00:00" && sta != nil {
			staCopy := *sta
			// Ensure the associated BSSID for the STA is also not broadcast/zero if it exists
			// And also check if the BSSID it points to actually exists in our valid bssInfos
			if staCopy.AssociatedBSSID == "ff:ff:ff:ff:ff:ff" || staCopy.AssociatedBSSID == "00:00:00:00:00:00" {
				staCopy.AssociatedBSSID = "" // Clear invalid association
			} else if staCopy.AssociatedBSSID != "" {
				// Check if the BSSID it points to is valid and exists
				validBss, bssExists := sm.bssInfos[staCopy.AssociatedBSSID]
				if !bssExists || validBss == nil || validBss.BSSID == "ff:ff:ff:ff:ff:ff" || validBss.BSSID == "00:00:00:00:00:00" {
					staCopy.AssociatedBSSID = "" // Clear if associated BSS is invalid or doesn't exist
				}
			}
			staList = append(staList, &staCopy)
		}
	}
	return Snapshot{BSSs: bssList, STAs: staList}
}

func (sm *StateManager) PruneOldEntries(timeout time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	nowMilli := now.UnixMilli()
	for bssidStr, bss := range sm.bssInfos {
		if (nowMilli - bss.LastSeen) > timeout.Milliseconds() {
			log.Printf("Pruning old BSS: %s (Last seen: %v)", bssidStr, time.UnixMilli(bss.LastSeen))
			for staMAC := range bss.AssociatedSTAs {
				if sta, exists := sm.staInfos[staMAC]; exists {
					if sta.AssociatedBSSID == bssidStr {
						sta.AssociatedBSSID = ""
						log.Printf("STA %s unassociated due to BSS %s pruning.", staMAC, bssidStr)
					}
				}
			}
			delete(sm.bssInfos, bssidStr)
		}
	}

	for staMAC, sta := range sm.staInfos {
		if (nowMilli - sta.LastSeen) > timeout.Milliseconds() {
			log.Printf("Pruning old STA: %s (Last seen: %v)", staMAC, time.UnixMilli(sta.LastSeen))
			if sta.AssociatedBSSID != "" {
				if bss, exists := sm.bssInfos[sta.AssociatedBSSID]; exists {
					delete(bss.AssociatedSTAs, staMAC)
					log.Printf("STA %s removed from BSS %s's association list due to STA pruning.", staMAC, sta.AssociatedBSSID)
				}
			}
			delete(sm.staInfos, staMAC)
		}
	}
}

/*
func (bss *BSSInfo) parseCapabilitiesFromRaw(parsedInfo *frame_parser.ParsedFrameInfo) {
	// This function is now largely superseded by parsing in frame_parser.go
	// Kept for reference or if a different strategy is chosen later.
	if len(parsedInfo.HTCapabilitiesRaw) > 0 && bss.HTCapabilities == nil {
		// bss.HTCapabilities = &HTCapabilities{} // Now populated from ParsedHTCaps
	}
	if len(parsedInfo.VHTCapabilitiesRaw) > 0 && bss.VHTCapabilities == nil {
		// bss.VHTCapabilities = &VHTCapabilities{} // Now populated from ParsedVHTCaps
	}
	if len(parsedInfo.HECapabilitiesRaw) > 0 && bss.HECapabilities == nil {
		bss.HECapabilities = &HECapabilities{}
	}
}
*/

func (sm *StateManager) UpdateBSS(bssid net.HardwareAddr, ssid string, channel int, signal int, security string, lastSeen time.Time) {
	bssidStr := bssid.String()
	bss, exists := sm.bssInfos[bssidStr]
	if !exists {
		bss = NewBSSInfo(bssidStr)
		sm.bssInfos[bssidStr] = bss
	}
	if ssid != "" {
		bss.SSID = ssid
	}
	if channel != 0 {
		bss.Channel = channel
	}
	if signal != 0 {
		bss.SignalStrength = signal
	}
	if security != "" {
		bss.Security = security
	}
	bss.LastSeen = lastSeen.UnixMilli()
}

func (sm *StateManager) UpdateSTA(mac net.HardwareAddr, associatedBSSID net.HardwareAddr, signal int, lastSeen time.Time) {
	macStr := mac.String()
	sta, exists := sm.staInfos[macStr]
	if !exists {
		sta = NewSTAInfo(macStr)
		sm.staInfos[macStr] = sta
	}
	if signal != 0 {
		sta.SignalStrength = signal
	}
	sta.LastSeen = lastSeen.UnixMilli()

	assocBSSIDStr := ""
	if associatedBSSID != nil {
		assocBSSIDStr = associatedBSSID.String()
	}

	if sta.AssociatedBSSID != assocBSSIDStr {
		// Remove from old BSS association if it exists
		if sta.AssociatedBSSID != "" {
			if oldBss, bssExists := sm.bssInfos[sta.AssociatedBSSID]; bssExists {
				delete(oldBss.AssociatedSTAs, macStr)
			}
		}
		// Add to new BSS association if it exists
		if assocBSSIDStr != "" {
			if newBss, bssExists := sm.bssInfos[assocBSSIDStr]; bssExists {
				newBss.AssociatedSTAs[macStr] = sta
			}
			// Note: We don't create a new BSS here if it doesn't exist based on UpdateSTA call
		}
		// Update the STA's associated BSSID field
		sta.AssociatedBSSID = assocBSSIDStr
	}
}

// ClearState resets the BSS and STA information in the StateManager.
func (sm *StateManager) ClearState() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.bssInfos = make(map[string]*BSSInfo)
	sm.staInfos = make(map[string]*STAInfo)
	log.Println("State Manager: All BSS and STA information has been cleared.")
}

// Helper function to check if a MAC address is Unicast
// (Not Broadcast ff:ff:ff:ff:ff:ff and not Multicast - first octet's LSB is 0)
// Moved to package level
func isUnicastMAC(mac net.HardwareAddr) bool {
	if mac == nil || len(mac) != 6 {
		return false // Invalid MAC
	}
	// Check for broadcast
	if mac.String() == "ff:ff:ff:ff:ff:ff" {
		return false
	}
	// Check for multicast (LSB of the first octet is 1)
	if mac[0]&0x01 != 0 {
		return false
	}
	// Check for zero MAC
	if mac.String() == "00:00:00:00:00:00" {
		return false
	}
	return true
}
