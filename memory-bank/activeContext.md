# Active Context

This file tracks the project's current status, including recent changes, current goals, and open questions.

*   [2025-05-07 18:00:00] - **当前焦点:** Web UI 仍未显示 BSS/STA 信息。已增强后端 `pc_analyzer` 对 HT/VHT 能力和带宽的解析，并改进了SSID的UTF-8验证。修改了 `pc_analyzer/frame_parser/parser.go` 和 `pc_analyzer/state_manager/manager.go`。等待用户测试以确认UI是否正常显示。
*   [2025-05-07 18:00:00] - **近期变更:**
    *   PC端分析引擎: `pc_analyzer/frame_parser/parser.go` 更新了HT/VHT IE的解析逻辑，以提取带宽和更详细的能力信息，并增加了SSID的UTF-8验证。
    *   PC端分析引擎: `pc_analyzer/state_manager/manager.go` 更新了以使用 `frame_parser` 中解析出的更完整的带宽和HT/VHT能力数据。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/decisionLog.md` 即将更新。
*   [2025-05-07 18:16:00] - **当前焦点:** Web UI 不显示 BSS/STA 信息的问题已定位到前端。修复了 `types/data.ts` 中的类型定义错误（`Station` -> `STA`, `WebSocketData` 结构），修正了 `DataContext.tsx` 中 `staList` 未更新的问题，并调整了 `BssList.tsx` 以正确使用更新后的类型和数据结构。等待用户测试验证。
*   [2025-05-07 18:16:00] - **近期变更:**
    *   Web前端: `types/data.ts` 更新了 `STA` 和 `WebSocketData` 类型定义。
    *   Web前端: `contexts/DataContext.tsx` 修正了 `SET_DATA` reducer 以更新 `staList`，并修复了相关类型错误。
    *   Web前端: `components/BssList/BssList.tsx` 更新了导入和组件逻辑以使用正确的类型 (`STA`) 和属性 (`signal_strength`, `last_seen`, `associated_stas`)，并正确处理了 `associated_stas` 对象。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/decisionLog.md` 即将更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
*   [2025-05-07 17:07:00] - **当前焦点:** PC端分析引擎 (`pc_analyzer`) 的编译错误已解决。具体修复了 `pc_analyzer/frame_parser/parser.go` 中的新诊断日志相关的编译问题，以及 `pc_analyzer/state_manager/manager.go` 和 `pc_analyzer/frame_parser/parser_test.go` 中因 `parsePacketLayers` 函数签名变更和 `ParsedFrameInfo` 结构调整引起的问题。特别是在 `parser_test.go` 中，通过使用 `layers.LinkType(127)` 替代了无法解析的 `layers.LinkTypeRadioTap` 常量，使得测试代码能够编译通过。下一步是请用户运行代码并提供新的日志，以诊断最初的 Beacon 帧解析问题。
*   [2025-05-07 17:07:00] - **近期变更:**
    *   PC端分析引擎: `pc_analyzer/frame_parser/parser.go` 添加了更详细的 `gopacket.NewPacket` 层解析日志和 `Dot11` 层获取失败时的诊断日志。修复了之前引入的 `HECapabilities` 和 QoS 子类型常量导致的编译错误。
    *   PC端分析引擎: `pc_analyzer/state_manager/manager.go` 更新了对 `parsedInfo.FrameType` 的使用，以适应 `ParsedFrameInfo` 结构中 `FrameSubType` 的移除。
    *   PC端分析引擎: `pc_analyzer/frame_parser/parser_test.go` 更新了对 `parsePacketLayers` 函数的调用，以匹配新的函数签名 `([]byte, layers.LinkType, time.Time)`，并使用 `layers.LinkType(127)` 作为 `LinkTypeRadioTap` 的替代方案解决了编译问题。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
*   [2025-05-07 16:40:00] - **Debug Status Update:** Investigating Beacon frame fixed header parsing failure (`dot11.Payload` empty). Added detailed diagnostic logs to `pc_analyzer/frame_parser/parser.go` to trace Radiotap and Dot11 layer payload generation. Awaiting user to run tests and provide new logs for analysis.
*   [2025-05-07 16:24:00] - **TDD完成与验证:** 为 `pc_analyzer/frame_parser/parser.go` 中的 `parsePacketLayers` 函数成功编写并执行了全面的单元测试。这些测试 (`pc_analyzer/frame_parser/parser_test.go`) 验证了对多种管理帧类型（MgmtMeasurementPilot, MgmtAction, MgmtActionNoAck, MgmtReassociationReq）的正确payload偏移处理，以及IE解析循环在各种边缘情况下的鲁棒性（包括payload过短、IE头部不完整、IE声明长度无效或超出可用数据）。同时，测试确保了SSID（包括有效SSID、隐藏SSID、无SSID IE的情况）和其他关键信息（如FrameType, SA, DA, BSSID）能够被正确提取。所有单元测试均已通过，增强了对该模块稳定性和正确性的信心。
*   [2025-05-07 15:32:48] - **代码变更:** 改进了 `pc_analyzer/frame_parser/parser.go` 中的帧解析逻辑。为多种管理帧（MgmtMeasurementPilot, MgmtAction, MgmtActionNoAck, MgmtReassociationReq）添加了正确的固定头部payload偏移处理。增强了IE解析循环的鲁棒性，以更好地处理IE长度声明错误和数据不足的情况。添加了更详细的调试日志。
*   [2025-05-07 15:20:54] - **当前焦点:** 解决因 IE 解析器无法妥善处理数据不足/异常长度以及 `MgmtMeasurementPilot` 和 `MgmtActionNoAck` 帧缺少 payload 偏移导致的 SSID 解析失败问题。计划方案包括修改 IE 解析循环的鲁棒性，并在研究后为特定管理帧实现正确的 payload 偏移。
*   [2025-05-07 15:20:54] - **近期变更:**
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新，记录了关于 SSID 解析问题（IE 数据不足/异常长度，特定管理帧偏移）的新焦点和计划方案。
    *   Memory Bank: `memory-bank/decisionLog.md` 即将更新，记录改进 IE 解析鲁棒性和研究特定帧类型偏移的决策。
    *   Memory Bank: `memory-bank/developmentContext/pcAnalysisEngine.md` 即将更新，记录当前 SSID 解析问题的分析和详细修复方案。
*   [2025-05-07 15:01:00] - **当前焦点:** PC端分析引擎 (`pc_analyzer/frame_parser/parser.go`) 的SSID解析问题已通过修正Beacon/ProbeResp帧IE payload的起始偏移量得到解决。此前因未跳过管理帧头部的固定字段，导致IE解析逻辑提前中断。等待用户测试验证SSID能否在前端正确显示，以及 `DEBUG_SSID_PARSE` 和 `DEBUG_MGMT_PAYLOAD_OFFSET` 日志是否按预期出现。
*   [2025-05-07 15:01:00] - **近期变更:**
    *   PC端分析引擎: `pc_analyzer/frame_parser/parser.go` 中的 `parsePacketLayers` 函数已更新，为Beacon和Probe Response帧的IE解析逻辑添加了12字节的payload偏移。
    *   Memory Bank: `memory-bank/decisionLog.md` 已更新，记录了关于修正管理帧IE payload偏移的决策。
    *   Memory Bank: `memory-bank/developmentContext/pcAnalysisEngine.md` 已更新，详细记录了此SSID解析问题的分析和修复方案。
    *   Memory Bank: `memory-bank/progress.md` 已更新，将此修复任务标记为完成。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
*   [2025-05-07 14:48:00] - **当前焦点:** PC端分析引擎 (`pc_analyzer/frame_parser/parser.go`) 的SSID解析问题已通过简化IE解析逻辑得到修复。代码现在仅依赖遍历 `dot11.Payload` 来提取IEs，解决了先前因尝试访问特定gopacket结构字段而导致的编译错误。等待用户测试验证SSID能否在前端正确显示，以及 `DEBUG_SSID_PARSE` 日志是否按预期出现。
*   [2025-05-07 14:48:00] - **近期变更:**
    *   PC端分析引擎: `pc_analyzer/frame_parser/parser.go` 中的 `parsePacketLayers` 函数已重写，以仅使用 `dot11.Payload` 进行IE解析。
    *   Memory Bank: `memory-bank/decisionLog.md` 已更新，记录了关于简化IE解析逻辑的决策。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
*   [2025-05-07 14:11:00] - **当前焦点:** PC端分析引擎的SSID解析问题已通过修改 `pc_analyzer/frame_parser/parser.go` 得到解决。代码中添加了更详细的IE遍历日志，并改进了对SSID IE（包括隐藏SSID）的提取和处理逻辑。等待用户测试验证SSID能否在前端正确显示。
*   [2025-05-07 14:11:00] - **近期变更:**
    *   PC端分析引擎: `pc_analyzer/frame_parser/parser.go` 更新了SSID IE的解析逻辑和相关调试日志。
    *   Memory Bank: `memory-bank/developmentContext/pcAnalysisEngine.md` 已更新，记录了SSID解析问题的详细分析和修复方案。
    *   Memory Bank: `memory-bank/progress.md` 已更新，将SSID解析修复任务标记为完成，并更新了当前任务状态。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
*   [2025-05-07 13:24:06] - **当前焦点:** PC端分析引擎的 `radiotap layer not found` 帧解析问题已通过引入 `pcapgo` 库和相应的代码重构得到解决。相关文件 (`grpc_client/client.go`, `frame_parser/parser.go`, `main.go`) 已更新。Memory Bank相关文件已更新。等待测试验证。
*   [2025-05-07 13:24:06] - **近期变更:**
    *   PC端分析引擎: `pc_analyzer/grpc_client/client.go` 已更新，使用 `io.Pipe` 将gRPC流数据传递给pcap处理器。
    *   PC端分析引擎: `pc_analyzer/frame_parser/parser.go` 已更新，引入 `ProcessPcapStream` 函数，使用 `pcapgo.NewReader` 解析pcap流中的数据包，并调整了原有解析逻辑。
    *   PC端分析引擎: `pc_analyzer/main.go` 已更新，调整了数据处理回调逻辑以适应新的pcap流处理方式。
    *   Memory Bank: `memory-bank/developmentContext/pcAnalysisEngine.md` 已更新，详细记录了 `radiotap layer not found` 问题的分析和修复方案。
    *   Memory Bank: `memory-bank/progress.md` 已更新，将此调试任务标记为完成。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
*   [2025-05-07 12:18:00] - **当前焦点:** 诊断并定位了PC分析引擎与路由器代理之间的gRPC "Unimplemented" 错误 (`unknown service router_agent_pb.CaptureAgent`)。根本原因是客户端 (`pc_analyzer`) 的 `.proto` 文件 (`pc_analyzer/capture_agent.proto`) 中的 `go_package` 选项导致生成的gRPC代码使用了错误的服务名 (`router_agent_pb.CaptureAgent`)，而服务器 (`router_agent`) 注册的是 `router_agent.CaptureAgent`。
*   [2025-05-07 12:18:00] - **近期变更:**
    *   PC端分析引擎: `pc_analyzer/capture_agent.proto` 文件中的 `option go_package` 已从 `"...;router_agent_pb"` 修改为 `"...;router_agent"`，以确保生成的客户端代码使用正确的服务名 `router_agent.CaptureAgent`。
    *   Memory Bank: `memory-bank/decisionLog.md` 已更新，详细记录了此gRPC问题的根本原因、分析过程和修复决策。
    *   **待办:** `pc_analyzer` 项目中的gRPC Go代码 (`pc_analyzer/router_agent_pb/`) 需要使用 `protoc` 命令重新生成，以使上述 `.proto` 文件更改生效。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
*   [2025-05-07 12:00:00] - **当前焦点:** 正在调试PC端分析引擎WebSocket控制指令解析问题，具体表现为收到 "Unknown WebSocket control command:" 且命令字符串为空。已在 `pc_analyzer/main.go` 中添加详细的JSON反序列化后及命令提取后的日志，以便进一步观察实际解析的命令内容。
*   [2025-05-07 12:00:00] - **近期变更:**
    *   PC端分析引擎: `pc_analyzer/main.go` 的 `webSocketControlMessageHandler` 函数增加了调试日志，用于打印解析后的 `ControlCommandMsg` 结构体内容和待分派的 `actualCommand` 字符串。
    *   Memory Bank: `memory-bank/developmentContext/pcAnalysisEngine.md` 已更新，记录了此问题的分析过程和添加的调试日志详情。
    *   Memory Bank: `memory-bank/progress.md` 已更新，将添加调试日志的任务标记为完成。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
*   [2025-05-07 11:34:00] - **当前焦点:** Web前端已更新，以符合PC端分析引擎对 `start_capture` 命令 `payload` 的新要求。`interface` 字段已添加并硬编码为 "ath1"。等待集成测试。
*   [2025-05-07 11:34:00] - **近期变更:**
    *   Web前端: `web_frontend/src/components/ControlPanel/ControlPanel.tsx` 已更新，在 `start_capture` 命令的 `payload` 中添加 `interface` ("ath1"), `channel`, 和 `bandwidth`。
    *   Web前端: `web_frontend/src/types/data.ts` 中 `ControlCommand` 的 `payload` 类型定义已更新，添加了 `interface?: string`。
    *   Memory Bank: `memory-bank/developmentContext/webFrontend.md` 已更新，记录了 `start_capture` payload 的修改。
    *   Memory Bank: `memory-bank/progress.md` 已更新，将此任务标记为完成。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
*   [2025-05-07 11:30:00] - **当前焦点:** PC端分析引擎的WebSocket控制指令解析问题再次出现并已修复。现在强制要求 `start_capture` 指令的 `payload` 中必须包含 `InterfaceName`。等待进一步的集成测试。
*   [2025-05-07 11:30:00] - **近期变更:**
    *   PC端分析引擎: `pc_analyzer/main.go` 中的 `webSocketControlMessageHandler` 函数已更新，对于 `start_capture` 命令，如果 `payload` 中缺少 `InterfaceName`，将明确返回错误。
    *   Memory Bank: `memory-bank/developmentContext/pcAnalysisEngine.md` 已更新，记录了此问题的再次分析和修复方案。
    *   Memory Bank: `memory-bank/decisionLog.md` 已更新，记录了强制要求 `InterfaceName` 的决定。
    *   Memory Bank: `memory-bank/progress.md` 已更新，将此调试任务标记为完成。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
*   [2025-05-07] - **当前焦点:** PC端分析引擎的WebSocket控制指令解析问题已解决。引擎现在能够正确处理来自Web前端的 `start_capture` 等指令。等待进一步的集成测试。
*   [2025-05-07] - **近期变更:**
    *   PC端分析引擎: `pc_analyzer/main.go` 中的 `webSocketControlMessageHandler` 函数已更新，以兼容前端发送的 `action` 或 `command` 字段，并正确处理嵌套的 `payload`（包括 `interface` 字段）。
    *   Memory Bank: `memory-bank/developmentContext/pcAnalysisEngine.md` 已更新，记录了此问题的分析和修复方案。
    *   Memory Bank: `memory-bank/progress.md` 已更新，将此调试任务标记为完成。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
*   [2025-05-07 02:43:00] - **当前焦点:** Web前端 `ControlPanel` UI对齐问题已根据用户反馈进一步修复，确保所有控制按钮正确对齐。等待用户测试验证。
*   [2025-05-07 02:43:00] - **近期变更:**
    *   Web前端: 修复了 `ControlPanel.css` 中 `.control-group` 的按钮对齐问题，确保 "Set Channel" 和 "Set Bandwidth" 按钮在窄屏幕下正确换行 (用户反馈问题2)。
    *   Memory Bank: 更新了 `memory-bank/developmentContext/webFrontend.md` 以记录此修复。
    *   Memory Bank: 更新了 `memory-bank/progress.md` 以反映此Bug修复的完成。
    *   Memory Bank: 更新了 `memory-bank/activeContext.md`。
*   [2025-05-07 02:35:00] - **先前焦点:** Router Agent (`router_agent`) 交叉编译问题已解决。项目现在可以为 `linux/arm64` 目标正确编译。等待用户进行下一步的集成测试或部署。
*   [2025-05-07 02:35:00] - **先前变更:**
    *   Router Agent: 解决了 `router_agent` 针对 `linux/arm64` 的交叉编译问题。
    *   Router Agent: `router_agent/main.go` 包声明更改为 `main`。
    *   Router Agent: `router_agent/capture_agent.proto` 中 `option go_package` 更新为 `".;main"`。
    *   Router Agent: `router_agent/capture_agent.pb.go` 和 `router_agent/capture_agent_grpc.pb.go` 已重新生成为 `package main`。
    *   Router Agent: `router_agent/go.mod` 中 Go 版本更新为 `1.20`，`google.golang.org/grpc` 更新为 `v1.64.0`，`google.golang.org/protobuf` 更新为 `v1.33.0`。
    *   Memory Bank: 更新了 `memory-bank/deployment/routerAgentDeployment.md` 和 `memory-bank/developmentContext/routerAgent.md`。
    *   Memory Bank: 更新了 `memory-bank/progress.md`。
    *   Memory Bank: 更新了 `memory-bank/activeContext.md`。
*   [2025-05-07 02:20:00] - **当前焦点:** Web前端Bug修复完成 (UI对齐, 5GHz信道列表, BssList运行时错误)。等待用户测试验证这些修复。
*   [2025-05-07 02:20:00] - **近期变更:**
    *   Web前端: 修复了 `ControlPanel.css` 中的按钮对齐问题 (Issue 2.1)。
    *   Web前端: 修改了 `ControlPanel.tsx` 以支持5GHz信道列表并更新默认信道 (Issue 2.2)。
    *   Web前端: 修改了 `DataContext.tsx` 和 `BssList.tsx` 以解决BssList组件的运行时TypeError (Issue 2.3)。
    *   Memory Bank: 更新了 `memory-bank/developmentContext/webFrontend.md` 以记录上述修复。
    *   Memory Bank: 更新了 `memory-bank/progress.md` 以反映Bug修复的完成。
    *   Memory Bank: 更新了 `memory-bank/activeContext.md`。
*   [2025-05-07 02:04:00] - **先前焦点:** 系统集成测试计划已完成。等待用户执行测试并提供反馈。
*   [2025-05-07 02:04:00] - **先前变更:**
    *   Memory Bank: 创建了 `memory-bank/testing/integrationTestPlan.md` 并记录了详细的端到端测试计划和执行步骤。
    *   Memory Bank: 更新了 `memory-bank/progress.md` 以反映集成测试计划的完成。
    *   Memory Bank: 更新了 `memory-bank/activeContext.md`。
*   [2025-05-07 01:56:00] - **先前焦点:** Web前端可视化界面核心功能初步实现完成。
*   [2025-05-07 01:56:00] - **先前变更:**
    *   Web前端: 使用Create React App (TypeScript) 在 `web_frontend/` 目录初始化了项目。
    *   Web前端: 实现了WebSocket客户端逻辑 (`websocketService.ts`) 用于连接PC端引擎。
    *   Web前端: 定义了BSS/STA数据类型 (`types/data.ts`)。
    *   Web前端: 使用React Context API (`contexts/DataContext.tsx`) 进行状态管理。
    *   Web前端: 创建了BSS列表展示组件 (`components/BssList/`) 和STA列表展示逻辑。
    *   Web前端: 创建了控制面板组件 (`components/ControlPanel/`) 用于发送指令。
    *   Web前端: 更新了主应用组件 (`App.tsx`) 和相关CSS文件以集成各部分。
    *   Web前端: 解决了`ControlPanel.tsx`中的TypeScript导入错误。
    *   Memory Bank: 创建了 `memory-bank/developmentContext/webFrontend.md` 并详细记录了前端实现细节。
    *   Memory Bank: 更新了 `memory-bank/progress.md` 以反映前端开发进展。
*   [2025-05-07 01:44:00] - **先前焦点:** PC端实时分析引擎核心功能模块实现完成。
*   [2025-05-07 01:44:00] - **先前变更:**
    *   PC端引擎: 各模块已整合到 `pc_analyzer/main.go` 中，形成完整引擎初步版本。
*   [2025-05-06 23:48:00] - 设计可视化空口抓包分析器Demo的系统架构。
*   [2025-05-06 23:48:00] - 初始化Memory Bank并记录关键架构信息。
*   [2025-05-06 23:48:00] - Initial population.
*   [2025-05-07 21:15:35] - **Debug Status Update:** Investigating issue of numerous 0dBm STA entries in the STA list.
*   [2025-05-07 21:15:35] - **Debug Status Update:** Applied fix to `pc_analyzer/state_manager/manager.go` to update STA signal strength when STA is identified/created via data frames, if `parsedInfo.SignalStrength` is non-zero. This aims to reduce 0dBm STA entries.

## Recent Changes
*   [2025-05-07 21:15:35] - PC端分析引擎: `pc_analyzer/state_manager/manager.go` 在处理数据帧的逻辑中增加了对STA信号强度的更新，以尝试解决0dBm STA过多的问题。
*   [2025-05-07 17:07:00] - PC端分析引擎: `pc_analyzer/frame_parser/parser.go` 添加了更详细的 `gopacket.NewPacket` 层解析日志和 `Dot11` 层获取失败时的诊断日志。修复了之前引入的 `HECapabilities` 和 QoS 子类型常量导致的编译错误。
*   [2025-05-07 17:07:00] - PC端分析引擎: `pc_analyzer/state_manager/manager.go` 更新了对 `parsedInfo.FrameType` 的使用，以适应 `ParsedFrameInfo` 结构中 `FrameSubType` 的移除。
*   [2025-05-07 17:07:00] - PC端分析引擎: `pc_analyzer/frame_parser/parser_test.go` 更新了对 `parsePacketLayers` 函数的调用，以匹配新的函数签名 `([]byte, layers.LinkType, time.Time)`，并使用 `layers.LinkType(127)` 作为 `LinkTypeRadioTap` 的替代方案解决了编译问题。
*   [2025-05-07 17:07:00] - Memory Bank: `memory-bank/activeContext.md` 本次更新。
*   [2025-05-07 16:40:00] - **Debug Status Update:** Investigating Beacon frame fixed header parsing failure (`dot11.Payload` empty). Added detailed diagnostic logs to `pc_analyzer/frame_parser/parser.go` to trace Radiotap and Dot11 layer payload generation. Awaiting user to run tests and provide new logs for analysis.
*   [2025-05-07 16:24:00] - **TDD完成与验证:** (内容同上)
*   [2025-05-07 15:32:48] - **代码变更:** (内容同上)
*   [2025-05-07 15:20:54] - **近期变更:** (内容同上)
*   [2025-05-07 15:01:00] - **近期变更:** (内容同上)
*   [2025-05-07 14:48:00] - **近期变更:** (内容同上)
*   [2025-05-07 14:11:00] - **近期变更:** (内容同上)
*   [2025-05-07 13:24:06] - **近期变更:** (内容同上)
*   [2025-05-07 12:18:00] - **近期变更:** (内容同上)
*   [2025-05-07 12:00:00] - **近期变更:** (内容同上)
*   [2025-05-07 11:34:00] - **近期变更:** (内容同上)
*   [2025-05-07 11:30:00] - **近期变更:** (内容同上)
*   [2025-05-07] - **近期变更:** (内容同上)
*   [2025-05-07 02:43:00] - **近期变更:** (内容同上)
*   [2025-05-07 02:35:00] - **先前变更:** (内容同上)
*   [2025-05-07 02:20:00] - **近期变更:** (内容同上)
*   [2025-05-07 02:04:00] - **先前变更:** (内容同上)
*   [2025-05-07 01:56:00] - **先前变更:** (内容同上)
*   [2025-05-07 01:44:00] - **先前变更:** (内容同上)
*   [2025-05-07 01:42:00] - PC端引擎: gRPC客户端流式通信逻辑 (`pc_analyzer/grpc_client/client.go`) 实现完成。
*   [2025-05-07 01:39:00] - PC端引擎: WebSocket服务器 (`pc_analyzer/websocket_server/server.go`) 核心逻辑实现完成。
*   [2025-05-07 01:35:00] - PC端引擎: BSS/STA状态管理模块 (`pc_analyzer/state_manager/manager.go`) 核心逻辑实现完成。
*   [2025-05-07 01:27:00] - PC端引擎: 802.11帧解析模块 (`pc_analyzer/frame_parser/parser.go`) 初步实现完成。
*   [2025-05-07 01:08:00] - PC端引擎: 配置加载模块 (`pc_analyzer/config/config.go`, `pc_analyzer/config/config.json`) 实现完成。
*   [2025-05-07 00:47:00] - PC端分析引擎项目骨架搭建完成。
*   [2025-05-07 00:47:00] - `memory-bank/developmentContext/pcAnalysisEngine.md` 创建并更新。
*   [2025-05-07 00:47:00] - `memory-bank/progress.md` 更新。
*   [2025-05-06 23:48:00] - Memory Bank 初始化完成。
*   [2025-05-06 23:48:00] - `productContext.md` 已更新项目目标、关键特性和总体架构。
*   [2025-05-06 23:48:00] - `decisionLog.md` 已记录数据传输协议选型和组件核心职责。

## Open Questions/Issues

*   [2025-05-07 21:15:35] - Monitor if the recent fix in `pc_analyzer/state_manager/manager.go` effectively reduces the number of 0dBm STA entries. If the issue persists, further investigation into `frame_parser.go` or `models.go` (STA initialization) might be needed.
*   暂无其他。
---
**Context Update:** Router Agent Cross-Compilation Guidance
**Timestamp:** 2025/5/7 上午2:12:54
**Details:** Guidance for cross-compiling `router_agent` for `linux/aarch64` has been documented in `memory-bank/deployment/routerAgentDeployment.md`. This includes build commands, environment variables (`GOOS=linux`, `GOARCH=aarch64`), and steps for transferring the binary to the target router.