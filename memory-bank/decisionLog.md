---
### Decision (Debug - Parser Robustness)
[2025-05-10 17:19:00] - [Enhance `parser.go` to be more tolerant of empty/malformed non-critical fields]

**Rationale:**
User reported that the frontend shows no data, and BSS/STA entries are not created as expected. Previous logs indicated many empty fields in `tshark`'s CSV output. The existing parsing logic in `ProcessRow` was strict, potentially discarding entire frames if non-critical fields like `radiotap.channel.freq`, `radiotap.dbm_antsignal`, or `wlan.duration` were present but malformed, or if `frame.time_epoch` parsing failed (which used `time.Now()` as a fallback, which is not ideal). This strictness could prevent valid frame data from reaching the state manager.

**Details:**
*   **Affected Files:**
    *   [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go)
*   **Changes Made:**
    1.  **Timestamp Fallback:** If `frame.time_epoch` parsing fails, `info.Timestamp` is now set to `time.Time{}` (zero value). The parsing failure for this field remains a critical error that causes the frame to be skipped.
    2.  **Relaxed Error Handling for `radiotap.channel.freq`:** If this field is present but unparsable, an error is logged, but it's no longer added to `parseErrors` (which would discard the frame). The field in `ParsedFrameInfo` retains its default.
    3.  **Relaxed Error Handling for `radiotap.dbm_antsignal`:** Similar to `radiotap.channel.freq`, parsing errors for present-but-malformed values are logged but do not cause frame discard.
    4.  **Relaxed Error Handling for `wlan.duration`:** Similar to the above, parsing errors for present-but-malformed values are logged but do not cause frame discard.
*   **Expected Outcome:** The parser should now be more resilient. Frames with issues in these less critical fields will still be processed for their valid data, increasing the likelihood of BSS/STA entries being created and data appearing on the frontend. Critical parsing errors (e.g., for `frame.len`, `wlan.fc.type_subtype`, or malformed MACs) will still lead to frame discard.
---
### Decision (Debug - CSV Data Parsing)
[2025-05-10 16:29:36] - [Request detailed logs for CSV row parsing in `parser.go`]

**Rationale:**
The `tshark` command execution and CSV header parsing are now successful, as indicated by the latest logs. However, the user still reports errors, suggesting the problem lies in the parsing of individual CSV data rows or subsequent data processing. To pinpoint the issue, detailed logging within the `ProcessRow` function (or its equivalent) in [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go) is required.

**Details:**
*   **Affected Files:**
    *   [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go)
*   **Next Step:** Ask the user to add specific `log.Printf` and `log.Errorf` statements to trace:
    1.  Raw CSV row data.
    2.  Parsing attempts for each key field (name, raw value).
    3.  Errors during field extraction or conversion (field name, raw value, error details).
    4.  Successfully parsed `ParsedFrameInfo` object summary before callback.
    5.  A marker before calling the `packetInfoHandler`.
*   This will provide granular insight into the data parsing flow and help identify where the error occurs.
---
### Decision (Debug - tshark Field Correction)
[2025-05-10 16:17:00] - [Correct invalid tshark fields and remove problematic ones]

**Rationale:**
Based on `tshark` error logs and `tshark_beacon_example.json`, several fields in the `defaultTsharkFields` list in `parser.go` were either incorrect or not consistently available, causing `tshark` to fail or report errors. The strategy is to correct known field names and remove fields that are problematic and not strictly essential for core functionality, ensuring `tshark` can execute successfully.

**Details:**
*   **Affected Files:**
    *   [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go)
*   **Field Name Changes and Removals:**
    *   Corrected `wlan.flags.retry` to `wlan.fc.retry`.
    *   Removed `radiotap.mcs.flags` (not found in example, caused error).
    *   Removed `radiotap.vht.mcs` (caused error, VHT MCS info can be inferred from WLAN layer if needed).
    *   Removed `radiotap.vht.nss` (caused error, VHT NSS info can be inferred from WLAN layer if needed).
    *   Removed `radiotap.he.mcs` (caused error, HE MCS info can be inferred from WLAN layer if needed).
    *   Removed `radiotap.he.bw` (caused error, HE BW info can be inferred from WLAN layer if needed).
    *   Removed `radiotap.he.gi` (caused error, HE GI info can be inferred from WLAN layer if needed).
    *   Removed `radiotap.he.nss` (caused error, HE NSS info can be inferred from WLAN layer if needed).
    *   Removed `wlan.he.phy.channel_width_set` (caused error, specific subfields exist in example, or can be inferred).
*   **Impact:** These changes are expected to resolve `tshark` startup errors related to invalid fields, allowing the packet parsing process to proceed. This prioritizes getting the core parsing pipeline functional.
# Decision Log

---
### Decision (Debug - tshark Stream Handling)
[2025-05-10 15:39:00] - [Modify tshark parser to handle gRPC stream input]

**Rationale:**
The `tshark` based parser, as initially implemented, expected a file path for pcap data. However, the gRPC `pcapStreamHandler` provides an `io.Reader` representing a live data stream. This mismatch caused the `WARN_APP: pcapStreamHandler received a stream that is not a file. TShark processing requires a file path.` error. The fix involves modifying the parser to accept an `io.Reader` and pipe this stream directly to `tshark`'s standard input.

**Details:**
*   **Affected Files:**
    *   [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go)
    *   [`desktop_app/WifiPcapAnalyzer/app.go`](desktop_app/WifiPcapAnalyzer/app.go)
*   **Changes in [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go):**
    1.  **`TSharkExecutor.StartStream` method:** A new method `StartStream(pcapStream io.Reader, tsharkPath string, fields []string)` was added. This method configures `tshark` to read from standard input (`-r -`) and sets `cmd.Stdin = pcapStream`.
    2.  **`ProcessPcapStream` function:** A new function `ProcessPcapStream(pcapStream io.Reader, tsharkPath string, pktHandler PacketInfoHandler) error` was created. This function:
        *   Uses `TSharkExecutor.StartStream` to launch `tshark` with the provided `io.Reader`.
        *   The rest of the logic (CSV parsing, frame processing) is similar to `ProcessPcapFile` but adapted for the stream context (e.g., logging messages indicate stream processing).
*   **Changes in [`desktop_app/WifiPcapAnalyzer/app.go`](desktop_app/WifiPcapAnalyzer/app.go):**
    1.  **`pcapStreamHandler` modification:** The `a.pcapStreamHandler` function was updated. Instead of trying to treat the `io.Reader` as a file path, it now directly calls the new `frame_parser.ProcessPcapStream` function, passing the `pcapStream` (which is the `io.Reader` from gRPC) and the `tsharkPath` from config.
*   **Expected Outcome:** The application can now correctly process live pcap data streamed from the gRPC agent by piping it directly to `tshark`'s standard input, resolving the file path requirement issue.

---
### Decision (Code - Parser Implementation)
[2025-05-10 15:30:00] - [Implement tshark-based parsing logic]

**Rationale:**
To address persistent issues with `gopacket`'s robustness in handling diverse 802.11 frames and to improve parsing accuracy, the decision was made to replace `gopacket` with `tshark` for frame parsing, as per the architecture defined in `memory-bank/developmentContext/pcAnalysisEngine.md` and recorded in a previous decision log entry ([2025-05-10] - [Architectural Shift: Replace `gopacket` with `tshark` for 802.11 Frame Parsing]). This entry details the implementation of that architectural decision.

**Details:**
*   **Affected Files:**
    *   [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go): Major rewrite.
    *   [`desktop_app/WifiPcapAnalyzer/config/config.go`](desktop_app/WifiPcapAnalyzer/config/config.go) & [`desktop_app/WifiPcapAnalyzer/config/config.json`](desktop_app/WifiPcapAnalyzer/config/config.json): Added `TsharkPath` configuration.
    *   [`desktop_app/WifiPcapAnalyzer/app.go`](desktop_app/WifiPcapAnalyzer/app.go): Updated to call new parser function and use `TsharkPath`.
    *   [`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go): Adapted to changes in `ParsedFrameInfo` (e.g., `FrameType` is now string, `WlanFcType`/`WlanFcSubtype` used for type checks).
    *   [`desktop_app/WifiPcapAnalyzer/frame_parser/parser_test.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser_test.go): Deleted as it was based on `gopacket`.
*   **Implementation Summary in [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go):**
    1.  **`TSharkExecutor` struct:** Created to manage `tshark` process execution, including starting the process with specified fields and handling its stdout/stderr.
    2.  **`CSVParser` struct:** Implemented to read `tshark`'s CSV output, parse the header row to create a field-to-index map, and read subsequent data rows.
    3.  **`FrameProcessor` struct:** Developed to convert a parsed CSV row (map of field names to string values) into the `ParsedFrameInfo` struct. This includes:
        *   Helper functions (`getString`, `getInt`, `getMAC`, etc.) for safe extraction and type conversion of field values.
        *   Logic to parse `wlan.fc.type_subtype` (hex string) into `WlanFcType` (uint8), `WlanFcSubtype` (uint8), and a descriptive `FrameType` string.
    4.  **`ParsedFrameInfo` struct:** Modified to align with fields available from `tshark` and requirements of downstream modules. Removed direct `gopacket` layer fields.
    5.  **`ProcessPcapFile` function (replaces `ProcessPcapStream`):** Orchestrates the new workflow:
        *   Takes pcap file path and tshark path as input.
        *   Uses `TSharkExecutor` to run `tshark`.
        *   Pipes `tshark`'s stdout to `CSVParser`.
        *   Iterates through CSV rows, using `FrameProcessor` to convert each row to `ParsedFrameInfo`.
        *   Calls the `PacketInfoHandler` callback with the `ParsedFrameInfo`.
        *   Includes basic error logging for `tshark` execution and CSV/row processing.
    6.  **`getPHYRateMbps` function:** Retained and adapted to calculate PHY rate based on fields now available in `ParsedFrameInfo` (populated from `tshark`'s radiotap output).
    7.  **`CalculateFrameAirtime` function:** Logic preserved; its inputs (frame length, PHY rate) are now derived from `tshark` data.
*   **Configuration:** `TsharkPath` added to `AppConfig` and `config.json` to specify the `tshark` executable location, defaulting to "tshark" (assuming it's in system PATH).
*   **Error Handling:** Basic error logging implemented for `tshark` process issues, CSV parsing errors, and individual row processing failures. `tshark`'s stderr is also logged.
*   **Expected Outcome:** A more robust and accurate frame parsing mechanism, leveraging `tshark`'s mature dissection engine, leading to better data quality for the application.

---
(Existing content will follow this new entry)
# Decision Log

This file records architectural and implementation decisions using a list format.

2025-05-06 23:48:00 - Initial population of architectural decisions.
*
      
---
### Decision (Code - Channel Utilization Calculation)
[2025-05-08 16:36:00] - [Refactor: Use MAC Duration/ID for Channel Utilization]

**Rationale:**
The previous method for calculating channel utilization relied on estimating frame airtime based on PHY rate and frame length, which can be inaccurate and complex. The IEEE 802.11 standard's `Duration/ID` field in the MAC header provides a value (in microseconds) used for Network Allocation Vector (NAV) updates, representing the time the channel is expected to be busy. Using this field offers a potentially more accurate and standards-based way to estimate channel occupancy from the perspective of MAC layer reservations.

**Details:**
*   **Affected Files:**
    *   [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0)
    *   [`desktop_app/WifiPcapAnalyzer/state_manager/models.go`](desktop_app/WifiPcapAnalyzer/state_manager/models.go:0)
    *   [`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go:0)
*   **Changes Made:**
    1.  Added `MACDurationID uint16` to `ParsedFrameInfo` struct in `parser.go` and populated it from `layers.Dot11.DurationID`.
    2.  Added `AccumulatedNavMicroseconds uint64` to `BSSInfo` struct in `models.go`.
    3.  Modified `ProcessParsedFrame` in `manager.go` to accumulate `uint64(parsedInfo.MACDurationID)` into `bss.AccumulatedNavMicroseconds`. Added logic to skip accumulation for Control PS-Poll frames (`layers.Dot11TypeCtrlPowersavePoll`), where the Duration/ID field contains AID instead of time. Removed the accumulation logic based on `CalculateFrameAirtime`.
    4.  Modified `PeriodicallyCalculateMetrics` in `manager.go` to calculate `ChannelUtilization` using `(float64(bss.AccumulatedNavMicroseconds) / (calculationWindowSeconds * 1_000_000)) * 100`.
    5.  Ensured `bss.AccumulatedNavMicroseconds` is reset to 0 after each calculation cycle in `PeriodicallyCalculateMetrics`.
*   **Expected Outcome:** Channel utilization metric is now calculated based on the aggregated NAV values from observed frames (excluding PS-Poll), providing an alternative measure of channel busyness based on MAC layer reservations. The dependency on the `CalculateFrameAirtime` function for this specific metric is removed.
---
### Decision (Debug - Backend Metrics Initialization)
[2025-05-08 15:53:00] - [Bug Fix Strategy: Initialize `lastCalcTime` in BSSInfo/STAInfo Models]

**Rationale:**
The new performance metrics (channel utilization, throughput) were displaying as "N/A" on the frontend. Investigation revealed that `BSSInfo` and `STAInfo` objects, when newly created by `NewBSSInfo` and `NewSTAInfo` in [`desktop_app/WifiPcapAnalyzer/state_manager/models.go`](desktop_app/WifiPcapAnalyzer/state_manager/models.go:0), did not have their internal `lastCalcTime` field initialized. This caused the `PeriodicallyCalculateMetrics` function in [`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go:0) to evaluate `lastCalcTime.IsZero()` as true for these new entries. Consequently, the metrics for these entries were set to 0 during their first calculation cycle. These zero values were then pushed to the frontend, likely resulting in the "N/A" display.

**Details:**
*   **Affected Files:**
    *   [`desktop_app/WifiPcapAnalyzer/state_manager/models.go`](desktop_app/WifiPcapAnalyzer/state_manager/models.go:0)
*   **Change Made:**
    *   Modified the `NewBSSInfo` function to initialize the `lastCalcTime` field to `time.Now()`.
    *   Modified the `NewSTAInfo` function to initialize the `lastCalcTime` field to `time.Now()`.
*   **Expected Outcome:** By initializing `lastCalcTime` upon object creation, the first call to `PeriodicallyCalculateMetrics` for a new BSS or STA will have a valid `lastCalcTime`. This will allow `elapsed := now.Sub(bss.lastCalcTime).Seconds()` to return a small positive value (close to `metricsCalcInterval`), enabling the calculation of non-zero initial metrics (assuming some data has been processed for that BSS/STA). This should prevent the metrics from being 0 by default and resolve the "N/A" display issue for newly appearing devices.
---
### Decision (Debug - PC Analyzer Frame Parsing Reversion)
[2025-05-07 23:30:00] - [Reversion: Remove GBK Fallback for SSID Decoding]

**Rationale:**
After implementing the GBK fallback for SSID decoding, user feedback indicated that it did not successfully decode the problematic SSIDs and resulted in similar unreadable output ("乱码"). Given that the fallback added complexity without providing the desired benefit in this specific case, the decision was made to revert the changes.

**Details:**
*   **Affected File:** `pc_analyzer/frame_parser/parser.go`
*   **Change Made:**
    *   The code block attempting GBK decoding within the `case layers.Dot11InformationElementIDSSID:` was removed.
    *   The logic was restored to simply check `utf8.Valid(ieInfo)`. If false, `ssidContent` is set directly to `"<Invalid/Undecodable SSID>"`.
    *   The unused imports for `golang.org/x/text/encoding/simplifiedchinese` and `golang.org/x/text/transform` were removed.
*   **Expected Outcome:** The SSID parsing logic is simplified back to only validating UTF-8. Non-UTF-8 SSIDs will be consistently marked as `"<Invalid/Undecodable SSID>"`. This removes the unsuccessful GBK decoding attempt and associated dependencies.
---
### Decision (Debug - PC Analyzer Frame Parsing)
[2025-05-07 23:24:00] - [Enhancement: Implement GBK Fallback for SSID Decoding]

**Rationale:**
User inquired about the possibility that SSIDs marked as `"<Invalid SSID Encoding>"` might be using a different encoding, such as GBK, rather than being purely invalid data. To potentially improve the display of SSIDs from networks configured with non-UTF-8 encodings common in certain regions, a fallback mechanism was implemented.

**Details:**
*   **Affected File:** `pc_analyzer/frame_parser/parser.go`
*   **Dependencies Added:** `golang.org/x/text/encoding/simplifiedchinese`, `golang.org/x/text/transform` (via `go get` in `pc_analyzer` directory).
*   **Change Made:**
    *   In the `parsePacketLayers` function, within the `case layers.Dot11InformationElementIDSSID:` block:
    *   If `utf8.Valid(ieInfo)` returns false:
        1.  A GBK decoder (`simplifiedchinese.GBK.NewDecoder()`) is created.
        2.  `transform.Bytes()` is used to attempt decoding `ieInfo` using the GBK decoder.
        3.  If the GBK decoding succeeds without error, the resulting byte slice is converted to a string and assigned to `ssidContent`. An informational log (`INFO_SSID_PARSE`) is generated.
        4.  If the GBK decoding fails, `ssidContent` is set to `"<Invalid/Undecodable SSID>"`, and a warning log (`WARN_SSID_PARSE`) including the GBK error is generated.
    *   The final debug log (`DEBUG_SSID_PARSE`) now reports the final `ssidContent` regardless of the decoding path taken (UTF-8, GBK, Hidden, or Undecodable).
*   **Expected Outcome:** The parser will now make a best effort to decode SSIDs that are not valid UTF-8 using the GBK encoding. This should increase the chances of correctly displaying SSIDs from networks using GBK, while still providing a clear indicator (`<Invalid/Undecodable SSID>`) if both UTF-8 and GBK decoding fail.
---
### Decision (Debug - PC Analyzer State Management)
[2025-05-07 23:11:00] - [Bug Fix Strategy: Add Strictness Check for New BSS Creation from Beacons/Probes]

**Rationale:**
User logs confirmed that even after fixing the parser to handle truncated Probe Responses, BSS entries could still be created with missing critical information (SSID, Security, Capabilities). This happened when IE parsing of a Beacon or Probe Response frame was prematurely terminated due to encountering a malformed IE (e.g., incorrect length declaration), but the parser still returned a partial `ParsedFrameInfo`. To prevent these incomplete entries from polluting the state, the decision was made to add a quality check within the state manager before creating a new BSS.

**Details:**
*   **Affected File:** `pc_analyzer/state_manager/manager.go`
*   **Change Made:**
    *   In the `ProcessParsedFrame` function, within the logic block that handles creating a new BSS (`if !bssExists { ... }`) specifically for `MgmtBeacon` or `MgmtProbeResp` frames:
    *   After `bss = NewBSSInfo(bssidStr)`, a check was added:
        ```go
        isSsidMissing := (parsedInfo.SSID == "" || parsedInfo.SSID == "[N/A]" || parsedInfo.SSID == "<Hidden SSID>" || parsedInfo.SSID == "<Invalid SSID Encoding>")
        isSecurityMissing := len(parsedInfo.RSNRaw) == 0
        areCapsMissing := parsedInfo.ParsedHTCaps == nil && parsedInfo.ParsedVHTCaps == nil

        if isSsidMissing && isSecurityMissing && areCapsMissing {
            log.Printf("WARN_STATE_MANAGER: Skipping creation of new BSS %s from %s due to severely incomplete information...", bssidStr, parsedInfo.FrameType.String())
            bss = nil // Prevent adding to map and subsequent updates
        } else {
            sm.bssInfos[bssidStr] = bss // Add to map only if info is deemed sufficient
            log.Printf("DEBUG_STATE_MANAGER: Created new BSS %s from %s", bssidStr, parsedInfo.FrameType.String())
        }
        ```
*   **Expected Outcome:** The state manager will now be more selective when creating new BSS entries from Beacons or Probe Responses. If the frame parsing resulted in missing SSID, security (RSN), and capability information (likely due to IE parsing issues), the BSS entry will not be created, leading to a cleaner and more reliable BSS list. Updates to existing BSS entries are not affected by this specific check.
---
### Decision (Debug - PC Analyzer State Management)
[2025-05-07 23:08:00] - [Configuration Change: Reduce Timeout for Pruning State Entries]

**Rationale:**
Following the increase in pruning frequency, the user requested a shorter timeout for entries to be considered old, aiming for even faster removal of outdated STA/BSS information. The timeout was reduced from 5 minutes to 2 minutes.

**Details:**
*   **Affected File:** `pc_analyzer/main.go`
*   **Change Made:**
    *   The timeout argument passed to `stateMgr.PruneOldEntries()` within the `pruneTicker` goroutine was changed from `5 * time.Minute` to `2 * time.Minute`.
    *   The ticker frequency remains at `30 * time.Second`.
*   **Expected Outcome:** BSS/STA entries that have not been seen for more than 2 minutes will now be pruned by the state manager during its next 30-second pruning cycle. This should make the displayed data reflect changes in the wireless environment more rapidly.
---
### Decision (Debug - PC Analyzer State Management)
[2025-05-07 23:06:00] - [Configuration Change: Increase Pruning Frequency for State Entries]

**Rationale:**
User observed potentially outdated STA entries and requested a more responsive aging mechanism. The existing `PruneOldEntries` function in `state_manager.go` was called by a ticker in `main.go` every 1 minute, with a 5-minute timeout for entries. To make the pruning more reactive, the ticker frequency was increased.

**Details:**
*   **Affected File:** `pc_analyzer/main.go`
*   **Change Made:**
    *   The `pruneTicker` initialization was changed from `time.NewTicker(1 * time.Minute)` to `time.NewTicker(30 * time.Second)`.
    *   The timeout argument passed to `stateMgr.PruneOldEntries()` remains `5 * time.Minute`.
*   **Expected Outcome:** The state manager will now check for and prune old BSS/STA entries every 30 seconds instead of every minute. This should lead to a faster removal of entries that have not been seen for longer than the 5-minute timeout period, making the displayed data more current.
---
### Decision (Debug - PC Analyzer Frame Parsing)
[2025-05-07 23:00:00] - [Bug Fix Strategy: Reject Incomplete Beacon/ProbeResp Frames in Parser]

**Rationale:**
User logs showed that BSS entries were being created with missing information (SSID "(Hidden)", security "Open", no capabilities) due to `MgmtProbeResp` frames having a payload shorter than their required fixed header (12 bytes). The `frame_parser.go` previously logged a warning but still returned a partially filled `ParsedFrameInfo`, leading `state_manager.go` to create an incomplete BSS.

**Details:**
*   **Affected File:** `pc_analyzer/frame_parser/parser.go`
*   **Change Made:**
    *   In the `parsePacketLayers` function, specifically in the `switch` case for `layers.Dot11TypeMgmtBeacon` and `layers.Dot11TypeMgmtProbeResp`:
    *   If `len(originalPayload)` is less than `fixedHeaderLen` (12 bytes), the function now returns `nil, fmt.Errorf(...)` instead of `info, nil`.
*   **Expected Outcome:** This change ensures that `parsePacketLayers` signals a critical error when essential management frames (Beacon, Probe Response) are too short to contain their fixed headers. Consequently, the calling function `ProcessPcapStream` will log this error and skip calling the packet handler for such frames. This will prevent `state_manager` from creating BSS entries based on these fundamentally incomplete and unusable frames, improving data quality.
---
### Decision (Debug - Frontend Code Quality)
[2025-05-07 22:43:00] - [Code Cleanup: Resolve Additional ESLint Unused Variable Warning]

**Rationale:**
After previous ESLint warning fixes, a new run of the frontend build process revealed one more `no-unused-vars` warning in `BssList.tsx` for the `STA` type import. This occurred because the component (`StaListItem`) that utilized this type had been removed from the file in a prior cleanup step.

**Details:**
*   **Affected File:** `web_frontend/src/components/BssList/BssList.tsx`
*   **Change Made:**
    *   Removed the unused import `STA` from `../../types/data`. The import statement was changed from `import { BSS, STA } from '../../types/data';` to `import { BSS } from '../../types/data';`.
*   **Expected Outcome:** All identified ESLint warnings related to unused variables in the specified frontend components are now resolved, contributing to cleaner code.
---
### Decision (Debug - Frontend Code Quality)
[2025-05-07 22:41:00] - [Code Cleanup: Resolve ESLint Unused Variable Warnings]

**Rationale:**
During frontend compilation, ESLint reported several `no-unused-vars` warnings in `StaList.tsx` and `BssList.tsx`. While not critical runtime errors, these warnings indicate dead code or unnecessary imports, which can affect code clarity and maintainability. The decision was to remove the unused code to improve code quality.

**Details:**
*   **Affected Files:**
    *   `web_frontend/src/components/StaList/StaList.tsx`
    *   `web_frontend/src/components/BssList/BssList.tsx`
*   **Changes Made in `web_frontend/src/components/StaList/StaList.tsx`:**
    *   Removed the unused import `BSS` from `../../types/data`.
    *   Removed the unused variable `allStas` (which was an alias for `staList` from `useAppState()`).
*   **Changes Made in `web_frontend/src/components/BssList/BssList.tsx`:**
    *   Removed the unused component definition for `StaListItem`.
    *   Removed the unused interface definition `StaListItemProps`.
*   **Expected Outcome:** ESLint warnings related to unused variables in these components are resolved, leading to cleaner code.
---
### Decision (Debug - Frontend Build Persistence)
[2025-05-07 22:38:00] - [Bug Fix Strategy: Persist `--openssl-legacy-provider` in `package.json`]

**Rationale:**
To avoid manually prepending `NODE_OPTIONS=--openssl-legacy-provider` every time `npm start` is run for `web_frontend` (due to Node.js v23.9.0 OpenSSL compatibility issues with `react-scripts@3.4.4`), the environment variable setting was made persistent by modifying the `start` script in `package.json`.

**Details:**
*   **Affected File:** `web_frontend/package.json`
*   **Change Made:**
    *   The `scripts.start` value was changed from `"react-scripts start"` to `"NODE_OPTIONS=--openssl-legacy-provider react-scripts start"`.
*   **Expected Outcome:** Users can now run `npm start` directly without needing to remember or manually type the `NODE_OPTIONS` flag, simplifying the development workflow while ensuring the OpenSSL legacy provider is used for compatibility.
---
### Decision (Debug - Frontend Build)
[2025-05-07 22:35:00] - [Bug Fix Strategy: Use `--openssl-legacy-provider` for Node.js Crypto Compatibility]

**Rationale:**
The `web_frontend` failed to start using `npm start` with the error `Error: error:0308010C:digital envelope routines::unsupported`. This error is common with newer Node.js versions (v17+ like the project's v23.9.0) where OpenSSL 3 might disable older cryptographic algorithms by default, which are still used by older versions of Webpack or its dependencies (via `react-scripts@3.4.4`).

**Details:**
*   **Affected Component:** `web_frontend` build process.
*   **Command Used:** `NODE_OPTIONS=--openssl-legacy-provider npm start`
*   **Change Made:**
    *   The `NODE_OPTIONS=--openssl-legacy-provider` environment variable was prepended to the `npm start` command. This instructs Node.js to use the legacy OpenSSL provider, which enables support for algorithms that might otherwise be unavailable.
*   **Expected Outcome:** The Webpack development server starts successfully, allowing the frontend application to compile and run. This resolved the immediate build failure. The application compiled with some ESLint warnings related to unused variables, which are non-critical for initial startup.
---
### Decision (Debug)
[2025-05-07 22:16:00] - [Bug Fix Strategy: Filter Multicast/Broadcast DA/RA in Data Frames for STA Creation]

**Rationale:**
User logs indicated that MAC addresses like `01:00:5e:7f:ff:fa` and `01:80:c2:00:00:00` (multicast/broadcast) were being incorrectly counted as STAs. These addresses appeared as the Destination Address (DA) in data frames. The existing `isUnicastMAC` function was correctly identifying these, but it was not applied to the DA (referred to as `parsedInfo.RA` in the data frame logic) when inferring the `staMAC`.

**Details:**
*   **Affected File:** `pc_analyzer/state_manager/manager.go`
*   **Change Made:**
    *   In the `ProcessParsedFrame` function, within the data frame processing section, when `staMAC` is derived from `parsedInfo.RA` (which corresponds to the DA of a frame from an AP, or the RA in other contexts), an `isUnicastMAC(parsedInfo.RA)` check was added.
    *   If `parsedInfo.RA` is not a unicast MAC, `staMAC` is not assigned from it, preventing the creation or update of an STA entry for these non-unicast addresses.
*   **Expected Outcome:** This change will prevent the state manager from creating STA entries for multicast or broadcast MAC addresses encountered as the destination in data frames, leading to a more accurate STA count and representation.
---
### Decision (Debug)
[2025-05-07 18:16:00] - [Bug Fix Strategy: Correct Frontend Data Handling for Web UI Display]

**Rationale:**
Following backend parsing enhancements, the Web UI still failed to display BSS/STA information. Log analysis confirmed the backend was sending correct and more complete data. The issue was then traced to the frontend's handling of this data, specifically incorrect type definitions and state update logic.

**Details:**
*   **Affected Files:**
    *   `web_frontend/src/types/data.ts`
    *   `web_frontend/src/contexts/DataContext.tsx`
    *   `web_frontend/src/components/BssList/BssList.tsx`
*   **Changes Made in `web_frontend/src/types/data.ts`:**
    1.  **Renamed `Station` to `STA`:** The interface for station information was renamed to `STA` to match usage in other parts of the frontend and align with common terminology. Field names within `STA` (e.g., `mac_address`, `signal_strength`, `last_seen`) were updated to match the snake_case format sent by the backend.
    2.  **Corrected `WebSocketData` Structure:** The `WebSocketData` interface was redefined to accurately reflect the nested structure ` { type: string; data: { bsss: BSS[]; stas: STA[] } } ` sent by the backend. This includes the `type` field (e.g., "snapshot") and the nested `data` object containing `bsss` and `stas` arrays.
    3.  **Updated `BSS` interface:** Ensured field names like `signal_strength`, `last_seen`, and `associated_stas` (now a map `{[mac: string]: STA }`) match backend data.
*   **Changes Made in `web_frontend/src/contexts/DataContext.tsx`:**
    1.  **Updated Imports:** Changed import from `Station` to `STA`.
    2.  **Corrected `SET_DATA` Reducer:** The reducer logic for the `SET_DATA` action was modified to correctly access the nested `bsss` and `stas` arrays from `action.payload.data.bsss` and `action.payload.data.stas` respectively. It also now correctly updates both `state.bssList` and `state.staList`.
    3.  **Action Payload Type:** The `SET_DATA` action's payload type in `type Action` was reverted to `WebSocketData` (as defined in `types/data.ts`) because the `handleMessage` function in `useEffect` already receives data of this type.
*   **Changes Made in `web_frontend/src/components/BssList/BssList.tsx`:**
    1.  **Updated Imports:** Changed import from `Station` to `STA`.
    2.  **Corrected Prop Usage:** Updated component to use correct property names from the `BSS` and `STA` types (e.g., `bss.signal_strength`, `sta.mac_address`, `bss.associated_stas`).
    3.  **Handled `associated_stas` Object:** Modified the rendering of associated stations to correctly iterate over the `associated_stas` object (which is a map) using `Object.values()` and `Object.keys().length` for counting.
*   **Expected Outcome:** These frontend corrections should ensure that the data received from the WebSocket is correctly typed, parsed, stored in the application state, and subsequently rendered by the UI components, resolving the issue of BSS/STA information not appearing.
---
### Decision (Debug)
[2025-05-07 18:00:00] - [Bug Fix Strategy: Enhance Backend Parsing for Web UI Data Display]

**Rationale:**
The Web UI was not displaying BSS/STA information, despite logs showing SSIDs being parsed. Analysis indicated that incomplete HT/VHT capabilities and bandwidth information in the data sent to the frontend, along with potential SSID encoding issues, might be contributing factors. The decision was to enhance the backend parsing to provide more complete data to the frontend.

**Details:**
*   **Affected Files:**
    *   `pc_analyzer/frame_parser/parser.go`
    *   `pc_analyzer/state_manager/manager.go`
*   **Changes Made in `pc_analyzer/frame_parser/parser.go`:**
    1.  **Updated `ParsedFrameInfo` struct:** Added `VHTOperationRaw []byte`, `ParsedHTCaps *HTCapabilityInfo`, and `ParsedVHTCaps *VHTCapabilityInfo` fields.
    2.  **Defined `HTCapabilityInfo` and `VHTCapabilityInfo` structs:** To hold parsed HT and VHT parameters.
    3.  **SSID Parsing:** Implemented UTF-8 validation for SSIDs using `unicode/utf8.Valid()`. Invalid SSIDs are now marked as `"<Invalid SSID Encoding>"`.
    4.  **IE Parsing Loop:**
        *   Stored raw bytes of VHT Operation IE (ID 192) into `ParsedFrameInfo.VHTOperationRaw`.
    5.  **Post-IE Loop Parsing Logic:**
        *   **HT Capabilities (ID 45):** If `HTCapabilitiesRaw` exists, parse it to populate `ParsedFrameInfo.ParsedHTCaps` (including `ChannelWidth40MHz`, `ShortGI20MHz`, `ShortGI40MHz`, `SupportedMCSSet`). Bandwidth is initially set based on `ChannelWidth40MHz`.
        *   **VHT Operation (ID 192):** If `VHTOperationRaw` exists, parse its channel width field to update/override `ParsedFrameInfo.Bandwidth` (20MHz, 40MHz, 80MHz, 160MHz, 80+80MHz).
        *   **VHT Capabilities (ID 191):** If `VHTCapabilitiesRaw` exists, parse it to populate `ParsedFrameInfo.ParsedVHTCaps` (including MCS maps, GI settings, SU/MU beamformer capabilities, and supported channel widths). Bandwidth may be further refined based on VHT supported channel widths if not set to a higher value by VHT Operation.
*   **Changes Made in `pc_analyzer/state_manager/manager.go`:**
    1.  **`ProcessParsedFrame` Updated:**
        *   Assign `parsedInfo.Bandwidth` to `bss.Bandwidth` and `sta.Bandwidth` (if STA model had bandwidth).
        *   If `parsedInfo.ParsedHTCaps` is not nil, copy its data to `bss.HTCapabilities` and `sta.HTCapabilities`.
        *   If `parsedInfo.ParsedVHTCaps` is not nil, copy its data to `bss.VHTCapabilities` and `sta.VHTCapabilities`, correctly mapping `ParsedVHTCaps.SupportedChannelWidthSet` (uint8) to the boolean fields (`ChannelWidth80MHz`, `ChannelWidth160MHz`, `ChannelWidth80Plus80MHz`) in `models.VHTCapabilities`.
    2.  Removed the unused `parseCapabilitiesFromRaw` helper function.
*   **Expected Outcome:** The backend should now parse and provide more comprehensive bandwidth and HT/VHT capability information to the frontend, and handle potentially problematic SSIDs more gracefully. This increases the likelihood of the Web UI correctly displaying BSS/STA information.
## Decision

* [2025-05-06 23:48:00] - **数据传输协议选型：路由器到PC端使用gRPC，PC端到Web前端使用WebSocket。**
      
## Rationale

*   **gRPC (路由器 -> PC):**
    *   **高效性:** 基于HTTP/2，支持双向流，适合实时传输原始帧数据。
    *   **跨语言:** Protobuf 定义接口，便于不同语言实现的组件间通信。
    *   **强类型:** 接口定义清晰，减少集成错误。
    *   **性能:** 比自定义TCP协议开发成本低，且有成熟的库支持。
*   **WebSocket (PC -> Web):**
    *   **实时性:** 浏览器原生支持，适合将分析结果实时推送到前端。
    *   **双向通信:** 支持前端发送控制指令回PC端。
    *   **广泛兼容性:** 现代浏览器普遍支持。

## Implementation Details

*   **路由器端抓包代理:** 实现gRPC服务端，流式传输捕获的原始帧数据。
*   **PC端实时分析引擎:**
    *   实现gRPC客户端，接收来自路由器的数据流。
    *   实现WebSocket服务端，向Web前端推送结构化数据，并接收控制指令。
*   **Web前端可视化界面:** 实现WebSocket客户端，接收数据并发送指令。

---
### Decision
[2025-05-06 23:48:00] - **组件核心职责定义**

**Rationale:**
明确各组件的边界和主要功能，确保模块化和关注点分离，便于开发和维护。

**Implications/Details:**
*   **路由器端抓包代理:**
    *   职责: 配置无线网卡的Monitor模式，根据指令启动/停止指定信道和带宽的802.11空口抓包，并将捕获到的原始帧数据（通常带有Radiotap头部）实时传输给PC端分析引擎。
    *   关键技术: `iw`命令, `tcpdump`/`libpcap` (或直接`nl80211`编程), gRPC Server。
*   **PC端实时分析引擎:**
    *   职责: 接收来自路由器的数据流，高速解析802.11帧（包括Radiotap头部），维护BSS与STA的关联状态，提取关键信息（SSID, BSSID, STA MAC, 信道，带宽，安全类型，设备能力等），计算基础指标，并将处理后的结构化数据通过WebSocket推送给Web前端。同时接收前端控制指令并转发给路由器代理。
    *   关键技术: gRPC Client, 802.11帧解析库 (如 `scapy` Python, `gopacket` Go, 或自定义C/C++解析器), WebSocket Server, 状态管理逻辑。
*   **Web前端可视化界面:**
    *   职责: 在浏览器中运行，通过WebSocket接收来自PC端分析引擎的数据，以用户友好的方式（如列表、树状结构、图表）展示BSS、STA及其关联关系和基本信息，并支持实时更新和发送控制指令。
    *   关键技术: HTML, CSS, JavaScript (React/Vue/Angular等框架可选), WebSocket Client, 可视化库 (如 D3.js, Chart.js)。
---
### Decision (Debug)
[2025-05-07] - [Bug Fix Strategy: Adapt PC Engine to handle flexible WebSocket command structures]

**Rationale:**
The PC-side analysis engine was failing to parse WebSocket control commands from the web frontend due to a mismatch in JSON structure. The frontend sent `{"action":"command_name", "payload":{...}}`, while the backend expected `{"command":"command_name", "flat_parameters...}`. To ensure robustness and compatibility with the existing frontend message format (as inferred from logs and `data.ts`), the PC engine's parsing logic was made more flexible rather than requiring immediate frontend changes.

**Details:**
*   **File Affected:** `pc_analyzer/main.go` (specifically the `webSocketControlMessageHandler` function).
*   **Changes Made:**
    *   The Go struct `ControlCommandMsg` used for unmarshalling incoming JSON was modified to include fields for both `action` and `command` keys.
    *   Logic was added to prioritize the `command` key but fall back to the `action` key if `command` is not present or empty.
    *   A nested `CommandPayload` struct was introduced to map the frontend's `payload` object.
    *   The `InterfaceName` field within `CommandPayload` was tagged with `json:"interface,omitempty"` to correctly parse the `interface` field sent by the frontend (as defined in `web_frontend/src/types/data.ts`).
*   **Outcome:** The PC engine can now correctly parse control commands like "start_capture" from the frontend, improving interoperability.
---
### Decision (Debug)
[2025-05-07 11:30:00] - [Bug Fix Strategy: Enforce InterfaceName for WebSocket Start Capture Command]

**Rationale:**
The PC-side analysis engine's WebSocket control message handler (`pc_analyzer/main.go`) was not strictly enforcing the presence of `InterfaceName` within the `payload` for "start_capture" commands. While a previous fix correctly parsed `{"action":"start_capture","payload":{}}` and allowed fallback from "command" to "action" keys, if the `payload` was empty or did not contain a valid `interface` field, the `InterfaceName` would be an empty string. The code logged this but did not prevent the gRPC command from being sent with an empty interface name, potentially leading to downstream failures in the router agent that manifested as "Unknown WebSocket control command" or similar errors reported by the user. The user's feedback "Command sent: start_capture 好像是空的？" suggested the payload or `InterfaceName` was indeed missing.

**Details:**
*   **File Affected:** `pc_analyzer/main.go` (specifically the `webSocketControlMessageHandler` function).
*   **Change Made:** Modified the `webSocketControlMessageHandler` to return an error if a "start_capture" command is received without a non-empty `InterfaceName` in its payload. This was achieved by uncommenting and activating the line: `return fmt.Errorf("START_CAPTURE command requires 'interface' in payload")`.
*   **Expected Outcome:** The PC engine will now explicitly reject "start_capture" commands that lack the required `interface` in their payload, providing a clearer error message and preventing attempts to operate with invalid parameters. This encourages the frontend to always provide a valid interface name.
---
### Decision (Code)
[2025-05-07 11:35:00] - [Web Frontend: Hardcode `interface` in `start_capture` payload]

**Rationale:**
The PC-side analysis engine now mandates an `interface` field within the `payload` of the `start_capture` WebSocket command. To quickly meet this requirement and facilitate immediate testing, the `interface` value in the Web Frontend (`web_frontend/src/components/ControlPanel/ControlPanel.tsx`) has been temporarily hardcoded to `"ath1"`. This value corresponds to the user's known monitor mode interface on their router.

**Details:**
*   **File Affected:** `web_frontend/src/components/ControlPanel/ControlPanel.tsx`
*   **Change Made:** In the `handleSendCommand` function, when the action is `start_capture`, the payload is constructed as `payload = { interface: "ath1", channel: ch, bandwidth: bandwidth };`.
*   **Future Consideration:** While hardcoding `"ath1"` allows for immediate functionality, a more robust long-term solution would involve allowing the user to select or input the desired interface name through the UI. This change defers that UI implementation for now.
*   **Type Definition Update:** The `ControlCommand` interface in `web_frontend/src/types/data.ts` was also updated to include `interface?: string;` in its `payload` definition to align with this change and prevent TypeScript errors.
---
### Decision (Debug)
[2025-05-07 12:16:00] - [gRPC Unimplemented Fix: Align Client Proto `go_package` for Correct Service Name]

**Rationale:**
The PC analysis engine (gRPC client) encountered an "Unimplemented" error (`unknown service router_agent_pb.CaptureAgent`) when calling the router agent (gRPC server). The root cause was a mismatch in the gRPC service name used by the client versus the name registered by the server.
- The server (`router_agent`) uses `package router_agent;` and `option go_package = ".;main";` in its `capture_agent.proto`. This results in the service `router_agent.CaptureAgent` being registered.
- The client (`pc_analyzer`) used `package router_agent;` and `option go_package = "...;router_agent_pb";` in its `capture_agent.proto`. This caused the `protoc-gen-go-grpc` tool to generate client-side code (specifically the `ServiceDesc` and `FullMethodName` constants in `pc_analyzer/router_agent_pb/capture_agent_grpc.pb.go`) that referenced the service as `router_agent_pb.CaptureAgent`.

The fix involves aligning the client's `go_package` option to ensure the generated code uses the correct service name derived from the protobuf package declaration (`router_agent.CaptureAgent`).

**Details:**
*   **File Affected by Fix:** `pc_analyzer/capture_agent.proto`
*   **Change Made:** The `option go_package` in `pc_analyzer/capture_agent.proto` was changed from `wifi-pcap-demo/pc_analyzer/router_agent_pb;router_agent_pb` to `wifi-pcap-demo/pc_analyzer/router_agent_pb;router_agent`.
*   **Required Follow-up:** The gRPC client code in `pc_analyzer` (specifically `pc_analyzer/router_agent_pb/capture_agent.pb.go` and `pc_analyzer/router_agent_pb/capture_agent_grpc.pb.go`) must be regenerated using `protoc` after this `.proto` file change. This will ensure the client attempts to connect to the correct `router_agent.CaptureAgent` service.
*   **Affected Components:** PC Analysis Engine (gRPC client), Router Agent (gRPC server interaction).
---
### Decision (Debug & Refactor)
[2025-05-07 13:24:06] - [PC Engine: Adopt `pcapgo` for parsing gRPC-streamed pcap data]

**Rationale:**
The PC analysis engine was encountering `Error parsing frame: radiotap layer not found`. This was because `tcpdump -w -` (used by the router agent) outputs data in pcap format, which was streamed via gRPC. The PC engine's gRPC client was receiving chunks of this pcap stream but attempting to parse each chunk directly as a raw 802.11 frame using `gopacket.NewPacket(data, layers.LayerTypeRadioTap, ...)`. This approach fails because pcap streams have their own file/global headers and per-packet record headers that must be processed before accessing the raw frame data.

To correctly handle this, the decision was made to process the incoming gRPC data as a continuous pcap stream.

**Details:**
*   **Affected Components:** `pc_analyzer/grpc_client/client.go`, `pc_analyzer/frame_parser/parser.go`, `pc_analyzer/main.go`.
*   **Key Changes:**
    1.  **`pc_analyzer/grpc_client/client.go`:** Modified to use an `io.Pipe`. Bytes received from the gRPC stream are written to the `pipeWriter`. The `pipeReader` is passed to a new handler dedicated to pcap stream processing. This allows for true stream processing without buffering the entire pcap data in memory. The `PacketHandler` type was changed to `PcapStreamHandler func(pcapStream io.Reader)`.
    2.  **`pc_analyzer/frame_parser/parser.go`:** A new function, `ProcessPcapStream(pcapStream io.Reader, pktHandler PacketInfoHandler)`, was introduced. (`PacketInfoHandler` is `func(info *ParsedFrameInfo)`). This function uses `github.com/google/gopacket/pcapgo.NewReader` to read from the provided `io.Reader` (the `pipeReader`).
    3.  Inside `ProcessPcapStream`, `pcapReader.ReadPacketData()` is called in a loop to extract individual packet data and capture information.
    4.  `gopacket.NewPacket()` is then called with the raw packet data obtained from `pcapReader.ReadPacketData()` and the `LinkType` obtained from `pcapReader.LinkType()`. This ensures correct parsing of each packet.
    5.  The core logic for parsing Radiotap and Dot11 layers (previously in the old `ParseFrame` function) was moved into a new helper `parsePacketLayers` and adapted. The old `ParseFrame` was commented out.
    6.  **`pc_analyzer/main.go`:** Imported `io`. The packet handling logic was refactored. An `packetInfoHandler` (of type `frame_parser.PacketInfoHandler`) and a `pcapStreamHandler` (of type `grpc_client.PcapStreamHandler`) were defined and initialized. The `pcapStreamHandler` calls `frame_parser.ProcessPcapStream`, passing the `packetInfoHandler`. Calls to `grpcClient.StreamPackets` were updated to use `pcapStreamHandler`.
*   **Libraries Used:** `github.com/google/gopacket/pcapgo` for pcap stream reading.
*   **Outcome:** This approach allows the PC engine to correctly interpret the pcap stream, identify individual packets, and parse them, thus resolving the "radiotap layer not found" error and enabling proper frame analysis.
---
### Decision (Debug)
[2025-05-07 14:12:00] - [Bug Fix Strategy: Enhance SSID IE Parsing in PC Analyzer]

**Rationale:**
The PC analysis engine (`pc_analyzer/frame_parser/parser.go`) was failing to extract SSIDs from Beacon and Probe Response frames, resulting in empty SSID fields in the data sent to the frontend. Log analysis and code review indicated that the existing Information Element (IE) parsing loop might not be correctly identifying or processing the SSID IE (Element ID 0), or that the specific debug log for SSID IE detection was not being reached. The fix aims to improve logging for IE iteration and specifically enhance the handling of the SSID IE.

**Details:**
*   **Affected File:** `pc_analyzer/frame_parser/parser.go` (specifically the `parsePacketLayers` function).
*   **Changes Made:**
    1.  **Added General IE Iteration Log:** Inserted `log.Printf("DEBUG_IE_ITERATION: IE ID: %d, IE Length: %d", ieID, ieLength)` before the `switch ieID` statement within the IE processing loop. This helps confirm if all IEs, including SSID (ID 0), are being iterated over.
    2.  **Enhanced SSID IE Handling:**
        *   Modified the `case layers.Dot11InformationElementIDSSID:` block.
        *   If the SSID IE's length (`ieLength`) is 0 (indicating a hidden SSID), the `info.SSID` field is now explicitly set to the string `"<Hidden SSID>"`.
        *   If `ieLength` is greater than 0, `info.SSID` is set to `string(ieInfo)` as before.
        *   The debug log for successful SSID IE parsing was updated to `log.Printf("DEBUG_SSID_PARSE: Found SSID IE for BSSID %s. Length: %d, SSID: [%s], Hex: %x", bssidForLog, ieLength, ssidContent, ieInfo)`, providing more context like BSSID, the parsed SSID string, its length, and hex representation.
*   **Expected Outcome:** These changes should ensure that SSIDs are correctly extracted from Beacon and Probe Response frames. The new logs will provide clearer insight into the IE parsing process, and hidden SSIDs will be explicitly marked.

---
### Decision (Debug)
[2025-05-07 14:46:00] - [Bug Fix Strategy: Simplify SSID IE Parsing in PC Analyzer to use Dot11.Payload exclusively]

**Rationale:**
Previous attempts to parse Information Elements (IEs) by accessing structured fields like `ApplicationLayer().InformationElements` or `Dot11MgmtBeacon.InformationElements` within `pc_analyzer/frame_parser/parser.go` consistently led to compilation errors. This indicated a likely mismatch between the assumed `gopacket` API/version and the one used in the project, or a misunderstanding of how `gopacket` exposes these parsed elements.
The user reported that `DEBUG_SSID_PARSE` logs were not appearing, while `DEBUG_IE_ITERATION` logs (presumably from iterating `dot11.Payload`) were. This suggested that the fundamental issue was the failure to enter the SSID parsing case, likely because the complex structured IE access attempts were failing or were based on incorrect assumptions.
To ensure a robust and compilable solution, the decision was made to revert the IE parsing logic to *exclusively* and directly iterate over the `dot11.Payload` field of the `*layers.Dot11` struct. This is a standard fallback mechanism in `gopacket` and aligns with the original code's approach to IE processing.

**Details:**
*   **Affected File:** `pc_analyzer/frame_parser/parser.go` (specifically the `parsePacketLayers` function).
*   **Changes Made:**
    1.  Removed all logic that attempted to access IEs via `dot11.ApplicationLayer()` or by type-asserting to specific management frame types (e.g., `*layers.Dot11MgmtBeacon`) and then accessing an `InformationElements` field.
    2.  The IE parsing loop now solely iterates over `dot11.Payload` (the `[]byte` field of `*layers.Dot11`).
    3.  Ensured that the `switch` statement within this loop correctly handles `layers.Dot11InformationElementIDSSID` and other relevant IEs (Rates, DSSet, TIM, HTCapabilities, VHTCapabilities, RSNInfo using the correct `layers.Dot11InformationElementIDRSNInfo` constant).
    4.  Corrected minor issues like using `dot11.Payload` as a field instead of a method call `dot11.Payload()`.
    5.  Maintained and verified the placement of `DEBUG_IE_ITERATION` and `DEBUG_SSID_PARSE` logs within this simplified payload parsing loop.
    6.  The entire file was rewritten using `write_to_file` to ensure a clean, compilable state after multiple problematic `apply_diff` attempts.
*   **Expected Outcome:** The `pc_analyzer/frame_parser/parser.go` file is now syntactically correct and compilable. The SSID parsing logic should correctly identify and process SSID IEs if they are present in the `dot11.Payload` of Beacon and Probe Response frames, leading to the appearance of `DEBUG_SSID_PARSE` logs and correct SSID population. This directly addresses the user's reported issue by ensuring the relevant code path for SSID parsing is reachable.
---
### Decision (Debug)
[2025-05-07 15:01:00] - [Bug Fix Strategy: Correct IE Parsing Offset for Beacon/ProbeResp Frames]

**Rationale:**
SSID parsing was failing for Beacon and Probe Response frames because the Information Element (IE) parsing logic in `pc_analyzer/frame_parser/parser.go` was processing the entire `dot11.Payload` without accounting for the fixed-length fields (Timestamp, Beacon Interval, Capability Info) at the beginning of these specific management frames. This caused the parser to misinterpret these fixed fields as IEs, leading to incorrect IE ID/Length decoding and premature termination of the IE loop, thus missing the actual SSID IE. The fix involves identifying these frame types and applying a 12-byte offset to the payload before IE parsing.

**Details:**
*   **Affected File:** `pc_analyzer/frame_parser/parser.go` (specifically the `parsePacketLayers` function).
*   **Changes Made:**
    1.  Within the `if dot11.Type.MainType() == layers.Dot11TypeMgmt` block, a `switch` statement was added for `dot11.Type`.
    2.  For `case layers.Dot11TypeMgmtBeacon` and `case layers.Dot11TypeMgmtProbeResp`:
        *   The `dot11.Payload` is sliced to skip the first 12 bytes (i.e., `originalPayload[12:]`) before passing it to the IE parsing loop.
        *   Logging was added to indicate when this offset is applied.
    3.  The IE parsing loop now iterates over this correctly offsetted `iePayload`.
*   **Expected Outcome:** SSIDs from Beacon and Probe Response frames should now be parsed correctly, as the IE parsing logic will operate on the actual sequence of IEs, resolving the "SSID: N/A" issue and ensuring `DEBUG_SSID_PARSE` logs appear as expected.
---
### Decision
[2025-05-07 15:21:57] - 改进 IE 解析逻辑以应对数据不足和异常长度，并为 `MgmtMeasurementPilot` 和 `MgmtActionNoAck` 等帧类型实现 payload 偏移。

**Rationale:**
日志表明 IE 解析因数据不足中断，特定管理帧可能未正确处理头部固定字段。这导致SSID等关键信息无法正确提取。

**Implications/Details:**
*   **IE 解析循环健壮性:** 修改 `pc_analyzer/frame_parser/parser.go` 中的IE解析循环，在每次迭代前检查声明的IE长度（`ieLength`）是否超过了帧内剩余的可用数据量（`availableDataAfterIDLen`）。如果超出，则记录警告并中断当前帧的IE解析，以防止因单个格式错误的IE导致整个帧的解析失败。
*   **特定管理帧偏移研究与应用:**
    *   针对 `MgmtMeasurementPilot` 帧，需根据 IEEE 802.11 标准研究其帧体结构，确定在信息元素字段开始前是否存在固定长度的字段。如果存在，则在解析IE之前应用此偏移量。
    *   针对 `MgmtActionNoAck` 帧（以及其他可能相关的Action帧子类型），同样需要根据 IEEE 802.11 标准研究其特定类别/动作的帧格式，确定固定头部长度并应用相应偏移。
    *   此研究结果将用于更新 `pc_analyzer/frame_parser/parser.go` 中 `parsePacketLayers` 函数内相应帧类型的处理逻辑。
*   **日志增强:** 配合上述修改，增强相关日志，包括IE解析中断、偏移应用等情况，便于调试和验证。

---
### Decision (Debug - Test Fix)
[2025-05-07 17:15:00] - [Use `layers.LinkType(127)` in Tests for `LinkTypeRadioTap`]

**Rationale:**
During compilation fixes for `pc_analyzer/frame_parser/parser_test.go` after modifying the `parsePacketLayers` function signature, the compiler reported `undefined: layers.LinkTypeRadioTap`. While the constant `LinkTypeRadioTap` with value 127 is expected to exist in `gopacket/layers`, it could not be resolved in the test environment for unknown reasons (potentially build tags, module inconsistencies, or version issues). To allow the tests to compile and proceed with debugging the main application logic, the decision was made to use the explicit type conversion `layers.LinkType(127)` instead of the constant name `layers.LinkTypeRadioTap` within the test file.

**Details:**
*   **Affected File:** `pc_analyzer/frame_parser/parser_test.go`
*   **Change Made:** Replaced all occurrences of `layers.LinkTypeRadioTap` with `layers.LinkType(127)` when calling the `parsePacketLayers` function.
*   **Implication:** This is a workaround. The underlying reason why the constant name is undefined in the test context should ideally be investigated later. However, using the correct integer value ensures the test logic uses the appropriate `LinkType` for Radiotap frames.
*   **Outcome:** The test file `pc_analyzer/frame_parser/parser_test.go` now compiles successfully.

---
### Decision (Debug)
[2025-05-07 21:15:35] - [Bug Fix Strategy: Update STA Signal Strength from Data Frames]

**Rationale:**
User reported numerous 0dBm STA entries. Analysis of `pc_analyzer/state_manager/manager.go` revealed that when STAs are identified or created based on data frames, their signal strength was not being updated from `parsedInfo.SignalStrength`. If `NewSTAInfo()` initializes signal strength to 0, these STAs would remain at 0dBm unless later updated by a management frame with non-zero signal.

**Details:**
*   **Affected File:** `pc_analyzer/state_manager/manager.go`
*   **Change Made:** In the `ProcessParsedFrame` function, within the section handling data frames (specifically after a STA is confirmed or newly created, around line 352), a block was added to update `sta.SignalStrength` if `parsedInfo.SignalStrength` is not zero:
    ```go
    // Update signal strength if available from the data frame context
    if parsedInfo.SignalStrength != 0 {
        sta.SignalStrength = parsedInfo.SignalStrength
    }
    ```
*   **Expected Outcome:** STAs identified through data frames will now have their signal strength populated if the parsed frame information contains a non-zero signal value. This should reduce the number of STAs appearing with 0dBm signal strength.

---
### Decision (Debug - gopacket Parsing Errors)
[2025-05-08 17:42:00] - [Strategy: Investigate and Mitigate gopacket Parsing Errors Affecting Metrics]

**Rationale:**
Extensive `gopacket` parsing errors (e.g., `Dot11 length X too short`, `ERROR_NO_DOT11_LAYER`) are preventing the extraction of `Duration/ID` and `TransportPayloadLength`, leading to "N/A" metrics on the frontend. A multi-pronged approach is needed to diagnose and address these issues.

**Details:**
*   **Affected Components:**
    *   PC端分析引擎: [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0)
    *   Potentially the raw pcap data itself.
*   **Key Decisions & Investigation Paths:**
    1.  **Prioritize Wireshark Analysis:** The primary step is for the user to analyze the problematic pcap files with Wireshark. This will help determine if the packets are inherently malformed/truncated or if the issue lies primarily with `gopacket`'s interpretation. Wireshark's "Expert Information" and byte-level inspection are crucial.
    2.  **Enhance Parser Robustness (Iterative):**
        *   **Stricter Error Handling:** In [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0), if `gopacket.NewPacket` returns an error in `ErrorLayer()`, or if the `Dot11` layer cannot be decoded, the `parsePacketLayers` function should more definitively return an error to `ProcessPcapStream`. This will ensure that critically flawed packets are skipped entirely and do not lead to partially processed data being used for metric calculation.
        *   **Review `gopacket` Decoding Options:** Investigate if alternative `DecodeOptions` for `gopacket.NewPacket` (e.g., `gopacket.Lazy`) could offer better resilience against certain types of corruption, while still allowing necessary fields to be parsed. This requires careful testing to avoid unintended side effects (like layers not being parsed when needed).
    3.  **Radiotap Integrity Check:** Given the `ERROR_NO_DOT11_LAYER: Dot11 layer is nil. Radiotap present: true.` errors, special attention should be paid to the Radiotap headers in Wireshark. Incorrect Radiotap length or field values could mislead `gopacket`.
    4.  **Iterative Refinement:** Based on Wireshark findings, further refinements to the parsing logic in [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0) may be necessary. This could include more specific handling for certain frame types or error conditions if patterns emerge.
*   **Expected Outcome:**
    *   Clearer understanding of whether the root cause is malformed pcap data or `gopacket` parsing limitations.
    *   Improved robustness of the packet parser to gracefully handle or skip malformed packets.
    *   Reduction in parsing errors, leading to more successful extraction of data needed for metric calculations, ultimately resolving the "N/A" display on the frontend.
---
### Decision (Architecture - Frame Parsing Engine)
[2025-05-10] - [Architectural Shift: Replace `gopacket` with `tshark` for 802.11 Frame Parsing]

**Rationale:**
The existing `gopacket`-based parsing logic in [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0) has encountered significant difficulties in robustly handling various 802.11 frames, particularly those that might be malformed, non-standard, or contain complex vendor-specific information elements. This has led to parsing errors (e.g., "Dot11 length X too short", "ERROR_NO_DOT11_LAYER") and an inability to reliably extract all necessary data for metric calculation (throughput, channel utilization), resulting in "N/A" values on the frontend.

`tshark`, the command-line interface for Wireshark, possesses a highly mature, extensively tested, and robust packet dissection engine. By leveraging `tshark -T fields`, we can:
1.  Delegate the complexities of low-level frame parsing to a specialized and reliable tool.
2.  Precisely specify the exact fields required for analysis, minimizing data processing overhead.
3.  Improve resilience against malformed or non-standard frames, as `tshark` is generally more fault-tolerant.
4.  Simplify the Go-based parser's role to orchestrating `tshark` execution and parsing its structured CSV output.

**Details:**
*   **Affected Components:**
    *   PC端分析引擎: [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0) will be significantly refactored.
    *   Configuration: May require adding `tshark` path to [`desktop_app/WifiPcapAnalyzer/config/config.json`](desktop_app/WifiPcapAnalyzer/config/config.json:0).
*   **Architectural Changes:**
    1.  **`TSharkExecutor`:** A new component responsible for launching and managing the `tshark` child process. It will configure `tshark` with the pcap file path, the extensive list of required fields (specified in the "规范：使用 tshark 输出替换 gopacket 解析"), and output formatting options (`-T fields -E header=y -E separator=, -E quote=d -E occurrence=a`). It will provide the `tshark` standard output (CSV stream) for further processing and monitor its standard error for issues.
    2.  **`CSVParser`:** A component to read the CSV stream from `tshark`. It will parse the header row to map field names to column indices and then read subsequent data rows.
    3.  **`FrameProcessor`:** This component will take a parsed CSV row (map of field names to string values) and convert it into the existing `ParsedFrameInfo` struct. This involves type conversions, handling missing fields, and potentially deriving some values (e.g., `FrameType`/`SubType` from `wlan.fc.type_subtype`).
    4.  **`PhyRateCalculator`:** The existing logic for PHY rate calculation will be adapted to use input fields provided by `tshark`'s output.
    5.  **Main Workflow:** The `ProcessPcapStream` function in `parser.go` will be rewritten. Instead of using `pcapgo` and `gopacket` directly on the pcap data, it will:
        *   Use `TSharkExecutor` to run `tshark` on the input pcap file.
        *   Pipe `tshark`'s output to `CSVParser`.
        *   For each row from `CSVParser`, use `FrameProcessor` to create a `ParsedFrameInfo`.
        *   Pass the `ParsedFrameInfo` to the existing `PacketInfoHandler` callback.
*   **Field Extraction:** A comprehensive list of fields to be extracted via `tshark -e <field>` has been defined in the specification document, covering frame basics, MAC addresses, Radiotap info, BSS/STA details, and parameters for throughput/channel utilization.
*   **Error Handling:** Robust error handling will be implemented at each stage: `tshark` execution, CSV parsing, and individual field processing within `FrameProcessor`.
*   **Integration:** The `ParsedFrameInfo` struct will be largely preserved to minimize impact on the `StateManager` and other downstream components. The `PacketInfoHandler` interface remains the same.
*   **Expected Outcome:**
    *   Increased robustness and reliability of 802.11 frame parsing.
    *   More accurate extraction of a wider range of specified fields.
    *   Reduction in parsing-related errors that currently lead to "N/A" metrics.
    *   Simplification of the Go parsing code by offloading complex dissection to `tshark`.

**Memory Bank Update:**
*   [`memory-bank/developmentContext/pcAnalysisEngine.md`](memory-bank/developmentContext/pcAnalysisEngine.md:1) will be updated with a new section detailing this `tshark`-based parsing architecture.
---
---
### Decision (Debug)
[2025-05-10 16:01:00] - [Fix tshark Field Names for Robust Parsing]

**Rationale:**
The `tshark` process was failing to start due to invalid field names in its command-line arguments. User-provided logs indicated specific fields were not recognized by `tshark`. This fix updates these incorrect field names to their valid equivalents based on analysis of `tshark`'s typical field naming conventions and a provided `tshark_beacon_example.json` example.

**Details:**
*   **Affected components/files:**
    *   [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go)
*   **Field Name Changes:**
    *   `wlan.flags.retry` changed to `wlan.fc.retry`
    *   `radiotap.mcs.fmt` changed to `radiotap.mcs.flags`
    *   `radiotap.he.data.mcs` changed to `radiotap.he.mcs`
    *   `radiotap.he.data.bw` changed to `radiotap.he.bw`
    *   `radiotap.he.data.gi` changed to `radiotap.he.gi`
    *   `radiotap.he.data.spatial_streams` changed to `radiotap.he.nss`
    *   `wlan.ext_tag.he_operation.bss_color` changed to `wlan.ext_tag.bss_color_information.bss_color`
    *   `wlan.ext_tag.he_phy_cap.chan_width_set` changed to `wlan.he.phy.channel_width_set` (attempted)
*   **Unaffected Fields (Assumed Correct or Data Dependent):**
    *   `radiotap.vht.nss` (kept as is, likely valid but data might not always be present)
    *   `radiotap.vht.mcs` (kept as is, likely valid but data might not always be present)
*   **Impact:** These changes are expected to resolve the `tshark` startup errors related to invalid fields, allowing the packet parsing process to proceed correctly. This ensures that all required information for downstream analysis (like HE/VHT parameters, retry flags) can be extracted if present in the capture.