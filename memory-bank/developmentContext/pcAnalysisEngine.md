# PC Analysis Engine - Development Context

This document details the development progress, specific challenges, and solutions related to the PC-side Real-time Analysis Engine.

## Initial Setup & Core Logic (2025-05-07 00:46:00 - 2025-05-07 01:44:00)

*   **Project Skeleton:** Created Go project structure in `pc_analyzer/`.
*   **gRPC Integration:**
    *   Copied `capture_agent.proto` from `router_agent/`.
    *   Updated `go_package` option in `pc_analyzer/capture_agent.proto` to `wifi-pcap-demo/pc_analyzer/router_agent_pb;router_agent_pb` (later identified as a source of gRPC service name mismatch and corrected).
    *   Generated Go gRPC code into `pc_analyzer/router_agent_pb/`.
    *   Implemented gRPC client logic in `pc_analyzer/grpc_client/client.go` for receiving streamed packet data.
*   **Configuration:** Implemented config loading from `config.json` via `pc_analyzer/config/config.go`.
*   **Frame Parsing (Initial):**
    *   Implemented initial 802.11 frame parsing logic in `pc_analyzer/frame_parser/parser.go` using `gopacket`.
    *   Focused on Radiotap and Dot11 layers to extract basic info (BSSID, SA, DA, channel, signal strength).
*   **State Management:** Implemented BSS/STA state management in `pc_analyzer/state_manager/manager.go` and `models.go`.
*   **WebSocket Server:** Implemented WebSocket server in `pc_analyzer/websocket_server/server.go` for:
    *   Pushing processed BSS/STA data to the web frontend.
    *   Receiving control commands from the web frontend.
*   **Main Orchestration:** Integrated all modules in `pc_analyzer/main.go`.

## Debugging & Refinements

### WebSocket Control Command Parsing (Multiple Iterations)

*   **Issue (Initial):** Engine failed to parse commands like `{"action":"start_capture","payload":{...}}` sent by the frontend. Expected a different structure.
*   **Fix 1:** Modified `pc_analyzer/main.go` (`webSocketControlMessageHandler`) to:
    *   Support both "action" and "command" keys for the command name.
    *   Handle nested `payload` structure.
    *   Correctly parse `interface` field from payload.
*   **Issue (Follow-up):** "start_capture" command was processed even with an empty/missing `interface` in the payload, leading to downstream errors. User reported "Command sent: start_capture 好像是空的？".
*   **Fix 2:** Updated `webSocketControlMessageHandler` in `pc_analyzer/main.go` to explicitly return an error if `InterfaceName` is missing in the "start_capture" payload.
    *   **Decision Log:** [2025-05-07 11:30:00] - Enforce InterfaceName for WebSocket Start Capture Command.
*   **Issue (Final Diagnosis - Empty Command String):** User reported "Unknown WebSocket control command:" with an empty command string.
*   **Fix 3 (Logging):** Added detailed debug logging in `webSocketControlMessageHandler` ( `pc_analyzer/main.go`) after JSON unmarshalling and before command dispatch to inspect the actual parsed command and payload. This helped confirm that the command string itself was sometimes empty, likely due to an issue in how the frontend was sending it or a parsing glitch.

### gRPC "Unimplemented" Error (Client-Server Service Name Mismatch)

*   **Issue:** PC Analysis Engine (client) failed to connect to Router Agent (server) with gRPC error: `rpc error: code = Unimplemented desc = unknown service router_agent_pb.CaptureAgent`.
*   **Root Cause:**
    *   Router Agent registered its service as `router_agent.CaptureAgent` (due to `package router_agent;` and `option go_package = ".;main";` in its `.proto`).
    *   PC Analyzer's `pc_analyzer/capture_agent.proto` had `option go_package = "...;router_agent_pb";`, causing its generated gRPC client code to look for `router_agent_pb.CaptureAgent`.
*   **Fix:**
    *   Modified `option go_package` in `pc_analyzer/capture_agent.proto` to `wifi-pcap-demo/pc_analyzer/router_agent_pb;router_agent`.
    *   **Action Required:** The gRPC Go code in `pc_analyzer/router_agent_pb/` needs to be regenerated using `protoc` for this change to take effect.
*   **Decision Log:** [2025-05-07 12:16:00] - Align Client Proto `go_package` for Correct Service Name.

### Frame Parsing: "radiotap layer not found"

*   **Issue:** The parser frequently logged `Error parsing frame: radiotap layer not found` when processing data streamed from the router agent (which uses `tcpdump -w -`).
*   **Root Cause:** `tcpdump -w -` outputs data in pcap format (including global pcap headers and per-packet pcap record headers). The gRPC client was receiving chunks of this pcap stream but `ParseFrame` was attempting to interpret each raw chunk directly as a `RadioTap` + `Dot11` frame without handling the pcap encapsulation.
*   **Fix (Refactor to pcap stream processing):**
    1.  **`pc_analyzer/grpc_client/client.go`:**
        *   Modified `StreamPackets` to use an `io.Pipe`. Bytes from the gRPC stream are written to `pipeWriter`.
        *   The `pipeReader` is passed to a new handler type: `PcapStreamHandler func(pcapStream io.Reader)`.
    2.  **`pc_analyzer/frame_parser/parser.go`:**
        *   Introduced `ProcessPcapStream(pcapStream io.Reader, pktHandler PacketInfoHandler)`.
        *   This function uses `github.com/google/gopacket/pcapgo.NewReader` to read from the `pipeReader`.
        *   It loops, calling `pcapReader.ReadPacketData()` to get individual packet data and `ci.Timestamp`.
        *   `gopacket.NewPacket()` is then called with this data and `pcapReader.LinkType()`.
        *   The core parsing logic was moved to a new helper `parsePacketLayers(packet gopacket.Packet, captureTimestamp time.Time) (*ParsedFrameInfo, error)`.
        *   The old `ParseFrame(rawData []byte)` was commented out.
    3.  **`pc_analyzer/main.go`:**
        *   Adapted the main data processing loop to use the new `PcapStreamHandler` and `PacketInfoHandler` types.
*   **Libraries:** Added `github.com/google/gopacket/pcapgo`.
*   **Outcome:** Resolved the "radiotap layer not found" error, enabling correct parsing of packets from the pcap stream.
*   **Decision Log:** [2025-05-07 13:24:06] - PC Engine: Adopt `pcapgo` for parsing gRPC-streamed pcap data.

### SSID Parsing Issue in Beacon/Probe Response Frames

*   **Issue (2025-05-07 14:04:00):** All BSSs sent to the frontend have an empty `ssid` field. `DEBUG_FRAME_PARSER` logs show SSID as `N/A` for Beacon and Probe Response frames, even though Wireshark confirms the presence of SSID Information Elements (IEs) with valid content.
*   **Analysis:**
    *   The `parsePacketLayers` function in `pc_analyzer/frame_parser/parser.go` iterates through IEs in management frames.
    *   It uses `case layers.Dot11InformationElementIDSSID:` to identify the SSID IE.
    *   The existing debug log `DEBUG_FRAME_PARSER_SSID_IE` was not appearing in user-provided logs, suggesting the SSID IE was not being correctly identified or its content was misinterpreted.
    *   The logic `info.SSID = string(ieInfo)` directly converts IE bytes to string. If `ieInfo` was empty or the case wasn't hit, `info.SSID` would be empty, leading to "N/A" in the final `DEBUG_FRAME_PARSER` log.
*   **Root Cause Hypothesis:** The `case layers.Dot11InformationElementIDSSID:` might not be matching correctly, or an issue occurs before this specific log line.
*   **Fix (Applied 2025-05-07 14:09:00):**
    1.  **Added General IE Iteration Log:** In `pc_analyzer/frame_parser/parser.go`, within the IE loop in `parsePacketLayers`, added `log.Printf("DEBUG_IE_ITERATION: IE ID: %d, IE Length: %d", ieID, ieLength)` before the `switch` statement to verify all encountered IEs.
    2.  **Enhanced SSID IE Handling:**
        *   Modified the `case layers.Dot11InformationElementIDSSID:` block:
            *   If `ieLength == 0` (hidden SSID), `info.SSID` is set to `"<Hidden SSID>"`.
            *   Otherwise, `info.SSID = string(ieInfo)`.
        *   Updated the specific SSID parsing debug log to: `log.Printf("DEBUG_SSID_PARSE: Found SSID IE for BSSID %s. Length: %d, SSID: [%s], Hex: %x", bssidForLog, ieLength, ssidContent, ieInfo)`, including BSSID for better context.
*   **Expected Outcome:** The new logs should clarify if the SSID IE (ID 0) is being seen by the parser. The enhanced handling will correctly assign SSIDs, including a placeholder for hidden ones.
---
### SSID Parsing Issue (Beacon/ProbeResp Fixed Fields) - 2025-05-07

**Problem:**
SSIDs were not being correctly parsed from Beacon and Probe Response frames by `pc_analyzer/frame_parser/parser.go`. The parsed SSID often appeared as "N/A" or empty, and `DEBUG_SSID_PARSE` logs were missing for these frames. Log messages like `DEBUG_IE_ITERATION: Breaking IE loop. Insufficient data for IE ID ...` suggested that the Information Element (IE) parsing loop was terminating prematurely.

**Analysis:**
The root cause was that the IE parsing logic directly used `dot11.Payload` as the source for IEs. However, for `layers.Dot11MgmtBeacon` and `layers.Dot11TypeMgmtProbeResp` frames, the `dot11.Payload` does not start directly with IEs. Instead, it begins with fixed-length fields:
*   Timestamp: 8 bytes
*   Beacon Interval: 2 bytes
*   Capability Information: 2 bytes
Total: 12 bytes.

The existing IE parsing loop was incorrectly interpreting these initial 12 bytes as IE headers (ID and Length). This led to incorrect length calculations for subsequent (non-existent at that position) IEs, causing the loop to break before reaching the actual IE sequence, including the SSID IE (Element ID 0).

**Fix Implementation (`pc_analyzer/frame_parser/parser.go`):**
The `parsePacketLayers` function was modified as follows:
1.  After obtaining `dot11Layer := packet.Layer(layers.LayerTypeDot11)` and `dot11 := dot11Layer.(*layers.Dot11)`, and before the IE parsing loop, a `switch` statement was introduced based on `dot11.Type`.
2.  **For `layers.Dot11TypeMgmtBeacon` and `layers.Dot11TypeMgmtProbeResp`:**
    *   The `originalPayload := dot11.Payload` is sliced to skip the first 12 bytes: `iePayload = originalPayload[12:]`. This `iePayload` is then used for IE parsing.
    *   A check `if len(originalPayload) >= 12` is performed to prevent out-of-bounds access on unexpectedly short payloads.
    *   Debug logging (`DEBUG_MGMT_PAYLOAD_OFFSET`) was added to confirm when this 12-byte offset is applied.
3.  **For other management frame types:** The `iePayload` defaults to the `originalPayload` (no offset), maintaining previous behavior for frames like Probe Requests, which typically start directly with IEs.
4.  The IE parsing loop (`for len(currentIEPayload) >= 2`) was updated to iterate over this `currentIEPayload` (which is initialized from the potentially offsetted `iePayload`).

**Expected Result:**
With this change, the IE parsing logic will operate on the correct byte slice (starting after the fixed fields for Beacon and Probe Response frames). This should allow for correct identification and parsing of the SSID IE and other IEs, resolving the "SSID: N/A" issue and ensuring the `DEBUG_SSID_PARSE` logs appear as intended.
---
## SSID 解析问题：IE 数据不足及特定管理帧偏移 (MgmtMeasurementPilot, MgmtActionNoAck)

**Timestamp:** 2025-05-07 15:22:25

**问题分析:**
用户提供的日志表明，SSID 解析失败以及 BSS/STA 信息不在 Web UI 上显示的主要原因是 PC 端分析引擎在解析 802.11 信息元素 (IE) 时遇到问题。具体表现为：
1.  **IE 解析因数据不足中断:** 对于某些 IE (如日志中的 ID 23 和 ID 186)，其声明的长度超出了帧中实际可用的数据量。这导致 IE 解析循环提前终止，后续的 IE (可能包括 SSID IE) 未被处理。此问题出现在 `MgmtMeasurementPilot` 和 `MgmtActionNoAck` 类型的帧中。
2.  **部分帧类型 SSID 为 N/A:** 日志中 `MgmtMeasurementPilot` 和 `MgmtReassociationReq` 帧的 SSID 字段显示为 "N/A"，表明未能成功提取 SSID。

**核心原因推断:**
*   IE 解析逻辑对异常长度或数据不足的 IE 处理不够健壮。
*   部分管理帧类型 (如 `MgmtMeasurementPilot`, `MgmtActionNoAck`) 可能没有像 Beacon/ProbeResp 帧那样正确处理其头部固定字段的偏移量，导致 IE 解析从错误的位置开始。

**修复规范与伪代码 (摘要):**

**目标:** 提高 IE 解析的鲁棒性，确保即使在遇到格式错误的单个 IE 时也能最大限度地解析其他有效 IE，并为所有相关的管理帧类型正确处理 payload 偏移，以准确提取 SSID 等信息。

**影响模块:** `pc_analyzer/frame_parser/parser.go`

**关键伪代码逻辑 (在 `parsePacketLayers` 函数内):**
*   **应用偏移 (针对特定管理帧):**
    *   `MgmtMeasurementPilot`: `TODO: 根据802.11标准研究固定头部长度并应用偏移。`
    *   `MgmtActionNoAck`: `TODO: 根据802.11标准研究固定头部长度并应用偏移。`
    *   其他管理帧（如 `MgmtBeacon`, `MgmtProbeResp`）已应用或将检查是否需要偏移。
*   **健壮的 IE 解析循环:**
    ```pseudocode
    LOOP WHILE payloadIndex < LENGTH(effectiveIEPayload):
      IF payloadIndex + 2 > LENGTH(effectiveIEPayload) THEN // 不足以读取ID和Length
        LOG_WARN_IE_ITERATION_INSUFFICIENT_HEADER(...)
        BREAK LOOP 
      ENDIF

      ieID = effectiveIEPayload[payloadIndex]
      ieLength = effectiveIEPayload[payloadIndex+1]
      payloadIndex = payloadIndex + 2 

      availableDataAfterIDLen = LENGTH(effectiveIEPayload) - payloadIndex
      IF ieLength > availableDataAfterIDLen THEN // 声明的长度超过实际剩余数据
        LOG_WARN_IE_ITERATION_INVALID_LENGTH(...)
        BREAK LOOP 
      ENDIF

      // ... (处理有效IE)
      payloadIndex = payloadIndex + ieLength 
    ENDLOOP
    ```

**日志增强建议 (摘要):**
*   `DEBUG_MGMT_PAYLOAD_OFFSET`: 记录原始payload长度、应用的偏移量、有效IE payload长度。
*   `WARN_IE_ITERATION_INSUFFICIENT_HEADER`: 记录因剩余payload不足以读取IE头部而中断。
*   `WARN_IE_ITERATION_INVALID_LENGTH`: 记录因IE声明长度超过可用数据而中断。
*   `DEBUG_IE_ITERATION`: 记录正在处理的每个IE的ID、名称和长度。

**后续步骤:**
1.  根据 802.11 标准，研究 `MgmtMeasurementPilot` 和 `MgmtActionNoAck` (以及其他可能相关的管理帧类型，如 `MgmtReassociationReq`) 的帧结构，确定它们在信息元素 (IEs) 字段开始前是否存在固定长度的字段。
2.  如果存在固定字段，在 `pc_analyzer/frame_parser/parser.go` 的 `parsePacketLayers` 函数中为这些帧类型实现正确的 payload 偏移。
3.  实施伪代码中描述的 IE 解析循环的健壮性改进。
4.  添加建议的增强日志。
5.  测试修复后的代码。

---
## Unit Testing for `parsePacketLayers` (2025-05-07)

*   **Objective:** To ensure the robustness and correctness of the `parsePacketLayers` function in `pc_analyzer/frame_parser/parser.go` after recent enhancements for IE parsing and management frame payload offsets.
*   **Test File:** `pc_analyzer/frame_parser/parser_test.go` was created.
*   **Covered Scenarios:**
    *   **Correct Payload Offset Application:**
        *   `MgmtMeasurementPilot` (3-byte fixed header)
        *   `MgmtAction` / `MgmtActionNoAck` (2-byte fixed header)
        *   `MgmtReassociationReq` (4-byte fixed header)
        *   `MgmtBeacon` / `MgmtProbeResp` (12-byte fixed header, tested implicitly via other IE tests using Beacon frames)
    *   **Robust IE Parsing Loop:**
        *   Frame payload too short for its declared fixed header (IE parsing skipped).
        *   Incomplete IE header (e.g., only IE ID present, length byte missing).
        *   Declared IE length exceeds actual available data in the payload.
    *   **SSID and Key Information Extraction:**
        *   Successful parsing of SSID, FrameType, SA, DA, BSSID.
        *   Frames containing a valid SSID IE.
        *   Frames containing a hidden SSID (SSID IE with length 0, parsed as `"<Hidden SSID>"`).
        *   Frames not containing any SSID IE (SSID remains empty).
        *   Frames containing multiple IEs, including SSID, Rates, and DSSet, ensuring all are parsed correctly.
        *   Correct channel determination (prioritizing RadioTap, then DSSet IE if RadioTap channel is 0/absent).
*   **Outcome:** All implemented unit tests passed successfully, verifying the intended improvements and robustness of the `parsePacketLayers` function. This significantly increases confidence in the frame parsing module.
*   **Dependencies:** Added `github.com/stretchr/testify/assert` to `pc_analyzer/go.mod` for assertions.
---
## gopacket Parsing Robustness Enhancements (2025-05-08)

*   **Issue:** Persistent `gopacket` parsing errors (e.g., `Dot11 length X too short`, `ERROR_NO_DOT11_LAYER`, `vendor extension size < 3`) were preventing the extraction of crucial data for metric calculation, leading to "N/A" values in the frontend.
*   **Strategy & Changes in [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0):**
    1.  **Stricter Error Handling for `packet.ErrorLayer()`:**
        *   In `parsePacketLayers`, if `gopacket.NewPacket` results in `packet.ErrorLayer()` not being `nil`, the function now immediately returns an error. This prevents attempts to process packets that `gopacket` itself has identified as fundamentally flawed.
    2.  **Mandatory Dot11 Layer:**
        *   If the `Dot11` layer cannot be decoded from the packet (i.e., `packet.Layer(layers.LayerTypeDot11)` is `nil`), `parsePacketLayers` now returns an error, regardless of whether a Radiotap layer was present. This ensures that only packets with a successfully parsed 802.11 MAC layer proceed to detailed IE (Information Element) parsing and subsequent processing.
    3.  **Adoption of `gopacket.Lazy` Decoding:**
        *   The call to `gopacket.NewPacket` in `parsePacketLayers` was changed from using `gopacket.Default` to `gopacket.Lazy`. This option decodes layers on-demand, which can improve performance and potentially offer more resilience against certain types of packet corruption by not attempting to decode all layers upfront if some are malformed. The critical layers (Radiotap, Dot11, and subsequently IP/TCP/UDP for payload length) are still explicitly accessed, triggering their decoding.
*   **Rationale:** These changes aim to make the packet parser more robust by:
    *   Quickly discarding packets that `gopacket` flags with errors at the initial decoding stage.
    *   Ensuring that a valid Dot11 layer, which is essential for most of the analysis, is present.
    *   Leveraging `gopacket.Lazy` to potentially bypass issues in non-critical or malformed layers that might otherwise halt decoding with `gopacket.Default`.
*   **Expected Outcome:** A reduction in unhandled parsing errors, leading to more reliable data extraction for metric calculations and fewer "N/A" displays on the frontend. Packets that are truly unparseable at the Dot11 level will be cleanly skipped.
## `tshark` 解析引擎替换 `gopacket` (2025-05-10)

**目标:** 将 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0) 中基于 `gopacket` 的 802.11 帧解析逻辑，替换为通过执行 `tshark -T fields` 命令并解析其 CSV 输出来获取帧信息。

**动机:**
*   `gopacket` 在处理某些损坏或非标准帧时可能遇到困难。
*   `tshark` 拥有更成熟和鲁棒的解析引擎。
*   `tshark -T fields` 可以精确指定所需字段，减少不必要的数据处理。

**核心架构组件:**

1.  **`TSharkExecutor` 模块/结构体:**
    *   **职责:** 管理 `tshark` 子进程的启动、停止和输出流的读取。
    *   **主要方法:**
        *   `Start(pcapFilePath string, tsharkPath string, fields []string) (io.ReadCloser, error)`: 启动 `tshark` 进程，配置必要的参数（如 `-r <pcap_file_path>`, `-T fields`, `-E header=y`, `-E separator=,`, `-E quote=d`, `-E occurrence=a` 以及所有必要的 `-e <field>` 参数），并返回其标准输出的 `io.ReadCloser`。
        *   `Stop()`: 终止 `tshark` 进程。
    *   **错误处理:** 监控 `tshark` 的 `stderr`，处理启动失败、运行时错误。

2.  **`CSVParser` 模块/结构体:**
    *   **职责:** 解析从 `TSharkExecutor` 获取的 `tshark` 输出 CSV 流。
    *   **主要方法:**
        *   `NewCSVParser(reader io.Reader) (*CSVParser, error)`: 初始化解析器。读取并解析 CSV 头部行，创建一个从字段名到其在 CSV 行中索引的映射。
        *   `ReadFrame() (map[string]string, error)`: 读取并解析 CSV 数据的一行，将其转换为一个从字段名到其字符串值的映射。处理行尾 (EOF) 和 CSV 解析错误。
    *   **依赖:** Go 标准库 `encoding/csv`。

3.  **`FrameProcessor` 模块/结构体:**
    *   **职责:** 将从 `CSVParser` 获取的原始字段名到值的映射，转换为结构化的 `ParsedFrameInfo` 对象。
    *   **主要方法:**
        *   `ProcessRow(row map[string]string) (*ParsedFrameInfo, error)`:
            *   使用辅助函数 (例如 `getString(row, fieldName)`, `getInt(row, fieldName)`, `getMac(row, fieldName)`, `getFloatEpochTime(row, fieldName)`) 从 `row` 中安全地提取和转换每个必要的字段值。这些辅助函数应能处理字段缺失和类型转换错误。
            *   根据需要解析 `wlan.fc.type_subtype` (十六进制) 为整数的 `FrameType` 和 `FrameSubType`。
            *   调用 `PhyRateCalculator` 来计算 PHY 速率，使用从 `row` 中提取的相关 Radiotap 字段。
            *   计算帧的物理传输时间（用于占空比）。
            *   填充 `ParsedFrameInfo` 结构体的所有相关字段。
    *   **错误处理:** 记录字段缺失、类型转换失败等问题。

4.  **`PhyRateCalculator` 类/结构体:**
    *   **职责:** (复用或调整现有逻辑) 根据从 `tshark` 输出中提取的 Radiotap MCS/VHT/HE 字段计算物理层速率。
    *   **主要方法:** `CalculatePhyRate(fields map[string]string) float64`。
    *   **逻辑:** 优先顺序为 HE -> VHT -> HT -> Legacy。如果 `tshark` 直接提供了 `radiotap.datarate` 并且其他更精确的 MCS 字段不可用，则可将其用作回退。

**主处理流程 (重写 `ProcessPcapStream` 或创建新函数在 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0)):**

1.  **初始化:**
    *   创建 `TSharkExecutor` 实例。
    *   从配置中获取 `tshark` 可执行文件路径和 pcap 文件路径。
    *   定义需要从 `tshark` 提取的完整字段列表 (基于规范)。
2.  **启动 `tshark`:**
    *   调用 `TSharkExecutor.Start(pcapFilePath, tsharkPath, fields)` 启动 `tshark` 进程，获取其标准输出流 (`stdoutReader`)。
    *   如果启动失败，记录错误并退出。
3.  **设置 CSV 解析:**
    *   创建 `CSVParser` 实例，传入 `stdoutReader`。
    *   如果头部解析失败，记录错误并退出。
4.  **帧处理循环:**
    *   创建 `FrameProcessor` 实例。
    *   循环:
        *   调用 `CSVParser.ReadFrame()` 获取下一行解析为 `map[string]string`。
        *   如果遇到 EOF 或不可恢复的 CSV 错误，则退出循环。
        *   调用 `FrameProcessor.ProcessRow(row)` 将原始数据转换为 `ParsedFrameInfo`。
        *   如果 `ProcessRow` 成功:
            *   将生成的 `ParsedFrameInfo` 对象传递给 `PacketInfoHandler` 回调 (供 `StateManager` 使用)。
        *   如果 `ProcessRow` 失败:
            *   记录详细错误信息（包括原始 CSV 行、错误详情、时间戳等）。
            *   继续处理下一行。
5.  **清理:**
    *   当循环结束或发生致命错误时，调用 `TSharkExecutor.Stop()` 来确保 `tshark` 进程被正确终止。
    *   关闭相关流。

**数据结构 (`ParsedFrameInfo`):**

*   现有 `ParsedFrameInfo` 结构体将保持，并根据需要进行最小调整，以确保所有从 `tshark` 解析的字段都能被正确存储或用于计算派生值。
*   例如，`WlanFcType` 和 `WlanFcSubtype` 将从解析 `wlan.fc.type_subtype` (十六进制字符串) 中获得。
*   确保有字段来存储计算出的 PHY 速率和帧通话时间。

**日志记录要求 (基于规范):**

*   **启动/关闭:** `tshark` 进程的启动命令、成功/失败状态、正常/异常退出。
*   **行解析失败:** 记录原始 CSV 行、错误信息、问题字段（如果可识别）、帧时间戳（如果可解析）。
*   **字段缺失:** 记录关键字段在 CSV 行中缺失的警告。
*   **无效值:** 记录字段值无法按预期类型解析的错误。
*   **摘要统计:** 定期记录已处理的总帧数、成功解析的帧数、失败解析的帧数。

**与现有代码的集成点:**

*   **[`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0):**
    *   `ProcessPcapStream` 函数将被重写以实现上述新的基于 `tshark` 的主处理流程。它将不再使用 `gopacket/pcapgo`。
    *   现有的 `parsePacketLayers` 函数将被 `FrameProcessor.ProcessRow` 的逻辑所取代。
    *   `getPHYRateMbps` (或其核心逻辑) 将被封装在 `PhyRateCalculator` 中。
    *   `CalculateFrameAirtime` 的逻辑将保留，但其输入参数（PHY 速率、帧长度）将从 `tshark` 解析出的数据中获取。
*   **[`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go:0):**
    *   `PacketInfoHandler` 回调接口 (`func(info *ParsedFrameInfo)`) 保持不变。`StateManager` 将继续通过此回调接收 `ParsedFrameInfo` 对象。
*   **配置 ([`desktop_app/WifiPcapAnalyzer/config/config.go`](desktop_app/WifiPcapAnalyzer/config/config.go:0), [`desktop_app/WifiPcapAnalyzer/config/config.json`](desktop_app/WifiPcapAnalyzer/config/config.json:0)):**
    *   应添加一个配置项来指定 `tshark` 可执行文件的路径，例如 `TsharkPath`。

**潜在风险和缓解措施 (基于规范):**

*   **`tshark` 性能:** 通过精确指定字段列表来最小化 `tshark` 的处理负担。流式处理 CSV 输出避免一次性加载所有数据。
*   **`tshark` 版本兼容性/字段名变更:** 在文档中记录所依赖的 `tshark` 版本。考虑使字段名列表可配置，以适应未来 `tshark` 的变化。
*   **CSV 解析复杂性:** 使用 Go 标准库 `encoding/csv`，它是一个成熟且经过良好测试的库。对多值字段（通过 `occurrence=a` 获取）的解析需要在 `FrameProcessor` 中特别处理。
*   **`tshark` 错误处理:** `TSharkExecutor` 必须监控 `tshark` 的 `stderr` 输出，并实现超时和可能的重试逻辑（如果适用）。

**Mermaid 图 (高级流程):**
```mermaid
graph TD
    A[Pcap File] --> B{TSharkExecutor};
    B -- Stdout (CSV Stream) --> C{CSVParser};
    C -- Parsed CSV Row (map[string]string) --> D{FrameProcessor};
    D -- ParsedFrameInfo --> E[PacketInfoHandler (StateManager)];
    F[PhyRateCalculator] <--> D;
    B -- Stderr --> G[Logging];
    C -- Errors --> G;
    D -- Errors --> G;

    subgraph Frame Parser Module (`parser.go`)
        B
        C
        D
        F
    end
```

**时间戳:** 2025-05-10

## Logging Library Migration

*   **Objective:** Migrate the Go backend to use `zerolog` for improved logging efficiency and maintainability.
*   **Changes:**
    *   Removed temporary `fmt.Println()` and `log.Println()` statements.
    *   Added `github.com/rs/zerolog` dependency in `go.mod`.
    *   Verified the existence of required logging configuration structure in `config.go` and `config.json`.
    *   Verified the existence of `zerolog` initialization logic in `main.go` and `logger.go`.
    *   Replaced existing `log.Printf` with `zerolog` usage (e.g., `logger.Log.Info().Msgf(...)`) throughout the Go backend code.
    *   Resolved compilation errors and import cycle issues when importing the `logger` package.
---