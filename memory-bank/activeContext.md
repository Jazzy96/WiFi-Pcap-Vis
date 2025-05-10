*   [2025-05-10 17:57:54] - **Debug Status Update (UI Display Issue):** Added detailed logging to [`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go) and [`desktop_app/WifiPcapAnalyzer/app.go`](desktop_app/WifiPcapAnalyzer/app.go) to investigate missing BSS updates and state snapshot event emissions, which are suspected causes for the UI not displaying BSS/STA information. Log `DEBUG_SM_BSS_UPDATE` added to BSS creation/update paths. Logs `DEBUG_APP_EVENT: Attempting to get snapshot and emit event.` and `DEBUG_APP_EVENT: Snapshot created. BSS count: %d, STA count: %d. Emitting event now.` added around snapshot generation and emission.
*   [2025-05-10 17:19:00] - **Debug Status Update (Parser Robustness):**
    *   Modified [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go) to improve tolerance for empty or malformed non-critical CSV fields from `tshark`.
    *   Specifically, parsing errors for `radiotap.channel.freq`, `radiotap.dbm_antsignal`, and `wlan.duration`, if the raw value was present but unparsable, will now be logged but will not cause the entire frame to be discarded. These fields will use their default zero-values in `ParsedFrameInfo`.
    *   The fallback for a failed `frame.time_epoch` parse was changed from `time.Now()` to `time.Time{}` (zero value), though a failure here still discards the frame.
    *   This aims to increase the number of frames successfully processed and sent to the state manager, addressing the issue of no data appearing on the frontend.
    *   Memory Bank: `decisionLog.md` and `progress.md` updated.
* [2025-05-10 16:35:04] - **Debug Status Update (CSV Data Parsing):**
    *   Added detailed logging to `ProcessRow` in [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go) to trace CSV row data, individual field parsing attempts/values, errors during extraction/conversion, and successfully parsed `ParsedFrameInfo` summaries.
    *   Also added logs before calling `packetInfoHandler` in `ProcessPcapFile` and `ProcessPcapStream`.
    *   Corrected `log.Errorf` to `log.Printf` to resolve compilation errors.
    *   This aims to help pinpoint errors occurring during the data row parsing stage.
    *   Memory Bank: `decisionLog.md` and `progress.md` updated.
*   [2025-05-10 16:17:00] - **Debug Status Update (tshark fields):**
    *   Corrected and removed problematic `tshark` fields in [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go) based on error logs and `tshark_beacon_example.json`.
    *   Specifically, `wlan.flags.retry` changed to `wlan.fc.retry`.
    *   Removed `radiotap.mcs.flags`, `radiotap.vht.mcs`, `radiotap.vht.nss`, `radiotap.he.mcs`, `radiotap.he.bw`, `radiotap.he.gi`, `radiotap.he.nss`, and `wlan.he.phy.channel_width_set`.
    *   This aims to resolve `tshark` execution errors.
    *   Memory Bank: `decisionLog.md` updated with this decision. `progress.md` to be updated.
# Active Context

*   [2025-05-10 15:30:00] - **当前焦点:** 完成 `gopacket` 到 `tshark` 解析逻辑的迁移。
*   [2025-05-10 15:30:00] - **近期变更 (tshark 解析迁移):**
    *   文件 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go): 完全重写以实现基于 `tshark` 的解析。引入 `TSharkExecutor`, `CSVParser`, `FrameProcessor`。`ProcessPcapFile` (原 `ProcessPcapStream`) 更新。`ParsedFrameInfo` 调整。
    *   文件 [`desktop_app/WifiPcapAnalyzer/config/config.go`](desktop_app/WifiPcapAnalyzer/config/config.go): 添加 `TsharkPath` 配置项。
    *   文件 [`desktop_app/WifiPcapAnalyzer/config/config.json`](desktop_app/WifiPcapAnalyzer/config/config.json): 添加 `tshark_path` 默认值。
    *   文件 [`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go): 修复因 `ParsedFrameInfo` 结构和类型更改引起的编译错误。
    *   文件 [`desktop_app/WifiPcapAnalyzer/app.go`](desktop_app/WifiPcapAnalyzer/app.go): 更新对 `frame_parser.ProcessPcapFile` 的调用，并处理 `TsharkPath` 配置。
    *   文件 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser_test.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser_test.go): 已删除，因其测试内容基于旧的 `gopacket` 实现。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
    *   Memory Bank: `memory-bank/decisionLog.md` 即将更新（记录迁移到 tshark 的实现细节）。

---
(Existing content will follow this new entry)
# Active Context

This file tracks the project's current status, including recent changes, current goals, and open questions.

*   [2025-05-08 16:37:00] - **当前焦点:** 完成信道占空比计算方法的重构，使用 MAC Duration/ID 字段。
*   [2025-05-08 16:37:00] - **近期变更 (信道占空比重构):**
    *   帧解析器 ([`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0)): `ParsedFrameInfo` 添加 `MACDurationID` 字段，并在解析时从 `layers.Dot11.DurationID` 填充。
    *   数据模型 ([`desktop_app/WifiPcapAnalyzer/state_manager/models.go`](desktop_app/WifiPcapAnalyzer/state_manager/models.go:0)): `BSSInfo` 添加 `AccumulatedNavMicroseconds` 字段用于累积 NAV 时间。
    *   状态管理器 ([`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go:0)):
        *   `ProcessParsedFrame`: 修改为累积 `MACDurationID` 到 `AccumulatedNavMicroseconds`，并添加了对 PS-Poll 帧 (`layers.Dot11TypeCtrlPowersavePoll`) 的排除逻辑。移除了基于 `CalculateFrameAirtime` 的旧累积逻辑。
        *   `PeriodicallyCalculateMetrics`: 修改为使用 `AccumulatedNavMicroseconds` 计算 `ChannelUtilization`，并在计算后重置该累加器。
    *   Memory Bank: `memory-bank/decisionLog.md` 已记录此重构决策。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
*   [2025-05-08 15:52:00] - **当前焦点:** 调试前端显示高级功能（实时信道占空比和吞吐量）为 "N/A" 的问题。
*   [2025-05-08 15:52:00] - **近期变更 (后端调试):**
    *   诊断 "N/A" 问题：主要原因定位为 [`desktop_app/WifiPcapAnalyzer/state_manager/models.go`](desktop_app/WifiPcapAnalyzer/state_manager/models.go:0) 中 `BSSInfo` 和 `STAInfo` 的 `lastCalcTime` 字段未初始化，导致初次指标计算为0。
    *   修复：已修改 [`desktop_app/WifiPcapAnalyzer/state_manager/models.go`](desktop_app/WifiPcapAnalyzer/state_manager/models.go:0) 中的 `NewBSSInfo` 和 `NewSTAInfo` 函数，将 `lastCalcTime` 初始化为 `time.Now()`。
    *   复杂度评估：记录了当前信道占空比和吞吐量计算方法的实现方式及其复杂程度。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
    *   Memory Bank: `memory-bank/decisionLog.md` 即将记录此调试决策。
*   [2025-05-08 14:48:00] - **当前焦点:** 完成新功能（实时信道占空比和吞吐量分析）的前端实现，应用于 Wails 应用 (`desktop_app/WifiPcapAnalyzer/frontend/`)。
*   [2025-05-08 14:48:00] - **近期变更 (前端性能分析功能):**
    *   TypeScript类型: [`desktop_app/WifiPcapAnalyzer/frontend/src/types/data.ts`](desktop_app/WifiPcapAnalyzer/frontend/src/types/data.ts:0) 中的 `BSS` 和 `STA` 接口已扩展，以包含新的性能指标字段（如 `channel_utilization_percent`, `total_throughput_mbps`, `throughput_ul_mbps`, `throughput_dl_mbps` 和相应的 `historical_` 字段）。
    *   DataContext: [`desktop_app/WifiPcapAnalyzer/frontend/src/contexts/DataContext.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/contexts/DataContext.tsx:0) 已更新，添加了 `selectedPerformanceTarget` 状态及相应的 action 和 reducer 逻辑，用于管理在详细性能面板中显示哪个 BSS 或 STA 的数据。
    *   BSS列表: [`desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/BssList/BssList.tsx:0) 已修改，以在其列表项中显示信道占空比和总吞吐量的当前值，并在点击时更新 `selectedPerformanceTarget`。
    *   STA列表: [`desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/StaList/StaList.tsx:0) 已修改，以在其列表项中显示上下行吞吐量的当前值，并在点击时更新 `selectedPerformanceTarget`。
    *   PerformanceDetailPanel: 创建了新的组件 [`desktop_app/WifiPcapAnalyzer/frontend/src/components/PerformanceDetailPanel/PerformanceDetailPanel.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/components/PerformanceDetailPanel/PerformanceDetailPanel.tsx:0) 和对应的 CSS Module [`desktop_app/WifiPcapAnalyzer/frontend/src/components/PerformanceDetailPanel/PerformanceDetailPanel.module.css`](desktop_app/WifiPcapAnalyzer/frontend/src/components/PerformanceDetailPanel/PerformanceDetailPanel.module.css:0)。此组件使用 `recharts` 图表库显示所选 BSS 或 STA 的详细性能指标和历史趋势图。
    *   App布局: [`desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx`](desktop_app/WifiPcapAnalyzer/frontend/src/App.tsx:0) 和 [`desktop_app/WifiPcapAnalyzer/frontend/src/App.css`](desktop_app/WifiPcapAnalyzer/frontend/src/App.css:0) 已调整，以在主界面右侧集成 `PerformanceDetailPanel`，并在有性能目标被选中时动态调整为四列布局。
    *   依赖安装: 已将 `recharts` 和 `@types/recharts` 添加到 [`desktop_app/WifiPcapAnalyzer/frontend/package.json`](desktop_app/WifiPcapAnalyzer/frontend/package.json:0)。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
    *   Memory Bank: `memory-bank/developmentContext/webFrontend.md` 即将更新图表库选择。
*   [2025-05-08 14:36:00] - **当前焦点:** 完成新功能（实时信道占空比和吞吐量分析）的后端实现，应用于 Wails 应用 (`desktop_app/WifiPcapAnalyzer/`)。
*   [2025-05-08 14:36:00] - **近期变更:**
    *   Wails后端: [`desktop_app/WifiPcapAnalyzer/state_manager/models.go`](desktop_app/WifiPcapAnalyzer/state_manager/models.go:0) 扩展了 `BSSInfo` 和 `STAInfo` 结构，添加了信道占空比和吞吐量相关指标字段（当前值、历史数据、内部计算字段）。
    *   Wails后端: [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0) 添加了 `CalculateFrameAirtime` 辅助函数（简化模型），并在 `ParsedFrameInfo` 中添加了 `FrameLength`, `PHYRateMbps`, `IsShortPreamble`, `IsShortGI` 字段及 `getPHYRateMbps` 辅助函数（简化模型）。
    *   Wails后端: [`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go:0) 更新了 `StateManager` 初始化逻辑，在 `ProcessParsedFrame` 中添加了通话时间和传输字节数的累积逻辑，并实现了 `PeriodicallyCalculateMetrics` 方法用于定期计算和更新指标及历史数据。更新了 `GetSnapshot` 以进行深拷贝。
    *   Wails后端: [`desktop_app/WifiPcapAnalyzer/app.go`](desktop_app/WifiPcapAnalyzer/app.go:0) 在 `startup` 方法中更新了 `StateManager` 的初始化，并添加了一个新的 goroutine 来定期调用 `PeriodicallyCalculateMetrics`。
    *   Wails后端: 确认了通过 `runtime.EventsEmit` 推送快照的逻辑无需修改即可包含新指标。
    *   Memory Bank: `memory-bank/activeContext.md` 本次更新。
    *   Memory Bank: `memory-bank/progress.md` 即将更新。
*   [2025-05-08 13:04:00] - **先前焦点:** 完成用户请求的 UI 细节调整，包括控制面板展开宽度。
*   [2025-05-08 13:04:00] - **先前变更:**
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
## [2025-05-08 下午4:44:00] - TDD 任务调整与 "N/A" 问题诊断

**当前状态:** 用户反馈 BSS 和 STA 的吞吐量及信道利用率指标在前端显示为 "N/A"。已确认 `lastCalcTime` 初始化问题已修复，但问题依旧存在。

**TDD 任务调整:**
*   暂停原定的测试用例编写计划。
*   转向辅助用户进行真实场景测试以诊断 "N/A" 问题。
*   识别代码中可能导致 "N/A" 的潜在错误点，并提出修复建议。

**已识别的潜在问题区域:**
1.  **数据累积:**
    *   `TransportPayloadLength` (用于吞吐量) 可能未被正确解析或始终为0。
    *   `MACDurationID` (用于BSS信道占空比) 可能未被正确解析、始终为0，或PS-Poll排除逻辑问题。
2.  **STA 信道占空比:** 仍使用旧的 `totalAirtime` 逻辑，可能与新调整不一致，或 `totalAirtime` 本身累积不正确。

**当前行动计划:**
1.  向用户明确指出了需要在 `parser.go` 和 `manager.go` 中关注和添加的 DEBUG 日志点，以追踪关键变量 (`TransportPayloadLength`, `MACDurationID`, 累积字节数，累积NAV时间，以及计算出的各项指标) 的值。
2.  等待用户提供带有这些详细日志的程序输出来进一步分析。
3.  建议对 STA 的信道占空比计算逻辑进行统一，使其也基于 `MACDurationID`。

**待观察的关键日志信息:**
*   `parser.go`: `DEBUG_PACKET_LAYERS`, `DEBUG_DOT11_INFO`, `DEBUG_FRAME_PARSER_SUMMARY`。
*   `manager.go`: `DEBUG_METRIC_ACCUM` (需添加), `DEBUG_NAV_SKIP`, `DEBUG_METRIC_CALC_BSS_PRE`/`_POST` (需添加), `DEBUG_METRIC_CALC_STA_PRE`/`_POST` (需添加)。
* [2025-05-08 16:50:59] - **代码变更 (后端日志增强):**
    *   为诊断前端指标 "N/A" 问题，在后端指标计算相关代码中添加了详细的 DEBUG 日志。
    *   文件 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0):
        *   记录了 `MACDurationID` 的解析值。
        *   记录了网络层/传输层协议识别情况及 `TransportPayloadLength`。
    *   文件 [`desktop_app/WifiPcapAnalyzer/state_manager/manager.go`](desktop_app/WifiPcapAnalyzer/state_manager/manager.go:0):
        *   `ProcessParsedFrame`: 记录了用于吞吐量计算的字节数累积 (`totalTxBytes` for BSS, `totalUplinkBytes`/`totalDownlinkBytes` for STA) 和用于信道占空比计算的 NAV 时间累积 (`AccumulatedNavMicroseconds` for BSS)，包括对 PS-Poll 帧的跳过逻辑。
        *   `PeriodicallyCalculateMetrics`: 记录了函数开始时的输入参数，BSS/STA 吞吐量和 BSS 信道占空比计算的中间值（累积值、时间窗口）和最终结果。记录了更新 BSS/STA 指标的动作。
        *   `GetSnapshot`: 记录了准备推送到 WebSocket 的 BSS 和 STA 指标值。
    *   日志级别和格式：使用标准 `log` 包，添加了 `DEBUG_METRIC_ACCUM`, `DEBUG_METRIC_CALC_BSS_PRE`/`_POST`, `DEBUG_METRIC_CALC_STA_PRE`/`_POST`, `DEBUG_SNAPSHOT_BSS`, `DEBUG_SNAPSHOT_STA` 等前缀。
* [2025-05-08 17:04:45] - 代码变更 (后端日志删减): 注释掉了在 parser.go 和 manager.go 中为诊断 "N/A" 问题添加的大量高频 DEBUG 日志，以减少日志输出。修复了由此产生的编译错误。

---
[2025-05-08 17:39:00] - **Debug Status Update: gopacket 解析错误导致指标 "N/A"**

**问题描述:**
用户报告后端日志中存在大量 `gopacket` 解析错误，导致吞吐量和信道占空比指标在前端显示为 "N/A"。

**错误分析:**
*   **关键错误类型:**
    *   `gopacket.NewPacket encountered an error: vendor extension size < 3`
    *   `gopacket.NewPacket encountered an error: Layer type not currently supported`
    *   `gopacket.NewPacket encountered an error: Dot11 length X too short, Y required` (及管理帧变体)
    *   `ERROR_NO_DOT11_LAYER: Dot11 layer is nil. Radiotap present: true.`
*   **对指标计算的直接影响:**
    *   这些错误，特别是 `ERROR_NO_DOT11_LAYER` 和 `Dot11 length X too short`，直接阻止了对 802.11 MAC 层的成功解析。
    *   无法解析 MAC 层导致无法提取 `Duration/ID` 字段（用于信道占空比计算，依赖 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:297`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:297)）。
    *   无法解析 MAC 层也使得无法继续解析更高层（LLC, IP, TCP/UDP），从而无法获取 `TransportPayloadLength`（用于吞吐量计算，依赖 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:607-632`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:607)）。
*   **错误集中性:** 错误较多地集中在管理帧（如 ProbeResp, AssocResp）及其信息元素（IEs）的解析上。但通用错误如 `ERROR_NO_DOT11_LAYER` 可能影响所有帧类型。

**问题原因假设:**
1.  **数据包本身问题:** pcap 文件中的数据包可能已损坏、不完整（如捕获截断）或格式异常（非标准帧）。
2.  **`gopacket` 限制/Bug:** `gopacket` 库在处理特定类型的 802.11 帧、Radiotap 头或供应商特定扩展时可能存在限制或 bug。Radiotap 解析问题也可能误导 Dot11 层定位。
3.  **解析逻辑问题 ([`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0)):** 当前 `gopacket.Default` 解码选项可能不够容错；错误捕获后的恢复逻辑可能不完善。

**建议的后续步骤:**
1.  **增强错误处理:** 在 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0) 中，对于导致无法解析 Dot11 层的严重错误，应更明确地跳过该数据包，避免其影响后续指标累加。
2.  **pcap 文件分析:** 强烈建议用户使用 Wireshark 打开原始 pcap 文件，检查报告错误的特定数据包，确认其结构是否异常，以及 Wireshark 是否也报类似错误。
3.  **`gopacket` 用法审查:** 考虑是否有更容错的 `gopacket` 解码选项（如 `gopacket.Lazy`，需谨慎评估副作用）。
4.  **关注 `Radiotap` 解析:** 深入调查 `ERROR_NO_DOT11_LAYER` 是否与 Radiotap 头解析的准确性有关。

**次要关注点:**
*   `TRA | No listeners for event 'state_snapshot'` 消息提示可能存在 WebSocket 通信问题，待主要解析问题解决后再关注。
*   [2025-05-08 17:46:00] - **代码变更 (后端解析器鲁棒性增强):**
    *   文件 [`desktop_app/WifiPcapAnalyzer/frame_parser/parser.go`](desktop_app/WifiPcapAnalyzer/frame_parser/parser.go:0):
        *   `parsePacketLayers`: 当 `packet.ErrorLayer()` 返回非 `nil` 错误时，函数现在会记录错误并返回 `nil, error`，以阻止数据包进一步处理。
        *   `parsePacketLayers`: 当 `dot11Layer` 为 `nil` 时 (即使 Radiotap 层存在)，函数现在也会记录错误并返回 `nil, error`，以确保不处理没有 Dot11 层的数据包。
        *   `parsePacketLayers`: `gopacket.NewPacket` 的解码选项从 `gopacket.Default` 修改为 `gopacket.Lazy`，以期提高对某些类型损坏数据包的容错性。
    *   目的是减少因 `gopacket` 解析错误导致下游指标计算失败的情况。
*   [2025-05-10 15:12:00] - **当前焦点:** 设计将 PC 端分析引擎中的 `gopacket` 解析替换为 `tshark` 的详细架构。
*   [2025-05-10 15:12:00] - **近期变更 (架构设计):**
    *   定义了基于 `tshark -T fields` 命令输出（CSV格式）的新解析流程。
    *   核心组件包括 `TSharkExecutor` (管理 `tshark` 进程), `CSVParser` (解析CSV输出), 和 `FrameProcessor` (将CSV行数据转换为 `ParsedFrameInfo`)。
    *   保留 `PhyRateCalculator` 和 `PacketInfoHandler` 接口，调整输入源。
    *   详细规划了错误处理、日志记录及与现有代码的集成点。
    *   确定了需要从 `tshark` 提取的完整字段列表。
    *   Memory Bank: [`memory-bank/developmentContext/pcAnalysisEngine.md`](memory-bank/developmentContext/pcAnalysisEngine.md:1) 已更新，增加了关于 `tshark` 替换方案的新架构章节。
    *   Memory Bank: [`memory-bank/decisionLog.md`](memory-bank/decisionLog.md:1) 已记录此架构决策。
    *   Memory Bank: [`memory-bank/activeContext.md`](memory-bank/activeContext.md:1) 本次更新。
    *   Memory Bank: [`memory-bank/progress.md`](memory-bank/progress.md:1) 即将更新。