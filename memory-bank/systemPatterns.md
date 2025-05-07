# System Patterns *Optional*

This file documents recurring patterns and standards used in the project.
It is optional, but recommended to be updated as the project evolves.

2025-05-06 23:48:00 - Initial population.
*

## Coding Patterns

*   (待定义)

## Architectural Patterns

*   [2025-05-06 23:48:00] - **分布式架构 (Distributed Architecture):** 系统由三个核心组件构成（路由器代理、PC分析引擎、Web前端），各组件独立部署和运行，通过网络通信。
    *   **理由:** 实时性要求、可扩展性、组件独立开发和部署。
*   [2025-05-06 23:48:00] - **Collector/Controller 模式:**
    *   **Collector:** 路由器端抓包代理，负责原始数据采集。
    *   **Controller:** PC端实时分析引擎，负责数据处理、分析和状态管理。
    *   **理由:** 清晰的角色划分，简化数据流和控制流。
*   [2025-05-06 23:48:00] - **微服务架构思想 (Consideration):** PC端分析引擎未来可考虑拆分为更小的微服务（数据接收、帧解析、状态管理、数据推送），以提高可维护性和独立扩展性。目前Demo阶段，PC端引擎可作为单体服务实现。

## Testing Patterns

*   (待定义)