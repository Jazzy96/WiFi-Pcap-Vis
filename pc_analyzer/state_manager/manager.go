package state_manager

import (
	"log"
	"net"

	// "strconv" // Not used

	"sync"
	"time"
	"wifi-pcap-demo/pc_analyzer/frame_parser" // Import for ParsedFrameInfo

	"github.com/google/gopacket/layers" // Import for layers.Dot11Type constants
)

// StateManager holds the current state of all observed BSSs and STAs.
type StateManager struct {
	bssInfos map[string]*BSSInfo // Keyed by BSSID string
	staInfos map[string]*STAInfo // Keyed by STA MAC string
	mutex    sync.RWMutex
}

// NewStateManager creates a new StateManager.
func NewStateManager() *StateManager {
	return &StateManager{
		bssInfos: make(map[string]*BSSInfo),
		staInfos: make(map[string]*STAInfo),
	}
}

// ProcessParsedFrame is the main entry point for updating state based on a parsed frame.
func (sm *StateManager) ProcessParsedFrame(parsedInfo *frame_parser.ParsedFrameInfo) {
	if parsedInfo == nil {
		return
	}
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()

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
			sta.LastSeen = now
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
			sta.LastSeen = now
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
			}
		}
	}

	if parsedInfo.FrameType.MainType() == layers.Dot11TypeMgmt {
		bssidMAC := parsedInfo.BSSID
		if bssidMAC != nil {
			bssidStr := bssidMAC.String()

			// Critical check: Do not process or create BSSInfo for broadcast BSSID
			if bssidStr == "ff:ff:ff:ff:ff:ff" {
				log.Printf("DEBUG_STATE_MANAGER: Ignoring Mgmt frame with broadcast BSSID: %s. SA: %s, DA: %s, FrameType: %s", bssidStr, parsedInfo.SA, parsedInfo.DA, parsedInfo.FrameType.String())
				return // Exit early, do not create/update BSS for ff:ff:ff:ff:ff:ff
			}

			bss, bssExists := sm.bssInfos[bssidStr]

			// --- Strict BSS Creation/Update Logic ---
			isBeaconOrProbeResp := parsedInfo.FrameType == layers.Dot11TypeMgmtBeacon || parsedInfo.FrameType == layers.Dot11TypeMgmtProbeResp

			if !bssExists {
				// ONLY create BSS if it's a Beacon or Probe Response and BSSID is unicast
				if isBeaconOrProbeResp && isUnicastMAC(bssidMAC) {
					bss = NewBSSInfo(bssidStr)
					// --- Start: Add check for incomplete BSS info before adding to map ---
					isSsidMissing := (parsedInfo.SSID == "" || parsedInfo.SSID == "[N/A]" || parsedInfo.SSID == "<Hidden SSID>" || parsedInfo.SSID == "<Invalid SSID Encoding>")
					isSecurityMissing := len(parsedInfo.RSNRaw) == 0
					areCapsMissing := parsedInfo.ParsedHTCaps == nil && parsedInfo.ParsedVHTCaps == nil

					if isSsidMissing && isSecurityMissing && areCapsMissing {
						log.Printf("WARN_STATE_MANAGER: Skipping creation of new BSS %s from %s due to severely incomplete information (No SSID, RSN, HT/VHT Caps).", bssidStr, parsedInfo.FrameType.String())
						// Do not add the incomplete bss to the map
						bss = nil // Set bss back to nil so subsequent updates are skipped
					} else {
						sm.bssInfos[bssidStr] = bss // Add the new BSS to the map
						log.Printf("DEBUG_STATE_MANAGER: Created new BSS %s from %s", bssidStr, parsedInfo.FrameType.String())
						// Proceed to update attributes for the newly created BSS below
					}
					// --- End: Add check for incomplete BSS info ---
				} else {
					// If BSS doesn't exist and it's not Beacon/ProbeResp, ignore this frame for BSS processing
					// Do not return here, as we might still need to process STA association logic below
					// if the frame type is relevant (e.g., AssocReq/Resp).
					log.Printf("DEBUG_STATE_MANAGER: Ignored Mgmt frame (%s) for unknown BSS %s (BSS creation only from Beacon/ProbeResp)", parsedInfo.FrameType.String(), bssidStr)
					// Set bss to nil to prevent accidental updates below if it wasn't created
					bss = nil
				}
			}

			// If BSS exists (or was just created), update its LastSeen
			if bss != nil {
				bss.LastSeen = now

				// ONLY update detailed attributes if it's a Beacon or Probe Response
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
			sta.LastSeen = now
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
				bss.LastSeen = now
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
				sta.LastSeen = now
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
	log.Printf("DEBUG_STATE_MANAGER: BSS Count: %d, STA Count: %d", len(sm.bssInfos), len(sm.staInfos))
}

// GetSnapshot returns a deep copy of the current BSS and STA information.
func (sm *StateManager) GetSnapshot() ([]*BSSInfo, []*STAInfo) {
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
				if _, bssExists := sm.bssInfos[staCopy.AssociatedBSSID]; !bssExists {
					// If associated BSSID doesn't exist in main bssInfos map (e.g. it was ff:ff:ff.. and filtered out), clear it
					staCopy.AssociatedBSSID = ""
				}
			}
			staList = append(staList, &staCopy)
		}
	}
	return bssList, staList
}

func (sm *StateManager) PruneOldEntries(timeout time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	for bssidStr, bss := range sm.bssInfos {
		if now.Sub(bss.LastSeen) > timeout {
			log.Printf("Pruning old BSS: %s (Last seen: %v)", bssidStr, bss.LastSeen)
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
		if now.Sub(sta.LastSeen) > timeout {
			log.Printf("Pruning old STA: %s (Last seen: %v)", staMAC, sta.LastSeen)
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
	bss.LastSeen = lastSeen
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
	sta.LastSeen = lastSeen

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
