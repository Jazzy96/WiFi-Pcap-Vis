# Active Context

This file tracks the project's current status, including recent changes, current goals, and open questions.

*   [2025-05-08 13:04:00] - **当前焦点:** 完成用户请求的 UI 细节调整，包括控制面板展开宽度。
*   [2025-05-08 13:04:00] - **近期变更:**
    *   Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0) 更新，调整了控制面板展开时的宽度比例为 `0.8fr`，并保持 BSS 和 STA 列表的 `2fr` 和 `3fr` 比例。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
*   [2025-05-08 12:51:00] - **当前焦点:** 完成用户请求的 UI 细节调整。
*   [2025-05-08 12:51:00] - **近期变更:**
    *   Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0) 更新，动态调整 `grid-template-columns` 以解决控制面板折叠后的空间分配问题，并调整了 BSS 和 STA 列表的宽度比例为约 2:3 (`0.5fr 2fr 3fr`)。
    *   Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx:0) 更新，为 Security 字段添加了 `fullWidthField` 类。
    *   Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css:0) 更新，添加了 `.fullWidthField` 类定义，使 Security 字段在展开的 BSS 详情中单独成行。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
*   [2025-05-08 12:26:00] - **当前焦点:** 根据用户反馈调整应用布局和视觉样式。
*   [2025-05-08 12:26:00] - **近期变更:**
    *   Web前端: 调整了 [`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0) 中的三列布局 (`grid-template-columns`)，为 BSS 和 STA 列表设置最小宽度，以防止在默认窗口大小下被挤压。
    *   Web前端: 为 [`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0) 中的 `.App` 添加了浅灰色背景，以区分白色卡片。
    *   Web前端: 增强了 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css:0) 中选中 BSS 的高亮效果（添加背景色）。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
*   [2025-05-08 11:50:00] - **当前焦点:** 根据用户反馈调整 BSS 和 STA 列表的样式和布局。
*   [2025-05-08 11:50:00] - **近期变更:**
    *   Web前端: 移除了 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx:0) 中的展开/折叠指示器。
    *   Web前端: 更新了 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css:0) 以添加选中项高亮效果，并调整了 BSS 摘要中 Signal/Ch/STAs 字段的布局和宽度。
    *   Web前端: 将 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx:0) 从使用 `Table` 组件改回为使用 `Card` 组件展示每个 STA。
    *   Web前端: 更新了 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.module.css:0) 以适应新的 STA 卡片布局。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
*   [2025-05-08 11:26:00] - **当前焦点:** 根据用户反馈调整应用布局。
*   [2025-05-08 11:26:00] - **近期变更:**
    *   Web前端: 从 [`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0) 中移除了顶部 header。
    *   Web前端: 更新了 [`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0) 以移除 header 样式，并将主内容区调整为三列水平布局 (`ControlPanel`, `BssList`, `StaList`)。为 `ControlPanel` 的折叠状态添加了过渡效果和最小宽度。
    *   Web前端: 更新了 [`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0) 中的容器类名以匹配 `App.css` 中的更改。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
*   [2025-05-08 11:17:00] - **当前焦点:** 完成“企业级 Wi-Fi 抓包分析软件” UI/UX 重新设计的剩余部分。
*   [2025-05-08 11:17:00] - **近期变更:**
    *   Web前端: 创建了通用UI组件 (`Button`, `Input`, `Card`, `Table`, `Tabs`, `Icon`) 于 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/common/`](desktop_app/WifiPcapAnalyzer/frontend/src/components/common/) 目录。
    *   Web前端: 更新了 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.tsx:0) 以使用新的通用 `Button` 和 `Input` 组件。
    *   Web前端: 更新了 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx:0) 以使用新的通用 `Card` 组件。
    *   Web前端: 更新了 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx:0) 以使用新的通用 `Table` 组件。
    *   Web前端: 确认了旧 CSS 文件 ([`desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.css:0), [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.css:0), [`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.css:0)) 不再被引用，并已成功删除。
    *   Web前端: 确认了 [`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0) 和 [`desktop_app/WifiPcapAnalyzer/frontend/src/index.css`](desktop_app/WifiPcapAnalyzer/frontend/src/index.css:0) 符合新的 UI/UX 规范。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
*   [2025-05-08 04:22:00] - **当前焦点:** 实施“企业级 Wi-Fi 抓包分析软件”的 UI/UX 重新设计。已完成全局样式、基础布局以及核心组件 (`ControlPanel`, `BssList`, `StaList`) 的样式重构和到 CSS Modules 的迁移。
*   [2025-05-08 04:22:00] - **近期变更:**
    *   Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/index.css`](desktop_app/WifiPcapAnalyzer/frontend/src/index.css:0) 已更新，定义了全局 CSS 变量和基础 HTML 元素样式，符合新的 UI/UX 规范。
    *   Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0) 已更新，以实现新的网格布局和排版规则。
    *   Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.tsx:0) 已更新，使用新的 CSS Modules 文件 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/ControlPanel/ControlPanel.module.css:0) 并应用了新的 UI/UX 设计。旧的 `ControlPanel.css` 文件已不再直接使用。
    *   Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx:0) 已更新，使用新的 CSS Modules 文件 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css:0) 并应用了新的 UI/UX 设计。旧的 `BssList.css` 文件已不再直接使用。
    *   Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx:0) 已更新，使用新的 CSS Modules 文件 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.module.css:0) 并应用了新的 UI/UX 设计。旧的 `StaList.css` 文件已不再直接使用。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
*   [2025-05-08 04:00:00] - **先前焦点:** 完成了针对“企业级 Wi-Fi 抓包分析软件” UI/UX 重新设计的前端架构方案制定。方案包括样式管理策略（CSS 模块、CSS 变量）、组件架构（原子化、组合化）、字体与资源管理（SF Pro, SVG 图标）以及可维护性考虑。
*   [2025-05-08 04:00:00] - **先前变更:**
    *   Memory Bank: `memory-bank/activeContext.md` 更新，记录前端架构方案制定完成。
    *   Memory Bank: `memory-bank/developmentContext/webFrontend.md` 更新以反映新的前端架构决策。
*   [2025-05-08 03:58:00] - **先前焦点:** 为 "企业级 Wi-Fi 抓包分析软件" 定义详细的 UI/UX 重新设计规范和伪代码。要求包括高端、极简、专业风格，特定配色方案（石墨灰、雾面白、科技蓝），网格布局，SF Pro 字体，8px 圆角，轻量阴影，并遵循 WCAG AA 对比度标准。
*   [2025-05-08 03:58:00] - **先前变更:**
    *   Memory Bank: `memory-bank/productContext.md` 已更新，添加了新的 UI/UX 设计规范章节。
*   [2025-05-07 18:00:00] - **先前焦点:** Web UI 仍未显示 BSS/STA 信息。已增强后端 `pc_analyzer` 对 HT/VHT 能力和带宽的解析，并改进了SSID的UTF-8验证。修改了 `pc_analyzer/frame_parser/parser.go` 和 `pc_analyzer/state_manager/manager.go`。等待用户测试以确认UI是否正常显示。
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

## Recent Changes
*   [2025-05-08 13:04:00] - Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0) 更新，调整了控制面板展开时的宽度比例为 `0.8fr`，并保持 BSS 和 STA 列表的 `2fr` 和 `3fr` 比例。
*   [2025-05-08 12:51:00] - Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0) 更新，动态调整 `grid-template-columns` 以解决控制面板折叠后的空间分配问题，并调整了 BSS 和 STA 列表的宽度比例为约 2:3 (`0.5fr 2fr 3fr`)。
*   [2025-05-08 12:51:00] - Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx:0) 更新，为 Security 字段添加了 `fullWidthField` 类。
*   [2025-05-08 12:51:00] - Web前端: [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.module.css:0) 更新，添加了 `.fullWidthField` 类定义，使 Security 字段在展开的 BSS 详情中单独成行。
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
*   [2025-05-07 22:16:00] - **Debug Status Update:** Applied fix to `pc_analyzer/state_manager/manager.go` to prevent multicast/broadcast MAC addresses (from data frame DA/RA) from being incorrectly registered as STAs. Added `isUnicastMAC` check for `parsedInfo.RA` when inferring `staMAC` in data frame processing logic.
*   [2025-05-07 22:35:00] - **Debug Status Update (Frontend):** Web frontend (`web_frontend`) failed to start with `npm start` due to `Error: error:0308010C:digital envelope routines::unsupported` on Node.js v23.9.0. Resolved by running `NODE_OPTIONS=--openssl-legacy-provider npm start`. Frontend compiled with some ESLint warnings (unused variables) but development server started. User terminated the process (SIGINT), presumably after confirming startup.
*   [2025-05-07 22:38:00] - **Debug Status Update (Frontend):** Modified `web_frontend/package.json` to include `NODE_OPTIONS=--openssl-legacy-provider` directly in the `scripts.start` command. This provides a persistent fix for the OpenSSL compatibility issue with Node.js v23.9.0, so `npm start` can be used without manually prepending the option.
*   [2025-05-07 22:41:00] - **Debug Status Update (Frontend):** Cleaned up ESLint warnings for unused variables:
    *   In `web_frontend/src/components/StaList/StaList.tsx`: Removed unused import `BSS` and unused variable `allStas`.
    *   In `web_frontend/src/components/BssList/BssList.tsx`: Removed unused component definition `StaListItem` and its props interface `StaListItemProps`.
*   [2025-05-07 22:43:00] - **Debug Status Update (Frontend):** Cleaned up an additional ESLint warning in `web_frontend/src/components/BssList/BssList.tsx` by removing the unused import `STA`.
*   [2025-05-07 23:00:00] - **Debug Status Update (PC Analyzer):** Modified `pc_analyzer/frame_parser/parser.go` in `parsePacketLayers` function. For `MgmtBeacon` and `MgmtProbeResp` frames, if the `originalPayload` is shorter than the `fixedHeaderLen` (12 bytes), the function now returns `nil, error` instead of `info, nil`. This prevents `state_manager` from creating BSS entries based on these critically incomplete frames. This addresses the issue of BSSs being created with "(Hidden)" SSID and "Open" security due to truncated Probe Response frames.
*   [2025-05-07 23:06:00] - **Debug Status Update (PC Analyzer):** Modified `pc_analyzer/main.go` to increase the frequency of `PruneOldEntries` execution. The `pruneTicker` is now set to `30 * time.Second` (previously 1 minute), while the timeout for entries remains 5 minutes. This should make the aging out of old STA/BSS entries more responsive.
*   [2025-05-07 23:08:00] - **Debug Status Update (PC Analyzer):** Modified `pc_analyzer/main.go` to reduce the timeout for `PruneOldEntries`. The timeout argument passed to the function is now `2 * time.Minute` (previously 5 minutes). The pruning check frequency remains 30 seconds. This should result in faster aging out of old STA/BSS entries.
*   [2025-05-07 23:11:00] - **Debug Status Update (PC Analyzer):** Modified `pc_analyzer/state_manager/manager.go` in `ProcessParsedFrame` function. Added a check when creating a new BSS from a `MgmtBeacon` or `MgmtProbeResp`. If the parsed info lacks SSID (is empty, N/A, Hidden, or Invalid), RSN info, and both HT/VHT capabilities, the state manager will now log a warning and skip creating the new BSS entry. This aims to prevent polluting the BSS list with entries derived from frames where IE parsing failed significantly (e.g., due to malformed vendor IEs).
*   [2025-05-07 23:24:00] - **Debug Status Update (PC Analyzer):** Modified `pc_analyzer/frame_parser/parser.go` to implement GBK fallback decoding for SSIDs. If the raw SSID bytes are not valid UTF-8, the parser now attempts to decode them using GBK. If GBK decoding is successful, the result is used; otherwise, the SSID is marked as `"<Invalid/Undecodable SSID>"`. Added `golang.org/x/text` dependency.
*   [2025-05-07 23:30:00] - **Debug Status Update (PC Analyzer):** Reverted the SSID decoding logic in `pc_analyzer/frame_parser/parser.go`. Removed the GBK fallback attempt based on user feedback that it did not resolve the issue and added complexity. The logic now simply checks for valid UTF-8; if invalid, the SSID is marked as `"<Invalid/Undecodable SSID>"`. Unused `golang.org/x/text/...` imports were also removed.