package state_manager

import (
	"WifiPcapAnalyzer/config"       // Import for config.GlobalConfig
	"WifiPcapAnalyzer/frame_parser" // Import for ParsedFrameInfo
	"WifiPcapAnalyzer/logger"
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
	// log.Printf("DEBUG_SM_PROCESS_FRAME_INPUT: Processing ParsedFrameInfo: BSSID=%s, SA=%s, DA=%s, SSID=%s, Type=%s, Signal=%d",
	// 	parsedInfo.BSSID, parsedInfo.SA, parsedInfo.DA, parsedInfo.SSID, parsedInfo.FrameType, parsedInfo.SignalStrength)

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
			log.Printf("DEBUG_FRAME_DATA_LENGTH: Using TransportPayloadLength=%d for data frame", frameDataLength)
		} else {
			// 备选方案: 如果无法获取Transport层负载长度，使用帧总长度减去估计的头部大小
			// 假设头部大小为64字节 (Radiotap + 802.11 头部 + LLC/SNAP + IP头部)
			estimatedHeaderSize := 64
			if parsedInfo.FrameLength > estimatedHeaderSize {
				frameDataLength = parsedInfo.FrameLength - estimatedHeaderSize
				log.Printf("DEBUG_FRAME_DATA_LENGTH: TransportPayloadLength不可用，使用备选方案: FrameLength(%d) - 估计头部(%d) = %d",
					parsedInfo.FrameLength, estimatedHeaderSize, frameDataLength)
			} else {
				// 如果帧太小，至少使用一些最小值
				frameDataLength = 1
				log.Printf("DEBUG_FRAME_DATA_LENGTH: 帧长度过小(%d)，使用最小值1字节", parsedInfo.FrameLength)
			}
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
			// 添加对BitRate的更新
			if parsedInfo.BitRate > 0 { // Only update if BitRate is valid
				sta.BitRate = parsedInfo.BitRate
			}
			// Update capabilities if needed (logic omitted for brevity, similar to below)
			// log.Printf("DEBUG_SM_UPDATE: STA %s updated. LastSeen: %v, Signal: %d", macStr, time.UnixMilli(sta.LastSeen), sta.SignalStrength)
		} else {
			// STA not confirmed, check pending list
			firstSeenTime, pendingExists := sm.pendingSTAInfos[macStr]
			if pendingExists {
				// Exists in pending list, check time window
				if now.Sub(firstSeenTime) < confirmationWindow {
					// Seen again within the window, confirm it!
					// log.Printf("DEBUG_SM_UPDATE: Confirming STA %s (seen again within %v)", macStr, confirmationWindow)
					delete(sm.pendingSTAInfos, macStr) // Remove from pending
					sta = NewSTAInfo(macStr)           // Create new STA
					sta.LastSeen = nowMilli
					if parsedInfo.SignalStrength != 0 {
						sta.SignalStrength = parsedInfo.SignalStrength
					}
					if parsedInfo.BitRate > 0 { // Only update if BitRate is valid
						sta.BitRate = parsedInfo.BitRate
					}
					// Update capabilities for the newly created STA
					updateSTACapabilities(sta, parsedInfo)
					sm.staInfos[macStr] = sta // Add to confirmed list
					// log.Printf("DEBUG_SM_UPDATE: STA %s created/confirmed. Associated BSSID: %s, Signal: %d", sta.MACAddress, sta.AssociatedBSSID, sta.SignalStrength)
				} else {
					// Seen again, but outside the window. Reset the timer.
					// log.Printf("DEBUG_SM_UPDATE: Re-pending STA %s (seen again after %v)", macStr, now.Sub(firstSeenTime))
					sm.pendingSTAInfos[macStr] = now // Update timestamp
				}
			} else {
				// First time seeing this STA, add to pending list
				// log.Printf("DEBUG_SM_UPDATE: Pending STA %s (first seen)", macStr)
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

	// --- BSS Processing ---
	// Check WlanFcType for Management type (0)
	// log.Printf("DEBUG_SM_BSS_CHECK_TYPE: Frame BSSID: %s, SA: %s, DA: %s, FrameType: %s, WlanFcType: %d", parsedInfo.BSSID, parsedInfo.SA, parsedInfo.DA, parsedInfo.FrameType, parsedInfo.WlanFcType)
	if parsedInfo.WlanFcType == 0 { // 0 corresponds to Dot11TypeMgmt
		bssidMAC := parsedInfo.BSSID
		// log.Printf("DEBUG_SM_BSS_CHECK_MAC: Frame BSSID: %s, bssidMAC: %v, isUnicast: %t", parsedInfo.BSSID, bssidMAC, isUnicastMAC(bssidMAC))
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
			// log.Printf("DEBUG_SM_BSS_CHECK_BEACON_PROBE: Frame BSSID: %s, FrameType: %s, isBeaconOrProbeResp: %t", parsedInfo.BSSID, parsedInfo.FrameType, isBeaconOrProbeResp)

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
					// 带宽识别：优先使用parsedInfo.Bandwidth，该字段已经经过优化的带宽识别逻辑处理
					// 先更新capabilities然后再根据优先级确定带宽，避免capabilities信息丢失
					updateBSSCapabilities(bss, parsedInfo)

					if parsedInfo.Bandwidth != "" {
						// 优先使用Parse阶段计算的带宽
						bss.Bandwidth = parsedInfo.Bandwidth
					} else if bss.HECapabilities != nil && bss.HECapabilities.ChannelWidth160MHz {
						// 如果HE支持160MHz
						bss.Bandwidth = "160MHz"
					} else if bss.HECapabilities != nil && bss.HECapabilities.ChannelWidth80Plus80MHz {
						// 如果HE支持80+80MHz
						bss.Bandwidth = "80+80MHz"
					} else if bss.HECapabilities != nil && bss.HECapabilities.ChannelWidth40_80MHzIn5G {
						// 如果HE支持80MHz (基于ChannelWidth40_80MHzIn5G)
						bss.Bandwidth = "80MHz"
					} else if bss.VHTCapabilities != nil && bss.VHTCapabilities.ChannelWidth160MHz {
						// 如果VHT支持160MHz
						bss.Bandwidth = "160MHz"
					} else if bss.VHTCapabilities != nil && bss.VHTCapabilities.ChannelWidth80Plus80MHz {
						// 如果VHT支持80+80MHz
						bss.Bandwidth = "80+80MHz"
					} else if bss.VHTCapabilities != nil && bss.VHTCapabilities.ChannelWidth80MHz {
						// 如果VHT支持80MHz
						bss.Bandwidth = "80MHz"
					} else if bss.HTCapabilities != nil && bss.HTCapabilities.ChannelWidth40MHz {
						// 如果HT支持40MHz
						bss.Bandwidth = "40MHz"
					} else {
						// 默认20MHz
						bss.Bandwidth = "20MHz"
					}

					// Update Security
					if len(parsedInfo.RSNRaw) > 0 {
						logger.Log.Info().Msgf("INFO_BSS_SECURITY_UPDATE: RSN elements found for BSS %s: RSNRaw=%v", bssidStr, parsedInfo.RSNRaw)
						logger.Log.Info().Msgf("INFO_BSS_SECURITY_DETAIL: parsedInfo.Security='%s', RSNRaw length=%d",
							parsedInfo.Security, len(parsedInfo.RSNRaw))

						// 检查Security字段是否包含有效的安全类型
						if parsedInfo.Security != "" && parsedInfo.Security != "Open" {
							logger.Log.Info().Msgf("INFO_BSS_SECURITY_UPDATE: Using security type from frame parser: %s -> %s", bss.Security, parsedInfo.Security)
							bss.Security = parsedInfo.Security
						} else {
							logger.Log.Info().Msgf("INFO_BSS_SECURITY_UPDATE: Using default RSN/WPA2/WPA3 for BSS %s", bssidStr)
							bss.Security = "RSN/WPA2/WPA3"
						}
					} else {
						if bss.Security == "" {
							logger.Log.Info().Msgf("INFO_BSS_SECURITY_UPDATE: No RSN elements for BSS %s, setting security to: %s", bssidStr, parsedInfo.Security)
							bss.Security = parsedInfo.Security
						}
					}
				}
				// log.Printf("DEBUG_SM_UPDATE: BSS %s updated. LastSeen: %v, SSID: %s, Signal: %d", bssidStr, time.UnixMilli(bss.LastSeen), bss.SSID, bss.SignalStrength)
				// log.Printf("DEBUG_SM_BSS_UPDATE: BSS %s created/updated. SSID: '%s', Channel: %d, Signal: %d, LastSeen: %v", bss.BSSID, bss.SSID, bss.Channel, bss.SignalStrength, time.UnixMilli(bss.LastSeen))
			} else {
				// BSS not confirmed, check pending list (only for Beacon/ProbeResp)
				if isBeaconOrProbeResp {
					firstSeenTime, pendingExists := sm.pendingBSSInfos[bssidStr]
					if pendingExists {
						// Exists in pending, check time window
						if now.Sub(firstSeenTime) < confirmationWindow {
							// Seen again within window, try to confirm
							// log.Printf("DEBUG_SM_UPDATE: Confirming BSS %s (seen again within %v)", bssidStr, confirmationWindow)
							delete(sm.pendingBSSInfos, bssidStr) // Remove from pending

							// --- Apply RSSI and Completeness Filters before confirming ---
							minRSSI := config.GlobalConfig.MinBSSCreationRSSI
							// log.Printf("DEBUG_SM_BSS_FILTER_RSSI: BSSID: %s, Signal: %d, MinRSSI: %d, Pass: %t", bssidStr, parsedInfo.SignalStrength, minRSSI, parsedInfo.SignalStrength >= minRSSI)
							if parsedInfo.SignalStrength < minRSSI {
								// log.Printf("DEBUG_STATE_MANAGER: Confirmation failed for BSS %s. Signal %d dBm < threshold %d dBm.", bssidStr, parsedInfo.SignalStrength, minRSSI)
							} else {
								isSsidMissing := (parsedInfo.SSID == "" || parsedInfo.SSID == "[N/A]" || parsedInfo.SSID == "<Hidden SSID>" || parsedInfo.SSID == "<Invalid SSID Encoding>")
								isSecurityMissing := len(parsedInfo.RSNRaw) == 0
								areCapsMissing := parsedInfo.ParsedHTCaps == nil && parsedInfo.ParsedVHTCaps == nil
								passCompleteness := !(isSsidMissing && isSecurityMissing && areCapsMissing)
								// log.Printf("DEBUG_SM_BSS_FILTER_COMPLETE: BSSID: %s, SSIDMissing: %t, SecurityMissing: %t, CapsMissing: %t, Pass: %t", bssidStr, isSsidMissing, isSecurityMissing, areCapsMissing, passCompleteness)
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
									// Update Capabilities
									updateBSSCapabilities(bss, parsedInfo)
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
									// log.Printf("DEBUG_SM_UPDATE: BSS %s created/confirmed. SSID: %s, Channel: %d, Signal: %d", bss.BSSID, bss.SSID, bss.Channel, bss.SignalStrength)
									// log.Printf("DEBUG_SM_BSS_UPDATE: BSS %s created/updated. SSID: '%s', Channel: %d, Signal: %d, LastSeen: %v", bss.BSSID, bss.SSID, bss.Channel, bss.SignalStrength, time.UnixMilli(bss.LastSeen))
									// log.Printf("DEBUG_SM_BSS_UPDATE: BSS %s created/updated. SSID: %s, Channel: %d, Signal: %d", bss.BSSID, bss.SSID, bss.Channel, bss.SignalStrength) // Original log, will be replaced by the more detailed one below
								}
							}
						} else {
							// Seen again, but outside window. Reset timer.
							// log.Printf("DEBUG_SM_UPDATE: Re-pending BSS %s (seen again after %v)", bssidStr, now.Sub(firstSeenTime))
							sm.pendingBSSInfos[bssidStr] = now // Update timestamp
						}
					} else {
						// First time seeing this BSS, add to pending list
						// log.Printf("DEBUG_SM_UPDATE: Pending BSS %s (first seen)", bssidStr)
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

	// Accumulate metrics for confirmed BSS and STA
	if parsedInfo.BSSID != nil {
		bssidStr := parsedInfo.BSSID.String()
		if bss, exists := sm.bssInfos[bssidStr]; exists && bss != nil {
			// bss.totalAirtime += frameAirtime // Airtime calculation removed
			if frameDataLength > 0 {
				bss.totalTxBytes += int64(frameDataLength)
				log.Printf("DEBUG_METRIC_ACCUM: BSSID: %s, Added Bytes: %d, Total Bytes: %d for Throughput", bssidStr, frameDataLength, bss.totalTxBytes)
			}

			// Accumulate MACDurationID
			// Check if the frame is a Control frame and PS-Poll subtype
			isCtrlPSPoll := false
			// Check WlanFcType for Control (1) and WlanFcSubtype for PS-Poll (10)
			if parsedInfo.WlanFcType == 1 && parsedInfo.WlanFcSubtype == 10 { // 10 corresponds to SubtypeCtrlPSPoll
				isCtrlPSPoll = true
				log.Printf("DEBUG_NAV_SKIP: Skipping NAV accumulation for PS-Poll frame. BSSID: %s, SA: %s, DurationID: %d", bssidStr, parsedInfo.SA, parsedInfo.MACDurationID)
			}

			if !isCtrlPSPoll {
				bss.AccumulatedNavMicroseconds += uint64(parsedInfo.MACDurationID)
				log.Printf("DEBUG_METRIC_ACCUM: BSSID: %s, Added NAV Microseconds: %d, Total NAV Microseconds: %d for Channel Utilization", bssidStr, parsedInfo.MACDurationID, bss.AccumulatedNavMicroseconds)
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

				// 首先判断是否为BSS/AP MAC地址
				isSAaBSS := false
				if _, bssExists := sm.bssInfos[saStr]; bssExists {
					isSAaBSS = true
				}

				// 如果源地址是BSS，那么这个包是从AP发往STA的（下行）
				if isSAaBSS && parsedInfo.DA != nil {
					daStr := parsedInfo.DA.String()
					if staDest, staDestExists := sm.staInfos[daStr]; staDestExists && staDest != nil {
						staDest.totalDownlinkBytes += int64(frameDataLength)
						// 更新累积下行统计 - 接收字节数
						staDest.RxBytes += int64(frameDataLength)
						// 更新累积下行统计 - 接收包数
						staDest.RxPackets++
						// 如果是重传包，更新重传计数
						if parsedInfo.RetryFlag {
							staDest.RxRetries++
						}
						log.Printf("DEBUG_METRIC_ACCUM: BSS to STA: %s to %s, Added Downlink Bytes: %d, Total Downlink: %d, RxBytes: %d, RxPackets: %d",
							saStr, daStr, frameDataLength, staDest.totalDownlinkBytes, staDest.RxBytes, staDest.RxPackets)
					}
				} else {
					// 如果源地址是STA，且目标地址是关联的BSS，则为上行
					if sta.AssociatedBSSID != "" && parsedInfo.DA != nil && parsedInfo.DA.String() == sta.AssociatedBSSID {
						sta.totalUplinkBytes += int64(frameDataLength)
						// 更新累积上行统计 - 发送字节数
						sta.TxBytes += int64(frameDataLength)
						// 更新累积上行统计 - 发送包数
						sta.TxPackets++
						// 如果是重传包，更新重传计数
						if parsedInfo.RetryFlag {
							sta.TxRetries++
						}
						log.Printf("DEBUG_METRIC_ACCUM: STA to BSS: %s to %s, Added Uplink Bytes: %d, Total Uplink: %d, TxBytes: %d, TxPackets: %d",
							saStr, sta.AssociatedBSSID, frameDataLength, sta.totalUplinkBytes, sta.TxBytes, sta.TxPackets)
					} else {
						// 如果STA已关联但目标不是BSS，可能是STA之间的通信或其他流量
						// 对于已关联的STA，查看目标MAC是否是另一个STA
						if sta.AssociatedBSSID != "" && parsedInfo.DA != nil {
							daStr := parsedInfo.DA.String()
							if _, daIsInStaList := sm.staInfos[daStr]; daIsInStaList {
								// STA到另一个STA的流量，这里我们暂时计为上行
								sta.totalUplinkBytes += int64(frameDataLength)
								// 更新累积上行统计 - 发送字节数
								sta.TxBytes += int64(frameDataLength)
								// 更新累积上行统计 - 发送包数
								sta.TxPackets++
								// 如果是重传包，更新重传计数
								if parsedInfo.RetryFlag {
									sta.TxRetries++
								}
								log.Printf("DEBUG_METRIC_ACCUM: STA to Another STA: %s to %s, Added as Uplink Bytes: %d, Total Uplink: %d, TxBytes: %d, TxPackets: %d",
									saStr, daStr, frameDataLength, sta.totalUplinkBytes, sta.TxBytes, sta.TxPackets)
							} else {
								// STA发往非AP非STA的目标，记为上行
								sta.totalUplinkBytes += int64(frameDataLength)
								// 更新累积上行统计 - 发送字节数
								sta.TxBytes += int64(frameDataLength)
								// 更新累积上行统计 - 发送包数
								sta.TxPackets++
								// 如果是重传包，更新重传计数
								if parsedInfo.RetryFlag {
									sta.TxRetries++
								}
								log.Printf("DEBUG_METRIC_ACCUM: STA to Other: %s to %s, Added as Uplink Bytes: %d, Total Uplink: %d, TxBytes: %d, TxPackets: %d",
									saStr, parsedInfo.DA.String(), frameDataLength, sta.totalUplinkBytes, sta.TxBytes, sta.TxPackets)
							}
						} else {
							// 未关联的STA，将其流量计为上行
							sta.totalUplinkBytes += int64(frameDataLength)
							// 更新累积上行统计 - 发送字节数
							sta.TxBytes += int64(frameDataLength)
							// 更新累积上行统计 - 发送包数
							sta.TxPackets++
							// 如果是重传包，更新重传计数
							if parsedInfo.RetryFlag {
								sta.TxRetries++
							}
							log.Printf("DEBUG_METRIC_ACCUM: Unassociated STA: %s, Added as Uplink Bytes: %d, Total Uplink: %d, TxBytes: %d, TxPackets: %d",
								saStr, frameDataLength, sta.totalUplinkBytes, sta.TxBytes, sta.TxPackets)
						}
					}
				}

				if sta.totalUplinkBytes != originalUplink || sta.totalDownlinkBytes != originalDownlink {
					// 此日志在上面的具体UL/DL日志外是冗余的
					// log.Printf("DEBUG_METRIC_ACCUM_STA_BYTES: STA: %s, Added Bytes: %d. Uplink: %d -> %d, Downlink: %d -> %d", saStr, frameDataLength, originalUplink, sta.totalUplinkBytes, originalDownlink, sta.totalDownlinkBytes)
				}
			}

			// Accumulate MACDurationID for STA (similar to BSS NAV accumulation)
			// Check if the frame is a Control frame and PS-Poll subtype
			isCtrlPSPoll := false
			// Check WlanFcType for Control (1) and WlanFcSubtype for PS-Poll (10)
			if parsedInfo.WlanFcType == 1 && parsedInfo.WlanFcSubtype == 10 { // 10 corresponds to SubtypeCtrlPSPoll
				isCtrlPSPoll = true
				// log.Printf("DEBUG_NAV_SKIP: Skipping NAV accumulation for PS-Poll frame. STA: %s, DurationID: %d", saStr, parsedInfo.MACDurationID)
			}

			if !isCtrlPSPoll {
				sta.AccumulatedNavMicroseconds += uint64(parsedInfo.MACDurationID)
				// log.Printf("DEBUG_METRIC_ACCUM: STA: %s, Added NAV Microseconds: %d, Total NAV Microseconds: %d for Channel Utilization", saStr, parsedInfo.MACDurationID, sta.AccumulatedNavMicroseconds)
			}
		}
	}

	// 处理下行流量：当DA是STA且SA是STA关联的BSS时
	if parsedInfo.DA != nil && isUnicastMAC(parsedInfo.DA) {
		daStr := parsedInfo.DA.String()

		// 查找目标地址对应的STA
		if staDest, exists := sm.staInfos[daStr]; exists && staDest != nil {
			// 如果SA存在并且是BSS
			if parsedInfo.SA != nil && isUnicastMAC(parsedInfo.SA) {
				saStr := parsedInfo.SA.String()

				// 检查源MAC是否是BSS
				_, isSAaBSS := sm.bssInfos[saStr]

				// 如果源MAC是BSS，或者源MAC与STA关联的BSS相同
				if (isSAaBSS || (staDest.AssociatedBSSID != "" && saStr == staDest.AssociatedBSSID)) && frameDataLength > 0 {
					staDest.totalDownlinkBytes += int64(frameDataLength)
					// 更新累积下行统计 - 接收字节数
					staDest.RxBytes += int64(frameDataLength)
					// 更新累积下行统计 - 接收包数
					staDest.RxPackets++
					// 如果是重传包，更新重传计数
					if parsedInfo.RetryFlag {
						staDest.RxRetries++
					}
					log.Printf("DEBUG_METRIC_ACCUM: BSS to STA (Down): %s to %s, Added Downlink Bytes: %d, Total Downlink: %d, RxBytes: %d, RxPackets: %d",
						saStr, daStr, frameDataLength, staDest.totalDownlinkBytes, staDest.RxBytes, staDest.RxPackets)
				}
			}

			// 还要考虑传输地址是BSS但源地址不一定是BSS的情况（可能是来自其他设备通过BSS中继的帧）
			if parsedInfo.TA != nil && isUnicastMAC(parsedInfo.TA) {
				taStr := parsedInfo.TA.String()

				// 检查传输MAC是否是BSS
				_, isTAaBSS := sm.bssInfos[taStr]

				// 如果传输MAC是BSS或与STA关联的BSS相同，且之前未计算为下行
				if (isTAaBSS || (staDest.AssociatedBSSID != "" && taStr == staDest.AssociatedBSSID)) &&
					frameDataLength > 0 && parsedInfo.SA == nil {
					staDest.totalDownlinkBytes += int64(frameDataLength)
					// 更新累积下行统计 - 接收字节数
					staDest.RxBytes += int64(frameDataLength)
					// 更新累积下行统计 - 接收包数
					staDest.RxPackets++
					// 如果是重传包，更新重传计数
					if parsedInfo.RetryFlag {
						staDest.RxRetries++
					}
					log.Printf("DEBUG_METRIC_ACCUM: BSS(TA) to STA: %s to %s, Added Downlink Bytes: %d, Total Downlink: %d, RxBytes: %d, RxPackets: %d",
						taStr, daStr, frameDataLength, staDest.totalDownlinkBytes, staDest.RxBytes, staDest.RxPackets)
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

		originalChannelUtilization := bss.ChannelUtilization
		originalThroughput := bss.Throughput // Commented out as it's unused after commenting logs

		if bss.lastCalcTime.IsZero() {
			log.Printf("DEBUG_METRIC_CALC_BSS_INIT: BSSID: %s, First calculation cycle (lastCalcTime is zero). Setting metrics to 0.", bssID)
			bss.ChannelUtilization = 0
			bss.Throughput = 0
		} else {
			elapsed := now.Sub(bss.lastCalcTime).Seconds()
			log.Printf("DEBUG_METRIC_CALC_BSS_ELAPSED: BSSID: %s, Time since last calculation: %.2fs", bssID, elapsed)
			if elapsed > 0 {
				totalWindowMicroseconds := calculationWindowSeconds * 1_000_000
				if totalWindowMicroseconds > 0 {
					// Calculate channel utilization from NAV accumulation
					accumulatedNavSeconds := float64(bss.AccumulatedNavMicroseconds) / 1000000.0 // Convert microseconds to seconds
					bss.ChannelUtilization = (accumulatedNavSeconds / calculationWindowSeconds) * 100
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
				bss.Util = bss.ChannelUtilization
				bss.Thrpt = bss.Throughput
			} else {
				log.Printf("DEBUG_METRIC_CALC_BSS_NO_ELAPSED: BSSID: %s, Elapsed time is not positive (%.2fs). Setting metrics to 0 for this cycle.", bssID, elapsed)
				bss.ChannelUtilization = 0
				bss.Throughput = 0
				bss.Util = bss.ChannelUtilization
				bss.Thrpt = bss.Throughput
			}
		}
		log.Printf("DEBUG_METRIC_CALC_BSS_POST: BSSID: %s, Calculated ChannelUtilization: %.2f%% (was %.2f%%), Throughput: %d bps (was %d bps)", bssID, bss.ChannelUtilization, originalChannelUtilization, bss.Throughput, originalThroughput)
		log.Printf("DEBUG_METRIC_UPDATE_BSS: Updating BSS %s: ChannelUtil=%.2f, Throughput=%d", bssID, bss.ChannelUtilization, bss.Throughput)

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

		originalSTAChannelUtilization := sta.ChannelUtilization
		originalSTAUplinkThroughput := sta.UplinkThroughput
		originalSTADownlinkThroughput := sta.DownlinkThroughput

		if sta.lastCalcTime.IsZero() {
			log.Printf("DEBUG_METRIC_CALC_STA_INIT: STA: %s, First calculation cycle. Setting metrics to 0.", staMAC)
			sta.ChannelUtilization = 0
			sta.UplinkThroughput = 0
			sta.DownlinkThroughput = 0
		} else {
			elapsed := now.Sub(sta.lastCalcTime).Seconds()
			log.Printf("DEBUG_METRIC_CALC_STA_ELAPSED: STA: %s, Time since last calculation: %.2fs", staMAC, elapsed)
			if elapsed > 0 {
				// STA Channel Utilization - now using AccumulatedNavMicroseconds similar to BSS
				accumulatedNavSeconds := float64(sta.AccumulatedNavMicroseconds) / 1000000.0 // Convert microseconds to seconds
				sta.ChannelUtilization = (accumulatedNavSeconds / calculationWindowSeconds) * 100
				if sta.ChannelUtilization < 0 {
					sta.ChannelUtilization = 0
				}
				if sta.ChannelUtilization > 100.0 {
					sta.ChannelUtilization = 100.0
				}
				sta.UplinkThroughput = int64(float64(sta.totalUplinkBytes*8) / calculationWindowSeconds)
				sta.DownlinkThroughput = int64(float64(sta.totalDownlinkBytes*8) / calculationWindowSeconds)
				sta.Util = sta.ChannelUtilization
				sta.Thrpt = sta.UplinkThroughput + sta.DownlinkThroughput
			} else {
				log.Printf("DEBUG_METRIC_CALC_STA_NO_ELAPSED: STA: %s, Elapsed time is not positive (%.2fs). Setting metrics to 0.", staMAC, elapsed)
				sta.ChannelUtilization = 0
				sta.UplinkThroughput = 0
				sta.DownlinkThroughput = 0
				sta.Util = sta.ChannelUtilization
				sta.Thrpt = sta.UplinkThroughput + sta.DownlinkThroughput
			}
		}
		log.Printf("DEBUG_METRIC_CALC_STA_POST: STA: %s, Calculated CU: %.2f%% (was %.2f%%), UL: %d bps (was %d), DL: %d bps (was %d)", staMAC, sta.ChannelUtilization, originalSTAChannelUtilization, sta.UplinkThroughput, originalSTAUplinkThroughput, sta.DownlinkThroughput, originalSTADownlinkThroughput)
		log.Printf("DEBUG_METRIC_UPDATE_STA: Updating STA %s: ChannelUtil=%.2f, UplinkTput=%d, DownlinkTput=%d", staMAC, sta.ChannelUtilization, sta.UplinkThroughput, sta.DownlinkThroughput)

		// Update history
		sta.HistoricalChannelUtilization = append(sta.HistoricalChannelUtilization, sta.ChannelUtilization)
		sta.HistoricalUplinkThroughput = append(sta.HistoricalUplinkThroughput, sta.UplinkThroughput)
		sta.HistoricalDownlinkThroughput = append(sta.HistoricalDownlinkThroughput, sta.DownlinkThroughput)

		// Maintain history limit
		if len(sta.HistoricalChannelUtilization) > sm.maxHistoryPoints {
			sta.HistoricalChannelUtilization = sta.HistoricalChannelUtilization[1:]
		}
		if len(sta.HistoricalUplinkThroughput) > sm.maxHistoryPoints {
			sta.HistoricalUplinkThroughput = sta.HistoricalUplinkThroughput[1:]
		}
		if len(sta.HistoricalDownlinkThroughput) > sm.maxHistoryPoints {
			sta.HistoricalDownlinkThroughput = sta.HistoricalDownlinkThroughput[1:]
		}

		// Reset counters for next calculation cycle
		sta.totalAirtime = 0 // Reset old airtime counter (even though it's not used anymore)
		sta.totalUplinkBytes = 0
		sta.totalDownlinkBytes = 0
		sta.AccumulatedNavMicroseconds = 0 // Reset NAV counter
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
	// log.Printf("DEBUG_SM_EVENT_EMIT: Preparing state snapshot. BSS count: %d, STA count: %d", len(bssList), len(staList))
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
	// log.Println("State Manager: All BSS and STA information has been cleared.")
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

// Update STA capabilities
func updateSTACapabilities(sta *STAInfo, parsedInfo *frame_parser.ParsedFrameInfo) {
	// Update HT capabilities
	if parsedInfo.ParsedHTCaps != nil {
		if sta.HTCapabilities == nil {
			sta.HTCapabilities = &HTCapabilities{}
		}
		// Basic fields
		sta.HTCapabilities.ChannelWidth40MHz = parsedInfo.ParsedHTCaps.ChannelWidth40MHz
		sta.HTCapabilities.ShortGI20MHz = parsedInfo.ParsedHTCaps.ShortGI20MHz
		sta.HTCapabilities.ShortGI40MHz = parsedInfo.ParsedHTCaps.ShortGI40MHz
		sta.HTCapabilities.PrimaryChannel = parsedInfo.ParsedHTCaps.PrimaryChannel

		// Additional fields
		sta.HTCapabilities.LDPCCoding = parsedInfo.ParsedHTCaps.LDPCCoding
		sta.HTCapabilities.FortyMhzIntolerant = parsedInfo.ParsedHTCaps.FortyMhzIntolerant
		sta.HTCapabilities.TxSTBC = parsedInfo.ParsedHTCaps.TxSTBC
		sta.HTCapabilities.RxSTBC = parsedInfo.ParsedHTCaps.RxSTBC
		sta.HTCapabilities.MaxAMSDULength = parsedInfo.ParsedHTCaps.MaxAMSDULength
		sta.HTCapabilities.DSSCck = parsedInfo.ParsedHTCaps.DSSCck
		sta.HTCapabilities.HTDelayedBlockAck = parsedInfo.ParsedHTCaps.HTDelayedBlockAck
		sta.HTCapabilities.MaxAMPDULength = parsedInfo.ParsedHTCaps.MaxAMPDULength
	}

	// Update VHT capabilities
	if parsedInfo.ParsedVHTCaps != nil {
		if sta.VHTCapabilities == nil {
			sta.VHTCapabilities = &VHTCapabilities{}
		}
		// Basic fields
		sta.VHTCapabilities.ShortGI80MHz = parsedInfo.ParsedVHTCaps.ShortGI80MHz
		sta.VHTCapabilities.ShortGI160MHz = parsedInfo.ParsedVHTCaps.ShortGI160MHz
		sta.VHTCapabilities.ChannelWidth80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 1)
		sta.VHTCapabilities.ChannelWidth160MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 2)
		sta.VHTCapabilities.ChannelWidth80Plus80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet == 3)

		// Additional fields
		sta.VHTCapabilities.MaxMPDULength = parsedInfo.ParsedVHTCaps.MaxMPDULength
		sta.VHTCapabilities.RxLDPC = parsedInfo.ParsedVHTCaps.RxLDPC
		sta.VHTCapabilities.TxSTBC = parsedInfo.ParsedVHTCaps.TxSTBC
		sta.VHTCapabilities.RxSTBC = parsedInfo.ParsedVHTCaps.RxSTBC
		sta.VHTCapabilities.SUBeamformerCapable = parsedInfo.ParsedVHTCaps.SUBeamformerCapable
		sta.VHTCapabilities.SUBeamformeeCapable = parsedInfo.ParsedVHTCaps.SUBeamformee
		sta.VHTCapabilities.MUBeamformerCapable = parsedInfo.ParsedVHTCaps.MUBeamformerCapable
		sta.VHTCapabilities.MUBeamformeeCapable = parsedInfo.ParsedVHTCaps.MUBeamformee
		sta.VHTCapabilities.BeamformeeSTS = parsedInfo.ParsedVHTCaps.BeamformeeSTS
		sta.VHTCapabilities.SoundingDimensions = parsedInfo.ParsedVHTCaps.SoundingDimensions
		sta.VHTCapabilities.MaxAMPDULengthExp = parsedInfo.ParsedVHTCaps.MaxAMPDULengthExp
		sta.VHTCapabilities.RxPatternConsistency = parsedInfo.ParsedVHTCaps.RxPatternConsistency
		sta.VHTCapabilities.TxPatternConsistency = parsedInfo.ParsedVHTCaps.TxPatternConsistency
		sta.VHTCapabilities.RxMCSMap = parsedInfo.ParsedVHTCaps.RxMCSMap
		sta.VHTCapabilities.TxMCSMap = parsedInfo.ParsedVHTCaps.TxMCSMap
		sta.VHTCapabilities.RxHighestLongGIRate = parsedInfo.ParsedVHTCaps.RxHighestLongGIRate
		sta.VHTCapabilities.TxHighestLongGIRate = parsedInfo.ParsedVHTCaps.TxHighestLongGIRate
		// 新增字段
		sta.VHTCapabilities.VHTHTCCapability = parsedInfo.ParsedVHTCaps.VHTHTCCapability
		sta.VHTCapabilities.VHTTXOPPSCapability = parsedInfo.ParsedVHTCaps.VHTTXOPPSCapability
		sta.VHTCapabilities.ChannelCenter0 = parsedInfo.ParsedVHTCaps.ChannelCenter0
		sta.VHTCapabilities.ChannelCenter1 = parsedInfo.ParsedVHTCaps.ChannelCenter1
		sta.VHTCapabilities.SupportedChannelWidthSet = parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet
	}

	// Update HE capabilities
	if parsedInfo.ParsedHECaps != nil {
		if sta.HECapabilities == nil {
			sta.HECapabilities = &HECapabilities{}
		}
		sta.HECapabilities.BSSColor = parsedInfo.ParsedHECaps.BSSColor
		sta.HECapabilities.MaxMCSForOneSS = parsedInfo.ParsedHECaps.MaxMCSForOneSS
		sta.HECapabilities.MaxMCSForTwoSS = parsedInfo.ParsedHECaps.MaxMCSForTwoSS
		sta.HECapabilities.MaxMCSForThreeSS = parsedInfo.ParsedHECaps.MaxMCSForThreeSS
		sta.HECapabilities.MaxMCSForFourSS = parsedInfo.ParsedHECaps.MaxMCSForFourSS
		sta.HECapabilities.RxHEMCSMap = parsedInfo.ParsedHECaps.RxHEMCSMap
		sta.HECapabilities.TxHEMCSMap = parsedInfo.ParsedHECaps.TxHEMCSMap
		sta.HECapabilities.HTCHESupport = parsedInfo.ParsedHECaps.HTCHESupport
		sta.HECapabilities.TwtRequesterSupport = parsedInfo.ParsedHECaps.TwtRequesterSupport
		sta.HECapabilities.TwtResponderSupport = parsedInfo.ParsedHECaps.TwtResponderSupport
		sta.HECapabilities.SUBeamformer = parsedInfo.ParsedHECaps.SUBeamformer
		sta.HECapabilities.SUBeamformee = parsedInfo.ParsedHECaps.SUBeamformee
		// 通道宽度相关字段
		sta.HECapabilities.ChannelWidth160MHz = parsedInfo.ParsedHECaps.ChannelWidth160MHz
		sta.HECapabilities.ChannelWidth80Plus80MHz = parsedInfo.ParsedHECaps.ChannelWidth80Plus80MHz
		sta.HECapabilities.ChannelWidth40_80MHzIn5G = parsedInfo.ParsedHECaps.ChannelWidth40_80MHzIn5G
	}
}

// Update BSS capabilities
func updateBSSCapabilities(bss *BSSInfo, parsedInfo *frame_parser.ParsedFrameInfo) {
	// Update HT capabilities
	if parsedInfo.ParsedHTCaps != nil {
		if bss.HTCapabilities == nil {
			bss.HTCapabilities = &HTCapabilities{}
		}
		// Basic fields
		bss.HTCapabilities.ChannelWidth40MHz = parsedInfo.ParsedHTCaps.ChannelWidth40MHz
		bss.HTCapabilities.ShortGI20MHz = parsedInfo.ParsedHTCaps.ShortGI20MHz
		bss.HTCapabilities.ShortGI40MHz = parsedInfo.ParsedHTCaps.ShortGI40MHz
		bss.HTCapabilities.PrimaryChannel = parsedInfo.ParsedHTCaps.PrimaryChannel

		// Additional fields
		bss.HTCapabilities.LDPCCoding = parsedInfo.ParsedHTCaps.LDPCCoding
		bss.HTCapabilities.FortyMhzIntolerant = parsedInfo.ParsedHTCaps.FortyMhzIntolerant
		bss.HTCapabilities.TxSTBC = parsedInfo.ParsedHTCaps.TxSTBC
		bss.HTCapabilities.RxSTBC = parsedInfo.ParsedHTCaps.RxSTBC
		bss.HTCapabilities.MaxAMSDULength = parsedInfo.ParsedHTCaps.MaxAMSDULength
		bss.HTCapabilities.DSSCck = parsedInfo.ParsedHTCaps.DSSCck
		bss.HTCapabilities.HTDelayedBlockAck = parsedInfo.ParsedHTCaps.HTDelayedBlockAck
		bss.HTCapabilities.MaxAMPDULength = parsedInfo.ParsedHTCaps.MaxAMPDULength
	}

	// Update VHT capabilities
	if parsedInfo.ParsedVHTCaps != nil {
		if bss.VHTCapabilities == nil {
			bss.VHTCapabilities = &VHTCapabilities{}
		}
		// Basic fields
		bss.VHTCapabilities.ShortGI80MHz = parsedInfo.ParsedVHTCaps.ShortGI80MHz
		bss.VHTCapabilities.ShortGI160MHz = parsedInfo.ParsedVHTCaps.ShortGI160MHz
		bss.VHTCapabilities.ChannelWidth80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 1)
		bss.VHTCapabilities.ChannelWidth160MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet >= 2)
		bss.VHTCapabilities.ChannelWidth80Plus80MHz = (parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet == 3)

		// Additional fields
		bss.VHTCapabilities.MaxMPDULength = parsedInfo.ParsedVHTCaps.MaxMPDULength
		bss.VHTCapabilities.RxLDPC = parsedInfo.ParsedVHTCaps.RxLDPC
		bss.VHTCapabilities.TxSTBC = parsedInfo.ParsedVHTCaps.TxSTBC
		bss.VHTCapabilities.RxSTBC = parsedInfo.ParsedVHTCaps.RxSTBC
		bss.VHTCapabilities.SUBeamformerCapable = parsedInfo.ParsedVHTCaps.SUBeamformerCapable
		bss.VHTCapabilities.SUBeamformeeCapable = parsedInfo.ParsedVHTCaps.SUBeamformee
		bss.VHTCapabilities.MUBeamformerCapable = parsedInfo.ParsedVHTCaps.MUBeamformerCapable
		bss.VHTCapabilities.MUBeamformeeCapable = parsedInfo.ParsedVHTCaps.MUBeamformee
		bss.VHTCapabilities.BeamformeeSTS = parsedInfo.ParsedVHTCaps.BeamformeeSTS
		bss.VHTCapabilities.SoundingDimensions = parsedInfo.ParsedVHTCaps.SoundingDimensions
		bss.VHTCapabilities.MaxAMPDULengthExp = parsedInfo.ParsedVHTCaps.MaxAMPDULengthExp
		bss.VHTCapabilities.RxPatternConsistency = parsedInfo.ParsedVHTCaps.RxPatternConsistency
		bss.VHTCapabilities.TxPatternConsistency = parsedInfo.ParsedVHTCaps.TxPatternConsistency
		bss.VHTCapabilities.RxMCSMap = parsedInfo.ParsedVHTCaps.RxMCSMap
		bss.VHTCapabilities.TxMCSMap = parsedInfo.ParsedVHTCaps.TxMCSMap
		bss.VHTCapabilities.RxHighestLongGIRate = parsedInfo.ParsedVHTCaps.RxHighestLongGIRate
		bss.VHTCapabilities.TxHighestLongGIRate = parsedInfo.ParsedVHTCaps.TxHighestLongGIRate
		bss.VHTCapabilities.VHTHTCCapability = parsedInfo.ParsedVHTCaps.VHTHTCCapability
		bss.VHTCapabilities.VHTTXOPPSCapability = parsedInfo.ParsedVHTCaps.VHTTXOPPSCapability
		bss.VHTCapabilities.ChannelCenter0 = parsedInfo.ParsedVHTCaps.ChannelCenter0
		bss.VHTCapabilities.ChannelCenter1 = parsedInfo.ParsedVHTCaps.ChannelCenter1
		bss.VHTCapabilities.SupportedChannelWidthSet = parsedInfo.ParsedVHTCaps.SupportedChannelWidthSet
	}

	// Update HE capabilities
	if parsedInfo.ParsedHECaps != nil {
		if bss.HECapabilities == nil {
			bss.HECapabilities = &HECapabilities{}
		}
		bss.HECapabilities.BSSColor = parsedInfo.ParsedHECaps.BSSColor
		bss.HECapabilities.MaxMCSForOneSS = parsedInfo.ParsedHECaps.MaxMCSForOneSS
		bss.HECapabilities.MaxMCSForTwoSS = parsedInfo.ParsedHECaps.MaxMCSForTwoSS
		bss.HECapabilities.MaxMCSForThreeSS = parsedInfo.ParsedHECaps.MaxMCSForThreeSS
		bss.HECapabilities.MaxMCSForFourSS = parsedInfo.ParsedHECaps.MaxMCSForFourSS
		bss.HECapabilities.RxHEMCSMap = parsedInfo.ParsedHECaps.RxHEMCSMap
		bss.HECapabilities.TxHEMCSMap = parsedInfo.ParsedHECaps.TxHEMCSMap
		bss.HECapabilities.HTCHESupport = parsedInfo.ParsedHECaps.HTCHESupport
		bss.HECapabilities.TwtRequesterSupport = parsedInfo.ParsedHECaps.TwtRequesterSupport
		bss.HECapabilities.TwtResponderSupport = parsedInfo.ParsedHECaps.TwtResponderSupport
		bss.HECapabilities.SUBeamformer = parsedInfo.ParsedHECaps.SUBeamformer
		bss.HECapabilities.SUBeamformee = parsedInfo.ParsedHECaps.SUBeamformee
		// 通道宽度相关字段
		bss.HECapabilities.ChannelWidth160MHz = parsedInfo.ParsedHECaps.ChannelWidth160MHz
		bss.HECapabilities.ChannelWidth80Plus80MHz = parsedInfo.ParsedHECaps.ChannelWidth80Plus80MHz
		bss.HECapabilities.ChannelWidth40_80MHzIn5G = parsedInfo.ParsedHECaps.ChannelWidth40_80MHzIn5G
	}
}
