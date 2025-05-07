# Decision Log

This file records architectural and implementation decisions using a list format.

2025-05-06 23:48:00 - Initial population of architectural decisions.
*
      
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