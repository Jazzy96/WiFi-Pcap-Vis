# Product Context

This file provides a high-level overview of the project and the expected product that will be created. Initially it is based upon projectBrief.md (if provided) and all other available project-related information in the working directory. This file is intended to be updated as the project evolves, and should be used to inform all other modes of the project's goals and context.

2025-05-06 23:48:00 - Initial population based on user request and requirements document.

*

## Project Goal

*   构建一个可视化802.11空口抓包分析器Demo，能够从路由器实时捕获无线空口数据，传输至PC端进行解析，并最终通过Web界面展示BSS (Basic Service Set) 与STA (Station) 关系及基本信息。

## Key Features

*   **路由器端抓包代理:**
    *   在路由器上运行，配置无线网卡为Monitor模式。
    *   根据指令启动/停止指定信道和带宽的802.11空口抓包。
    *   将捕获到的原始帧数据（带有Radiotap头部）实时传输给PC端分析引擎。
*   **PC端实时分析引擎:**
    *   在用户PC上运行（支持Windows/macOS）。
    *   接收来自路由器的数据流，高速解析802.11帧。
    *   维护BSS与STA的关联状态，提取关键信息（SSID, BSSID, STA MAC, 信道，带宽，安全类型等）。
    *   将处理后的结构化数据通过WebSocket推送给Web前端。
*   **Web前端可视化界面:**
    *   在浏览器中运行，通过WebSocket接收数据。
    *   以用户友好的方式展示BSS、STA及其关联关系和基本信息。
    *   支持实时更新和用户控制指令发送。

## Overall Architecture

*   采用分布式架构，主要包含三个核心组件：
    1.  **路由器端抓包代理 (Router-Side Capture Agent)**
    2.  **PC端实时分析引擎 (PC-Side Real-time Analysis Engine)**
    3.  **Web前端可视化界面 (Web Frontend Visualization UI)**
*   **数据流:** 路由器 (原始帧) -> gRPC -> PC引擎 (解析、状态管理) -> WebSocket -> Web前端 (可视化)。
*   **控制流:** Web前端 (用户指令) -> WebSocket -> PC引擎 -> gRPC -> 路由器代理 (执行操作)。
## UI/UX Design Specifications (as of 2025-05-08)

This section outlines the design guidelines for the UI/UX redesign of the "Enterprise Wi-Fi Capture Analysis Software".

**Overall Style:**
*   **Target:** High-end, minimalist, professional.
*   **Feel:** Natural, stable, not overly flashy, suitable for enterprise Wi-Fi testing clients.

**Color Palette:**
*   **Base:**
    *   Graphite Gray: `#1F242B` (Primary background, dark elements)
    *   Misty White: `#F5F7F9` (Primary content area background, light elements)
*   **Accent:**
    *   Tech Blue: `#1E90FF` (Interactive elements, highlights, key chart color)

**Typography & Layout:**
*   **Layout:** Grid-based.
*   **Font:**
    *   Primary: "SF Pro"
    *   Fallback: "Helvetica Neue", Arial, sans-serif (or "Roboto", "Open Sans", "Lato")
*   **Rounded Corners:** 8px.
*   **Shadows:** Lightweight shadows.
*   **Contrast:** Adhere to WCAG AA standards.

**Key Design Principles:**
*   Clarity and ease of use for technical users.
*   Professional aesthetic suitable for enterprise software.
*   Consistent application of styles across all UI elements.
*   Responsive handling of information display within the desktop application window.

---
*(Existing content follows)*