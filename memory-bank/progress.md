# Progress

This file tracks the project's progress using a task list format.

2025-05-06 23:48:00 - Initial population.
*
*   [2025-05-08 16:46:00] - **TDD 任务调整与后端指标 "N/A" 问题诊断:**
    *   根据用户反馈，暂停原定的为吞吐量和信道占空比新计算逻辑编写单元测试的计划。
    *   当前TDD任务调整为：协助用户进行真实场景测试以诊断前端指标显示 "N/A" 的问题。
    *   已分析代码 (`parser.go`, `manager.go`, `models.go`)，确认了新的计算逻辑（吞吐量基于TCP/UDP载荷，BSS信道占空比基于MAC Duration/ID）。
    *   指出了STA信道占空比计算仍依赖旧的 `totalAirtime`，可能需要后续统一。
    *   向用户提供了详细的日志添加建议（在 `parser.go` 和 `manager.go` 中），以追踪关键变量和计算步骤，帮助定位 "N/A" 问题的原因。
    *   等待用户提供带有增强日志的程序输出来进行下一步分析。
    *   已更新 `memory-bank/activeContext.md` 以反映此调试焦点。

*   [2025-05-08 16:37:00] - **后端信道占空比计算重构:**
    *   修改了帧解析器 ([`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0))、数据模型 ([`desktop_app/WifiPcapAnalyzer/state_manager/models.go`](desktop_app/WifiPcapAnalyzer/state_manager/models.go:0)) 和状态管理器 ([`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go:0))，以使用 802.11 MAC 头部的 `Duration/ID` 字段来计算信道占空比。
    *   添加了对 PS-Poll 控制帧的特殊处理逻辑，以避免错误地将 AID 累加为 NAV 时间。
    *   移除了旧的基于 `CalculateFrameAirtime` 的占空比计算逻辑。
*   [2025-05-08 15:53:00] - **后端 "N/A" 问题调试与初步修复:**
    *   定位到新功能指标（信道占空比、吞吐量）在前端显示 "N/A" 的主要原因是后端 `BSSInfo` 和 `STAInfo` 的 `lastCalcTime` 未初始化，导致首次指标计算结果为0。
    *   已修复 [`desktop_app/WifiPcapAnalyzer/state_manager/models.go`](desktop_app/WifiPcapAnalyzer/state_manager/models.go:0) 中的 `NewBSSInfo` 和 `NewSTAInfo` 函数，正确初始化 `lastCalcTime`。
    *   评估了当前指标计算方法的复杂度。
## Completed Tasks
*   [2025-05-08 14:48:00] - **Wails前端新功能实现 (实时信道占空比与吞吐量分析):**
    *   TypeScript类型 ([`desktop_app/WifiPcapAnalyzer/frontend/src/types/data.ts`](desktop_app/WifiPcapAnalyzer/frontend/src/types/data.ts:0)): `BSS` 和 `STA` 接口已扩展以包含新性能指标。
    *   DataContext ([`desktop_app/WifiPcapAnalyzer/frontend/src/contexts/DataContext.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/contexts/DataContext.tsx:0)): 已更新以管理 `selectedPerformanceTarget`。
    *   BSS列表 ([`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx:0)): 已修改以显示关键指标并更新选择。
    *   STA列表 ([`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx:0)): 已修改以显示关键指标并更新选择。
    *   PerformanceDetailPanel ([`desktop_app/WifiPcapAnalyzer/frontend/src/components/PerformanceDetailPanel/PerformanceDetailPanel.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/PerformanceDetailPanel/PerformanceDetailPanel.tsx:0)): 新建组件以使用 `recharts` 显示详细性能图表。
    *   App布局 ([`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0), [`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0)): 已调整以集成新面板。
    *   依赖安装: 已安装 `recharts` 和 `@types/recharts`。
*   [2025-05-08 14:36:00] - **Wails后端新功能实现 (实时信道占空比与吞吐量):**
    *   扩展了 [`desktop_app/WifiPcapAnalyzer/state_manager/models.go`](desktop_app/WifiPcapAnalyzer/state_manager/models.go:0) 中的 `BSSInfo` 和 `STAInfo` 数据结构以包含新指标字段。
    *   在 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0) 中添加了 `CalculateFrameAirtime` 和 `getPHYRateMbps` 辅助函数（简化模型），并更新了 `ParsedFrameInfo`。
    *   在 [`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go:0) 中实现了指标累积 (`ProcessParsedFrame`) 和定期计算 (`PeriodicallyCalculateMetrics`) 逻辑。
    *   更新了 [`desktop_app/WifiPcapAnalyzer/app.go`](desktop_app/WifiPcapAnalyzer/app.go:0) 的 `startup` 方法以初始化 `StateManager` 并启动指标计算的 goroutine。
    *   确认了通过 `runtime.EventsEmit` 推送快照的逻辑无需修改即可包含新指标。
*   [2025-05-08 13:04:00] - **Web前端UI细节调整 (根据用户反馈):**
    *   解决了控制面板折叠后，BSS 和 STA 列表未填充释放空间的问题 ([`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0))。
    *   确保了 BSS 条目展开后，Security 信息单独一行显示 ([`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx:0), [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css:0))。
    *   调整了 BSS 和 STA 列表的宽度比例为约 2:3 ([`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0))。
    *   调整了控制面板展开时的宽度，以确保内容完全显示 ([`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0))。
*   [2025-05-08 12:26:00] - **Web前端布局与样式调整 (根据用户反馈):**
    *   调整了三列布局的列宽定义，为 BSS 和 STA 列表设置最小宽度，防止被过度挤压 ([`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0))。
    *   为应用主背景添加了浅灰色，以区分白色卡片 ([`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0))。
    *   增强了选中 BSS 卡片的高亮效果（添加背景色） ([`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css:0))。
*   [2025-05-08 11:51:00] - **Web前端 BSS/STA 列表调整 (根据用户反馈):**
    *   调整了 BSS 列表项样式：为选中项添加高亮，移除展开指示器，调整了摘要字段布局 ([`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx:0), [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css:0))。
    *   将 STA 列表的显示方式从表格改回为每个 STA 使用独立的卡片展示 ([`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx:0), [`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.module.css:0))。
*   [2025-05-08 11:27:00] - **Web前端布局调整 (根据用户反馈):**
    *   移除了应用顶部的黑色背景 header ([`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0), [`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0))。
    *   将主内容区调整为三列水平布局，分别为 `ControlPanel`、`BssList` 和 `StaList` ([`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0), [`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0))。
    *   实现了 `ControlPanel` 折叠后水平空间动态调整的逻辑。
*   [2025-05-08 11:18:00] - **Web前端 UI/UX Redesign (Phase 2 - Common Components & Refactoring):**
    *   创建了通用UI组件 (`Button`, `Input`, `Card`, `Table`, `Tabs`, `Icon`) 于 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/common/`](desktop_app/WifiPcapAnalyzer/frontend/src/components/common/)。
    *   更新了 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.tsx:0) 以使用新的 `Button` 和 `Input` 组件。
    *   更新了 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx:0) 以使用新的 `Card` 组件。
    *   更新了 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx:0) 以使用新的 `Table` 组件。
    *   删除了不再使用的旧 CSS 文件 (`ControlPanel.css`, `BssList.css`, `StaList.css`)。
    *   确认了 [`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0) 和 [`desktop_app/WifiPcapAnalyzer/frontend/src/index.css`](desktop_app/WifiPcapAnalyzer/frontend/src/index.css:0) 符合新的 UI/UX 规范。
    *   确保了字体 (SF Pro) 和 SVG 图标 (通过 `Icon.tsx`) 按照架构文档管理。
*   [2025-05-08 04:23:00] - **Web前端 UI/UX Redesign (Phase 1):**
    *   实施了全局样式 ([`desktop_app/WifiPcapAnalyzer/frontend/src/index.css`](desktop_app/WifiPcapAnalyzer/frontend/src/index.css:0)): 定义了 CSS 变量（配色、字体、圆角、阴影），并更新了基础 HTML 元素样式。
    *   更新了基础布局 ([`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0)): 修改了应用的主体布局，引入了网格布局和新的排版规则。
    *   重构和重新设计了核心组件样式:
        *   `ControlPanel`: [`desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.tsx:0) 和新的 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.module.css:0)。
        *   `BssList`: [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx:0) 和新的 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css:0)。
        *   `StaList`: [`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx:0) 和新的 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.module.css:0)。
    *   所有上述组件均已迁移到使用 CSS Modules。
*   [2025-05-07 21:15:35] - **PC端分析引擎:** 修复了 `pc_analyzer/state_manager/manager.go` 中数据帧处理逻辑，确保在通过数据帧识别/创建STA时，如果 `parsedInfo.SignalStrength` 非零，则更新其信号强度。旨在解决STA列表中出现大量0dBm条目的问题。
*   [2025-05-07 18:00:00] - **PC端分析引擎:** 增强了HT/VHT能力和带宽的解析，并改进SSID的UTF-8验证。修改了 `pc_analyzer/frame_parser/parser.go` 和 `pc_analyzer/state_manager/manager.go` 以解决Web UI不显示BSS/STA信息的问题。
*   [2025-05-07 18:16:00] - **Web前端:** 修复了因类型定义错误和状态更新逻辑不正确导致的BSS/STA信息无法显示的问题。修改了 `types/data.ts`, `contexts/DataContext.tsx`, 和 `components/BssList/BssList.tsx`。

*   [2025-05-07 17:13:00] - PC端分析引擎: 解决了 `pc_analyzer/frame_parser/parser.go` 中因添加诊断日志引入的编译错误 (如未定义的常量 `layers.Dot11FrameControl`, `layers.Dot11InformationElementIDHECapabilities`, QoS子类型)。
*   [2025-05-07 17:13:00] - PC端分析引擎: 解决了 `pc_analyzer/state_manager/manager.go` 中因 `ParsedFrameInfo` 结构调整 (移除 `FrameSubType`) 导致的编译错误。
*   [2025-05-07 17:13:00] - PC端分析引擎: 解决了 `pc_analyzer/frame_parser/parser_test.go` 中因 `parsePacketLayers` 函数签名变更导致的编译错误，并通过使用 `layers.LinkType(127)` 作为 `LinkTypeRadioTap` 的替代方案解决了后续的符号未定义问题。
*   [2025-05-07 16:40:00] - **Debugging Task Status Update:** Diagnosing empty `dot11.Payload` for Beacon frames (causes "expected 12, got 0" error). Added diagnostic logs to `pc_analyzer/frame_parser/parser.go` to trace Radiotap and Dot11 payload states. Awaiting user test and log feedback. (This task is now superseded by the next current task, but the log addition was completed).
*   [2025-05-07 16:24:00] - **TDD完成:** 为 `pc_analyzer/frame_parser/parser.go` 中的 `parsePacketLayers` 函数编写了全面的单元测试 (`pc_analyzer/frame_parser/parser_test.go`)，覆盖了多种管理帧类型偏移、IE解析鲁棒性（包括payload过短、IE头部不完整、IE长度无效）以及SSID提取（包括隐藏SSID、无SSID、多IE）等场景。所有测试均已通过。
*   [2025-05-07 15:33:12] - **任务完成:** 改进 `pc_analyzer/frame_parser/parser.go` 中的SSID及IE解析逻辑。实现了对 `MgmtMeasurementPilot`、`MgmtActionNoAck`、`MgmtReassociationReq` 等帧的payload偏移，并增强了IE解析循环的鲁棒性及日志记录。
*   [2025-05-07 15:01:00] - PC端分析引擎: 修复了Beacon/ProbeResp帧中因固定头部字段导致SSID解析错误的问题。调整了 `pc_analyzer/frame_parser/parser.go` 以正确偏移IE payload。
*   [2025-05-07 14:48:00] - PC端分析引擎: 重写了 `pc_analyzer/frame_parser/parser.go` 中的IE解析逻辑，简化为仅依赖 `dot11.Payload`，以解决编译错误并确保SSID解析路径的稳健性。
*   [2025-05-07 14:09:00] - PC端分析引擎: 修复了SSID IE解析逻辑，在`pc_analyzer/frame_parser/parser.go`中增加了调试日志并改进了对空SSID和隐藏SSID的处理。
*   [2025-05-07 13:24:06] - PC端分析引擎: 修复了 `radiotap layer not found` 错误。通过引入 `pcapgo` 包，实现了对gRPC流中pcap格式数据的正确解析。修改了 `pc_analyzer/grpc_client/client.go`, `pc_analyzer/frame_parser/parser.go`, 和 `pc_analyzer/main.go`。
*   [2025-05-07 12:18:00] - PC端分析引擎: `pc_analyzer/capture_agent.proto` 文件已修改，调整了 `go_package` 选项以准备修正服务名。
*   [2025-05-07 12:18:00] - PC端分析引擎: 诊断了与路由器代理之间的gRPC "Unimplemented" 错误。根本原因已定位为客户端 `.proto` 文件中 `go_package` 选项导致的生成代码服务名不匹配。
*   [2025-05-07] - PC端分析引擎: 针对 "Unknown WebSocket control command: (空命令)" 问题，在 `pc_analyzer/main.go` 中添加了详细的调试日志，以帮助诊断命令解析失败的根本原因。
*   [2025-05-07 11:34:00] - Web前端: 修改 `start_capture` 命令以在其 `payload` 中包含 `interface` (硬编码为 "ath1"), `channel`, 和 `bandwidth`。更新了 `ControlPanel.tsx` 和 `types/data.ts`。
*   [2025-05-07 11:34:00] - Memory Bank: `memory-bank/developmentContext/webFrontend.md` 已更新，记录了上述 `start_capture` payload 的修改。
*   [2025-05-07 11:30:00] - PC端分析引擎: 再次修复了WebSocket控制指令解析问题。强制要求 "start_capture" 命令的 `payload` 中必须包含 `InterfaceName`。(`pc_analyzer/main.go` 已更新)。
*   [2025-05-07 11:30:00] - Memory Bank: `memory-bank/developmentContext/pcAnalysisEngine.md`, `memory-bank/decisionLog.md`, `memory-bank/activeContext.md` 已更新，记录了上述WebSocket问题的再次分析和修复详情。
*   [2025-05-07] - PC端分析引擎: 修复了WebSocket控制指令解析问题。引擎现在可以正确处理来自Web前端的 `start_capture` 等指令，兼容 `action` 和 `command` 字段，并正确解析嵌套的 `payload`。(`pc_analyzer/main.go` 已更新)。
*   [2025-05-07] - Memory Bank: `memory-bank/developmentContext/pcAnalysisEngine.md` 已更新，记录了上述WebSocket问题的分析和修复详情。
*   [2025-05-07 02:43:00] - Web前端: 进一步修复了 `ControlPanel` UI对齐错误 (Issue 2 - 用户反馈问题2)，确保 "Set Channel" 和 "Set Bandwidth" 按钮在窄屏幕下正确换行。修改了 `ControlPanel.css` 中的 `.control-group` 规则。
*   [2025-05-07 02:35:00] - Router Agent: 解决了 `router_agent` 针对 `linux/arm64` 的交叉编译问题 (涉及包声明、Go版本、gRPC版本及 `protoc` 生成配置)。
*   [2025-05-07 02:35:00] - Memory Bank: 更新了 `memory-bank/deployment/routerAgentDeployment.md` 和 `memory-bank/developmentContext/routerAgent.md` 以反映交叉编译修复和最新指导。
*   [2025-05-07 02:19:00] - Web前端: 修复了UI对齐错误 (Issue 2.1 - ControlPanel按钮溢出)。
*   [2025-05-07 02:19:00] - Web前端: 修复了信道列表不匹配问题 (Issue 2.2 - ControlPanel更新为5GHz信道列表)。
*   [2025-05-07 02:19:00] - Web前端: 修复了BssList组件运行时错误 (Issue 2.3 - TypeError reading length)。
*   [2025-05-07 02:19:00] - Memory Bank: `memory-bank/developmentContext/webFrontend.md` 已更新相关修复详情。
*   [2025-05-07 02:04:00] - 系统集成: 端到端集成测试计划 (`memory-bank/testing/integrationTestPlan.md`) 设计与文档编写完成。
*   [2025-05-07 01:56:00] - Web前端: 详细定义了Web前端与PC端引擎间的WebSocket消息格式 (记录于 `memory-bank/developmentContext/webFrontend.md`)。
*   [2025-05-07 01:56:00] - Web前端: 可视化界面核心功能初步实现完成 (React项目搭建, WebSocket连接, BSS/STA列表展示, 控制面板)。
*   [2025-05-07 01:56:00] - Web前端: `memory-bank/developmentContext/webFrontend.md` 已创建并记录Web前端开发细节。
*   [2025-05-07 01:44:00] - PC端引擎: 各模块已整合到 `pc_analyzer/main.go` 中，形成完整引擎初步版本。
*   [2025-05-07 01:42:00] - PC端引擎: gRPC客户端流式通信逻辑 (`pc_analyzer/grpc_client/client.go`) 实现完成。
*   [2025-05-07 01:39:00] - PC端引擎: WebSocket服务器的数据推送和控制指令接收逻辑 (`pc_analyzer/websocket_server/server.go`) 实现完成。
*   [2025-05-07 01:35:00] - PC端引擎: BSS/STA状态管理核心逻辑 (`pc_analyzer/state_manager/manager.go`) 实现完成。
*   [2025-05-07 01:27:00] - PC端引擎: 802.11帧解析逻辑 (`pc_analyzer/frame_parser/parser.go` 使用 `gopacket`) 初步实现完成。
*   [2025-05-07 01:08:00] - PC端引擎: 配置加载逻辑 (`pc_analyzer/config/config.go`, `pc_analyzer/config/config.json`) 实现完成。
*   [2025-05-07 00:46:00] - PC端实时分析引擎 (PC-Side Real-time Analysis Engine) Go语言项目骨架创建完成，包括目录结构、模块占位文件、gRPC代码生成及初步整理 (`pc_analyzer/` 目录下)。
*   [2025-05-07 00:46:00] - `pc_analyzer/capture_agent.proto` (go_package 更新), `pc_analyzer/router_agent_pb/capture_agent.pb.go`, `pc_analyzer/router_agent_pb/capture_agent_grpc.pb.go`, `pc_analyzer/main.go`, `pc_analyzer/grpc_client/client.go` 等文件已创建/更新。
*   [2025-05-07 00:46:00] - `memory-bank/developmentContext/pcAnalysisEngine.md` 已创建并记录PC端引擎开发细节。
*   [2025-05-07 00:18:00] - 路由器端抓包代理 (Router-Side Capture Agent) Go语言实现初步完成 (gRPC接口定义,核心逻辑, tcpdump集成)。
*   [2025-05-07 00:18:00] - `router_agent.proto`, `router_agent/main.go`, `router_agent/capture_agent.pb.go`, `router_agent/capture_agent_grpc.pb.go`, `router_agent/go.mod` 已创建/更新。
*   [2025-05-06 23:48:00] - Memory Bank 初始化。
*   [2025-05-06 23:48:00] - 初步系统架构设计完成。
*   [2025-05-06 23:48:00] - 关键架构决策已记录到 `decisionLog.md`。
*   [2025-05-06 23:48:00] - 项目上下文已更新到 `productContext.md`。
*   [2025-05-06 23:48:00] - 当前活动已更新到 `activeContext.md`。

## Current Tasks
*   [2025-05-08 13:04:00] - **Web前端UI细节调整 (根据用户反馈):**
    *   (已完成) 解决了控制面板折叠后，BSS 和 STA 列表未填充释放空间的问题。
    *   (已完成) 确保了 BSS 条目展开后，Security 信息单独一行显示。
    *   (已完成) 调整了 BSS 和 STA 列表的宽度比例为约 2:3。
    *   (已完成) 调整了控制面板展开时的宽度，以确保内容完全显示。
*   [2025-05-08 04:23:00] - **Web前端 UI/UX Redesign (Phase 2):**
    *   根据需要创建新的通用/原子组件 (例如 `Button`, `Card`, `Table`, `Icon`)。
    *   确保字体 (SF Pro) 和 SVG 图标按照架构文档正确引入和管理。
    *   进行全面的代码审查，确保所有更改都符合项目编码规范和代码质量标准。
    *   进行用户界面测试，确保所有组件在新设计下正常工作并符合WCAG AA对比度标准。
*   [2025-05-07 21:15:35] - **核心诊断任务:** 等待用户测试 `pc_analyzer` 以验证对 0dBm STA 问题的修复是否有效。如果问题仍然存在，需要进一步分析 `frame_parser.go` 是否总是为某些帧返回0信号强度，或者检查 `models.go` 中 `NewSTAInfo()` 的默认信号强度值。
*   [2025-05-07 17:13:00] - **核心诊断任务:** 等待用户运行 `pc_analyzer` 并提供新的日志。分析日志中的 `DEBUG_PACKET_LAYERS` 和 `ERROR_NO_DOT11_LAYER` 等信息，以确定 Beacon 帧是否被 `gopacket` 正确解析出 `Dot11` 层。这是解决SSID解析问题的关键步骤。 (此任务可能因后续修复而部分解决或改变焦点)
*   等待用户执行端到端测试并反馈结果，特别是关注SSID的正确性。
*   根据测试结果记录到 `memory-bank/testing/integrationTestResults.md`。
*   **待办:** `pc_analyzer` 项目中的gRPC Go代码 (`pc_analyzer/router_agent_pb/`) 需要使用 `protoc` 命令重新生成，以使 `.proto` 文件中 `go_package` 的修正生效 (关联gRPC "Unimplemented" 错误修复)。

## Next Steps

*   根据日志分析结果，进一步定位并修复 Beacon 帧解析问题。
*   完善各组件的文档和部署说明。
---
**Task:** Provide cross-compilation guidance for router_agent (linux/aarch64)
**Status:** Completed
**Timestamp:** 2025/5/7 上午2:12:43
**Details:** Documented detailed steps for cross-compiling the Go-based router_agent for a Linux aarch64 target, including environment variables, build commands, output verification, and transfer methods. Stored in `memory-bank/deployment/routerAgentDeployment.md`.
*   [2025-05-08 16:51:29] - **后端日志增强完成:**
    *   已在 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0) 和 [`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go:0) 中添加了详细的 DEBUG 日志记录，用于追踪吞吐量和信道占空比计算的关键变量和中间值。
    *   此举旨在帮助诊断前端指标显示 "N/A" 的问题。
    *   Memory Bank (`activeContext.md`) 已同步更新此变更。
* [2025-05-08 17:05:17] - **后端日志删减完成:**
    *   已注释掉 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0) 和 [`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go:0) 中为调试 "N/A" 问题添加的大量高频 DEBUG 日志。
    *   修复了因注释日志导致的编译错误。
    *   Memory Bank (`activeContext.md`, `progress.md`) 已更新。

*   [2025-05-08 17:42:00] - **`gopacket` 解析错误诊断:**
    *   分析了用户提供的 `gopacket` 解析错误日志。
    *   评估了这些错误对信道占空比 (`Duration/ID`) 和吞吐量 (`TransportPayloadLength`) 计算的直接影响。
    *   提出了问题原因的假设，包括数据包本身问题、`gopacket` 限制/Bug 以及当前解析逻辑问题。
    *   建议了后续步骤，包括增强错误处理、使用 Wireshark 分析 pcap 文件、审查 `gopacket` 用法及关注 Radiotap 解析。
    *   更新了 [`memory-bank/activeContext.md`](memory-bank/activeContext.md:0) 以反映分析结果。
*   [2025-05-08 17:46:00] - **后端解析器鲁棒性增强完成:**
    *   修改了 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0) 以增强 `gopacket` 解析错误的错误处理。
    *   当 `packet.ErrorLayer()` 返回错误或 `Dot11` 层无法解析时，现在会返回错误，阻止这些数据包被进一步处理。
    *   `gopacket.NewPacket` 调用已更改为使用 `gopacket.Lazy` 解码选项。
    *   此更改旨在提高解析器在遇到损坏或异常数据包时的健壮性，并减少因解析失败导致的下游指标计算问题。