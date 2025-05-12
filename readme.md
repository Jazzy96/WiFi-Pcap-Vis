# WiFi Pcap分析项目背景文档

## 1. 项目整体背景和目标

WiFi Pcap分析项目旨在构建一个可视化的802.11空口抓包分析器演示系统，能够实时捕获和分析无线网络中的802.11帧。该系统能够从路由器实时捕获无线空口数据，传输至PC端进行解析，并最终通过Web界面展示BSS (Basic Service Set) 与STA (Station) 之间的关系及基本信息。

### 1.1 核心业务价值

* **实时监控：** 提供无线网络设备（接入点和客户端）的实时监控能力，帮助管理员了解网络状态。
* **问题诊断：** 协助网络工程师快速定位和解决Wi-Fi网络中的问题。
* **性能分析：** 通过对信道占用率、吞吐量等指标的监控，评估网络性能。
* **教育价值：** 作为802.11协议实现细节的可视化工具，有助于学习和教学。

### 1.2 关键功能特性

* **实时数据捕获：** 能够实时捕获802.11无线帧（包含Radiotap头部信息）。
* **多类型帧解析：** 支持解析管理帧、控制帧和数据帧等多种802.11帧类型。
* **BSS与STA关系追踪：** 维护无线网络中基站和连接客户端之间的关联状态。
* **性能指标分析：** 计算和展示信道占用率、吞吐量等关键性能指标。
* **用户友好界面：** 提供直观的可视化界面，以图形化方式展示网络结构和性能数据。

## 2. 系统架构概述

本项目采用分布式架构，由三个主要组件组成，每个组件负责特定功能，通过定义良好的接口进行通信。

### 2.1 整体架构

系统由以下三个核心组件构成：

1. **路由器端抓包代理 (Router-Side Capture Agent)**：
   - 在路由器上运行的轻量级程序
   - 负责配置无线网卡为Monitor模式并捕获原始802.11帧
   - 将捕获的数据通过gRPC流式传输到PC端分析引擎

2. **PC端实时分析引擎 (PC-Side Real-time Analysis Engine)**：
   - 在用户PC上运行（支持Windows/macOS）
   - 接收来自路由器的数据流，高速解析802.11帧
   - 维护BSS与STA的关联状态，提取关键信息
   - 将处理后的结构化数据通过WebSocket推送给Web前端

3. **Web前端可视化界面 (Web Frontend Visualization UI)**：
   - 在浏览器中运行，通过WebSocket接收数据
   - 以用户友好的方式展示BSS、STA及其关联关系和基本信息
   - 支持实时更新和用户控制指令发送

### 2.2 数据流与控制流

* **数据流**：
  ```
  路由器 (原始帧) -> gRPC -> PC引擎 (解析、状态管理) -> WebSocket -> Web前端 (可视化)
  ```

* **控制流**：
  ```
  Web前端 (用户指令) -> WebSocket -> PC引擎 -> gRPC -> 路由器代理 (执行操作)
  ```

### 2.3 技术选型

* **gRPC (路由器 -> PC)**：
  - 基于HTTP/2，支持双向流，适合实时传输原始帧数据
  - 跨语言支持通过Protobuf定义接口，便于不同语言实现的组件间通信
  - 强类型接口定义，减少集成错误
  - 比自定义TCP协议开发成本低，且有成熟的库支持

* **WebSocket (PC -> Web)**：
  - 浏览器原生支持，适合将分析结果实时推送到前端
  - 支持双向通信，允许前端发送控制指令回PC端
  - 现代浏览器普遍支持
## 3. 各个子模块的详细实现

### 3.1 路由器端抓包代理

路由器端抓包代理是在目标路由器上运行的服务程序，负责捕获802.11帧并将其传输到PC端分析引擎。

#### 3.1.1 技术栈

* **语言**：Go (Golang)
* **gRPC库**：`google.golang.org/grpc`
* **Protobuf库**：`google.golang.org/protobuf`
* **系统工具**：`tcpdump`和`iw`命令

#### 3.1.2 核心组件

* **gRPC服务定义**（`capture_agent.proto`）：
  ```protobuf
  syntax = "proto3";
  
  package router_agent;
  
  option go_package = ".;main";
  
  // 控制指令类型
  enum ControlCommandType {
    UNKNOWN_COMMAND = 0;
    START_CAPTURE = 1;
    STOP_CAPTURE = 2;
    SET_CHANNEL = 3;
    SET_BANDWIDTH = 4;
  }
  
  // 控制指令消息
  message ControlRequest {
    ControlCommandType command_type = 1;
    string interface_name = 2; // e.g., "ath1"
    int32 channel = 3;         // e.g., 1, 6, 11, 36, 149
    string bandwidth = 4;      // e.g., "HT20", "HT40", "VHT80"
    string bpf_filter = 5;     // BPF filter string for tcpdump
  }
  
  // 控制指令响应
  message ControlResponse {
    bool success = 1;
    string message = 2;
  }
  
  // 抓包数据消息
  message CaptureData {
    bytes frame = 1; // 原始帧数据 (包含Radiotap)
  }
  
  // gRPC 服务定义
  service CaptureAgent {
    // PC端发送控制指令给路由器代理
    rpc SendControlCommand(ControlRequest) returns (ControlResponse);
  
    // 路由器代理向PC端流式发送抓包数据
    rpc StreamPackets(ControlRequest) returns (stream CaptureData);
  }
  ```

* **服务实现**（`main.go`）：
  - `server`结构体：实现`CaptureAgentServer`接口，管理抓包状态和`tcpdump`进程
  - `SendControlCommand`方法：处理来自PC端的控制命令，如启动/停止抓包
  - `StreamPackets`方法：将抓包数据流式传输到PC端分析引擎
  - `setInterfaceParams`方法：使用`iw`命令设置无线接口参数（如信道、带宽）

#### 3.1.3 关键实现细节

* **抓包实现**：
  - 使用`tcpdump -i <interface> -U -w - <bpf_filter>`命令捕获数据
  - `-U`标志确保每个包被及时刷新到输出
  - `-w -`标志将原始pcap数据写入标准输出
  - 数据通过管道读取并发送到gRPC流

* **错误处理**：
  - 优雅地处理`tcpdump`进程的启动、监控和终止
  - 使用`SIGINT`信号优雅停止`tcpdump`进程，必要时使用`SIGKILL`
  - 检测客户端断开连接（通过`stream.Context().Err()`）

#### 3.1.4 编译与部署

* **交叉编译**：
  ```bash
  GOOS=linux GOARCH=arm64 go build -o router_agent_arm64 .
  ```

* **部署到路由器**：
  ```bash
  scp router_agent_arm64 USER@ROUTER_IP:/path/on/router/
  chmod +x /path/on/router/router_agent_arm64
  ```

* **运行**：
  ```bash
  /path/on/router/router_agent_arm64
  # 或指定端口
  CAPTURE_AGENT_PORT=60051 /path/on/router/router_agent_arm64
  ```

### 3.2 PC端实时分析引擎

PC端实时分析引擎是项目的核心组件，负责接收和解析802.11帧数据，维护网络状态，并通过Wails框架向前端提供实时更新。

#### 3.2.1 技术栈

* **语言**：Go (Golang)
* **应用框架**：Wails (v2.5+)，提供Go和前端JavaScript集成
* **帧解析**：使用`tshark`命令行工具解析PCAP数据
* **gRPC客户端**：`google.golang.org/grpc`
* **日志管理**：`github.com/rs/zerolog`
* **并发控制**：使用Go原生sync包和context包

#### 3.2.2 核心模块

* **配置管理**（`config/config.go` & `config/config.json`）：
  - 加载和管理应用配置
  - 提供默认配置值和配置合并
  - 配置项包括gRPC服务器地址、日志级别、tshark路径、最小BSS创建信号强度等

* **gRPC客户端**（`grpc_client/client.go`）：
  - 连接到路由器端抓包代理
  - 发送控制命令（开始/停止捕获、设置信道和带宽等）
  - 接收PCAP数据流并通过io.Pipe转换为io.Reader供解析器使用
  - 实现错误处理与上下文取消

* **帧解析器**（`frame_parser/parser.go`）：
  - `TSharkExecutor`：管理tshark子进程和数据流
  - `CSVParser`：解析tshark输出的CSV格式数据
  - `FrameProcessor`：将CSV行转换为结构化的ParsedFrameInfo对象
  - 支持HT/VHT/HE能力解析，提取物理层和MAC层参数
  - 实现PHY速率计算与帧传输时间估算

* **状态管理器**（`state_manager/manager.go` & `models.go`）：
  - 维护BSS和STA的状态、关系和历史记录
  - 实现BSS/STA确认机制，使用先判定后确认的两阶段策略
  - 基于管理帧和数据帧处理BSS/STA关联关系
  - 计算性能指标（信道占用率、吞吐量、上下行速率）
  - 支持历史指标数据存储和定期清理过期条目
  - 生成状态快照并通过Wails事件推送到前端

* **日志管理**（`logger/logger.go`）：
  - 基于zerolog实现结构化日志
  - 支持多目标日志输出（控制台和文件）
  - 可配置的日志级别和格式化

* **主应用逻辑**（`app.go` & `main.go`）：
  - 作为Wails应用程序入口
  - 初始化和协调各模块
  - 提供暴露给前端的方法（StartCapture, StopCapture等）
  - 通过Wails事件机制将状态和错误信息传递给前端

#### 3.2.3 帧解析流程

1. **数据接收**：
   - gRPC客户端通过StreamPackets RPC调用从路由器接收PCAP数据流
   - 数据通过io.Pipe传递给tshark解析器

2. **tshark处理**：
   - 使用命令格式：`tshark -r - -T fields -E header=y -E separator=, -E quote=d -E occurrence=a -e <field1> -e <field2> ...`
   - `-r -`参数使tshark从标准输入读取PCAP数据
   - 指定约70个关键802.11字段，包括帧头信息、管理帧元素、QoS参数等
   - 输出结构化的CSV格式数据

3. **CSV解析**：
   - 首行解析为列名映射（HeaderMap），便于后续按名称访问字段
   - 逐行解析CSV数据为字段名到值的映射（map[string]string）
   - 处理多值字段和类型转换（字符串、整数、MAC地址等）

4. **帧处理**：
   - FrameProcessor将字段映射转换为结构化的ParsedFrameInfo对象
   - 解析帧类型、MAC地址、信号强度、信道、带宽等基本信息
   - 提取HT/VHT/HE能力信息，解析安全参数
   - 计算PHY速率和传输时间估算
   - 处理特殊字段如SSID（可能为隐藏或需要解码）

5. **状态更新**：
   - 处理后的帧信息通过回调传递给StateManager的ProcessParsedFrame方法
   - 根据帧类型（管理、控制或数据）和子类型进行不同处理
   - 更新BSS和STA状态，处理关联关系变更
   - 累积字节计数和传输时间用于性能指标计算
   - 定期（默认1秒）计算和更新性能指标

#### 3.2.4 关键数据结构

* **ParsedFrameInfo**（`frame_parser/parser.go`）：
  ```go
  type ParsedFrameInfo struct {
      Timestamp              time.Time          // 帧捕获时间戳
      FrameType              string             // 如"MgmtBeacon", "Data", "QoSData"等
      WlanFcType             uint8              // WLAN帧类型 (0=管理, 1=控制, 2=数据)
      WlanFcSubtype          uint8              // WLAN帧子类型
      BSSID                  net.HardwareAddr   // BSS标识符(MAC)
      SA                     net.HardwareAddr   // 源地址
      DA                     net.HardwareAddr   // 目标地址
      RA                     net.HardwareAddr   // 接收地址
      TA                     net.HardwareAddr   // 传输地址
      Channel                int                // 信道号
      Frequency              int                // 频率(MHz)
      SignalStrength         int                // 信号强度(dBm)
      Bandwidth              string             // 带宽(如"20MHz", "40MHz", "80MHz")
      SSID                   string             // 网络名称
      Security               string             // 安全类型(如"Open", "WPA2/WPA3")
      ParsedHTCaps           *HTCapabilityInfo  // HT能力信息
      ParsedVHTCaps          *VHTCapabilityInfo // VHT能力信息
      ParsedHECaps           *HECapabilityInfo  // HE能力信息(Wi-Fi 6)
      FrameLength            int                // 原始帧长度
      PHYRateMbps            float64            // 估算的物理层速率(Mbps)
      BitRate                float64            // 站点比特率
      TransportPayloadLength int                // 传输层负载长度
      MACDurationID          uint16             // MAC层持续时间/ID字段
      RetryFlag              bool               // 重传标志位
      RawFields              map[string]string  // 原始tshark字段
  }
  ```

* **BSSInfo**（`state_manager/models.go`）：
  ```go
  type BSSInfo struct {
      BSSID                      string             // MAC地址
      SSID                       string             // 网络名称
      Channel                    int                // 工作信道
      Bandwidth                  string             // 带宽
      Security                   string             // 安全类型
      SignalStrength             int                // 信号强度(dBm)
      LastSeen                   int64              // 上次检测时间(毫秒时间戳)
      HTCapabilities             *HTCapabilities    // HT能力信息
      VHTCapabilities            *VHTCapabilities   // VHT能力信息
      HECapabilities             *HECapabilities    // HE能力信息
      AssociatedSTAs             map[string]*STAInfo // 关联的站点
      
      // 性能指标
      ChannelUtilization         float64            // 信道利用率(%)
      Throughput                 int64              // 总吞吐量(bps)
      HistoricalChannelUtilization []float64       // 历史信道利用率
      HistoricalThroughput       []int64           // 历史吞吐量
      
      // 内部指标计算字段
      lastCalcTime               time.Time         // 上次计算时间
      totalTxBytes               int64             // 计算窗口内总传输字节数
      AccumulatedNavMicroseconds uint64            // 累积NAV持续时间(微秒)
  }
  ```

* **STAInfo**（`state_manager/models.go`）：
  ```go
  type STAInfo struct {
      MACAddress                 string            // MAC地址
      AssociatedBSSID            string            // 关联的BSS
      SignalStrength             int               // 信号强度(dBm)
      LastSeen                   int64             // 上次检测时间(毫秒时间戳)
      HTCapabilities             *HTCapabilities   // HT能力信息
      VHTCapabilities            *VHTCapabilities  // VHT能力信息
      HECapabilities             *HECapabilities   // HE能力信息
      
      // 性能指标
      ChannelUtilization         float64           // 信道利用率(%)
      UplinkThroughput           int64             // 上行吞吐量(bps)
      DownlinkThroughput         int64             // 下行吞吐量(bps)
      HistoricalChannelUtilization []float64       // 历史信道利用率
      HistoricalUplinkThroughput []int64           // 历史上行吞吐量
      HistoricalDownlinkThroughput []int64         // 历史下行吞吐量
      
      // 内部字段
      lastCalcTime               time.Time         // 上次计算时间
      totalAirtime               time.Duration     // 计算窗口内总传输时间
      totalUplinkBytes           int64             // 计算窗口内上行字节数
      totalDownlinkBytes         int64             // 计算窗口内下行字节数
      BitRate                    float64           // 当前比特率(Mbps)
  }
  ```

* **Snapshot**（`state_manager/models.go`）：
  ```go
  type Snapshot struct {
      BSSs []*BSSInfo      // BSS列表(深拷贝)
      STAs []*STAInfo      // STA列表(深拷贝)
  }
  ```

### 3.3 Web前端可视化界面

Web前端提供用户界面，展示无线网络结构和性能数据，并允许用户发送控制命令。

#### 3.3.1 技术栈

* **框架**：React（TypeScript）
* **状态管理**：React的useReducer结合Context API
* **通信**：Wails框架绑定（EventsOn API和直接Go函数调用）
* **样式**：CSS Modules & 全局CSS变量
* **图表**：Recharts库
* **UI组件**：自定义组件库（Button, Card, Input等）

#### 3.3.2 项目结构

```
desktop_app/WifiPcapAnalyzer/frontend/
├── src/
│   ├── components/
│   │   ├── common/              // 可复用的原子UI组件
│   │   │   ├── Button/          // 按钮组件
│   │   │   ├── Card/            // 卡片容器组件
│   │   │   ├── Input/           // 输入控件组件
│   │   │   ├── Table/           // 表格组件
│   │   │   ├── Tabs/            // 标签页组件
│   │   │   └── Icon/            // 图标组件
│   │   ├── BssList/             // BSS列表组件
│   │   ├── StaList/             // STA列表组件
│   │   ├── ControlPanel/        // 控制面板组件
│   │   └── PerformanceDetailPanel/ // 性能详情面板
│   ├── contexts/
│   │   └── DataContext.tsx      // 全局状态管理
│   ├── types/
│   │   └── data.ts              // TypeScript类型定义
│   ├── wailsjs/                 // Wails自动生成的绑定代码
│   │   ├── go/                  // Go后端绑定
│   │   └── runtime/             // Wails运行时
│   ├── App.tsx                  // 主应用组件
│   ├── App.css                  // 全局样式
│   └── index.tsx                // 入口文件
```

#### 3.3.3 核心组件

* **DataContext**（`contexts/DataContext.tsx`）：
  - 使用React的useReducer和Context API管理全局状态
  - 通过Wails的EventsOn API处理后端事件
  - 维护BSS列表、STA列表、捕获状态、UI状态等
  - 为组件提供状态读取(useAppState)和更新(useAppDispatch)入口

* **BssList**（`components/BssList/BssList.tsx`）：
  - 展示所有检测到的BSS
  - 使用Card组件以卡片形式展示每个BSS
  - 显示SSID、BSSID、信道、信号强度、关联的STA数量和性能指标
  - 支持卡片展开查看更多详情
  - 支持选择BSS以查看关联的STA和性能详情
  - 按信号强度和STA数量排序

* **StaList**（`components/StaList/StaList.tsx`）：
  - 展示与选定BSS关联的STA
  - 使用Card组件以卡片形式展示每个STA
  - 显示MAC地址、信号强度、利用率、吞吐量和比特率
  - 支持选择STA查看详细性能指标
  - 根据选择的BSS动态更新显示内容

* **ControlPanel**（`components/ControlPanel/ControlPanel.tsx`）：
  - 提供用户控制界面
  - 支持面板收起/展开
  - 允许选择5GHz信道和带宽
  - 使用自定义Input组件（select类型）提供下拉选择
  - 包含启动/停止抓包的按钮
  - 直接调用Wails绑定的Go函数(StartCapture, StopCapture)

* **PerformanceDetailPanel**（`components/PerformanceDetailPanel/PerformanceDetailPanel.tsx`）：
  - 展示选中的BSS或STA的详细性能指标
  - 根据选择的目标类型(BSS/STA)显示不同指标
  - 使用Recharts库创建可视化图表
  - 显示历史性能数据的线图
  - BSS指标：信道利用率、总吞吐量、信号强度、关联STA数等
  - STA指标：信号强度、上行/下行吞吐量、发送/接收比特率等

#### 3.3.4 数据结构

* **BSS** (`types/data.ts`)：
  ```typescript
  interface BSS {
    bssid: string;               // MAC地址
    ssid: string;                // 网络名称
    channel: number;             // 工作信道
    bandwidth: string;           // 带宽
    security: string;            // 安全类型
    signal_strength: number | null; // 信号强度
    last_seen: string;           // 上次检测时间
    associated_stas: { [mac: string]: STA }; // 关联的STA
    ht_capabilities?: HTCabilities;  // 高吞吐量能力
    vht_capabilities?: VHTCabilities; // 非常高吞吐量能力
    // 性能指标
    channel_utilization_percent: number; // 信道利用率
    total_throughput_mbps: number;       // 总吞吐量
    historical_channel_utilization?: { timestamp: number; value: number }[]; // 历史信道利用率
    historical_total_throughput?: { timestamp: number; value: number }[];    // 历史吞吐量
  }
  ```

* **STA** (`types/data.ts`)：
  ```typescript
  interface STA {
    mac_address: string;         // MAC地址
    associated_bssid?: string;   // 关联的BSS
    signal_strength: number | null; // 信号强度
    last_seen: string;           // 上次检测时间
    ht_capabilities?: HTCabilities;  // 高吞吐量能力
    vht_capabilities?: VHTCabilities; // 非常高吞吐量能力
    // 性能指标
    rx_bytes: number;            // 接收字节数
    tx_bytes: number;            // 发送字节数
    rx_packets: number;          // 接收包数
    tx_packets: number;          // 发送包数
    rx_retries: number;          // 接收重试次数
    tx_retries: number;          // 发送重试次数
    rx_bitrate_mbps: number;     // 接收比特率
    tx_bitrate_mbps: number;     // 发送比特率
    throughput_ul_mbps: number;  // 上行吞吐量
    throughput_dl_mbps: number;  // 下行吞吐量
    historical_throughput_ul?: { timestamp: number; value: number }[]; // 历史上行吞吐量
    historical_throughput_dl?: { timestamp: number; value: number }[]; // 历史下行吞吐量
  }
  ```

#### 3.3.5 UI/UX设计规范

* **整体风格**：
  - 高端、简约、专业
  - 自然、稳定，适合企业Wi-Fi测试客户
  - 基于卡片的布局，提供层次感和组织结构

* **配色方案**：
  - 石墨灰（#1F242B）：主背景、暗元素
  - 薄雾白（#F5F7F9）：主内容区背景、亮元素
  - 科技蓝（#1E90FF）：交互元素、强调部分、关键图表颜色
  - 使用CSS变量管理颜色，便于主题切换

* **排版与布局**：
  - 响应式网格布局，支持面板收缩
  - 主要字体：SF Pro（备用：Helvetica Neue, Arial等）
  - 圆角：8px
  - 轻质阴影
  - 符合WCAG AA对比度标准

#### 3.3.6 前后端交互

* **数据接收**：
  - 通过Wails的EventsOn API订阅后端事件
  - 主要事件包括'state_snapshot'、'capture_status'和'error'
  - 收到事件后通过dispatch更新状态

* **发送控制命令**：
  - 直接调用Wails绑定的Go函数
  - 主要命令包括StartCapture和StopCapture
  - 支持指定接口名称、信道、带宽等参数
## 4. 各模块之间的交互方式

### 4.1 路由器代理与PC引擎之间的交互

#### 4.1.1 gRPC通信流程

1. **建立连接**：
   - PC引擎使用gRPC客户端连接到路由器代理
   - 连接建立后，PC引擎可以发送控制命令和请求数据流

2. **发送控制命令**：
   - PC引擎调用`SendControlCommand` RPC方法
   - 携带`ControlRequest`消息，指定命令类型和参数
   - 路由器代理执行相应操作，并返回`ControlResponse`

3. **数据流传输**：
   - PC引擎调用`StreamPackets` RPC方法，请求数据流
   - 路由器代理启动抓包，将每个捕获的帧封装为`CaptureData`消息
   - 通过gRPC流将数据发送到PC引擎

#### 4.1.2 控制命令示例

* **启动抓包**：
  ```protobuf
  // 请求
  {
    command_type: START_CAPTURE,
    interface_name: "ath1",
    channel: 149,
    bandwidth: "HT40",
    bpf_filter: "type mgt or type data"
  }
  
  // 响应
  {
    success: true,
    message: "Capture started on interface ath1"
  }
  ```

* **停止抓包**：
  ```protobuf
  // 请求
  {
    command_type: STOP_CAPTURE
  }
  
  // 响应
  {
    success: true,
    message: "Capture stopped"
  }
  ```

### 4.2 PC引擎与Web前端之间的交互

#### 4.2.1 WebSocket通信流程

1. **建立连接**：
   - Web前端通过Wails绑定与PC引擎建立WebSocket连接
   - 连接建立后，前端可以接收数据和发送控制命令

2. **数据推送**：
   - PC引擎定期生成BSS和STA状态快照
   - 通过WebSocket将快照推送到前端
   - 前端更新UI展示最新数据

3. **控制命令**：
   - 前端通过WebSocket发送控制命令
   - PC引擎处理命令并转发到路由器代理（如需要）
   - PC引擎返回命令执行结果

#### 4.2.2 WebSocket消息示例

* **状态快照**（PC引擎 -> Web前端）：
  ```json
  {
    "type": "snapshot",
    "data": {
      "bsss": [
        {
          "bssid": "AA:BB:CC:DD:EE:FF",
          "ssid": "MyWiFi",
          "channel": 149,
          "bandwidth": "40MHz",
          "security": "WPA2-PSK",
          "signal_strength": -50,
          "channel_utilization": 25.5,
          "throughput": 54.3,
          "associated_stas": {
            "11:22:33:44:55:66": {
              "mac": "11:22:33:44:55:66",
              "signal_strength": -55,
              "last_seen": 1678886400000,
              "throughput": 12.4,
              "bitrate": 86.7
            }
          }
        }
      ]
    }
  }
  ```

* **控制命令**（Web前端 -> PC引擎）：
  ```json
  {
    "action": "start_capture",
    "payload": {
      "interface": "ath1",
      "channel": 149,
      "bandwidth": "40"
    }
  }
  ```

## 5. 关键数据流程图

### 5.1 系统初始化流程

```
┌─────────────┐     ┌──────────────┐     ┌────────────────┐
│  Web前端    │     │   PC引擎     │     │  路由器代理    │
└─────┬───────┘     └───────┬──────┘     └────────┬───────┘
      │                     │                     │
      │                     │     初始化配置      │
      │                     ├─────────────────────┤
      │                     │                     │
      │                     │   启动日志服务      │
      │                     ├─────────────────────┤
      │                     │                     │
      │     初始化UI        │                     │
      ├─────────────────────┤                     │
      │                     │                     │
      │    建立WebSocket    │                     │
      ├────────────────────►│                     │
      │                     │                     │
      │                     │    建立gRPC连接     │
      │                     ├────────────────────►│
      │                     │                     │
      │                     │ 启动状态管理服务    │
      │                     ├─────────────────────┤
      │                     │                     │
      │                     │ 启动指标计算服务    │
      │                     ├─────────────────────┤
      │                     │                     │
      │ 初始化完成通知      │                     │
      │◄────────────────────┤                     │
      │                     │                     │
```

### 5.2 帧捕获与处理流程

```
┌─────────────┐     ┌──────────────┐     ┌────────────────┐
│  Web前端    │     │   PC引擎     │     │  路由器代理    │
└─────┬───────┘     └───────┬──────┘     └────────┬───────┘
      │                     │                     │
      │                     │                     │
      │  发送start_capture  │                     │
      ├────────────────────►│                     │
      │                     │                     │
      │                     │ SendControlCommand  │
      │                     ├────────────────────►│
      │                     │                     │
      │                     │                     │ 启动tcpdump
      │                     │                     ├─────────────
      │                     │                     │
      │                     │  StreamPackets      │
      │                     ├────────────────────►│
      │                     │                     │
      │                     │                     │ 捕获802.11帧
      │                     │                     ├─────────────
      │                     │                     │
      │                     │ 帧数据(CaptureData) │
      │                     │◄────────────────────┤
      │                     │                     │
      │                     │ 使用tshark解析帧    │
      │                     ├─────────────────────┤
      │                     │                     │
      │                     │ 更新BSS/STA状态     │
      │                     ├─────────────────────┤
      │                     │                     │
      │                     │ 计算性能指标        │
      │                     ├─────────────────────┤
      │                     │                     │
      │    状态更新通知     │                     │
      │◄────────────────────┤                     │
      │                     │                     │
      │     更新UI展示      │                     │
      ├─────────────────────┤                     │
      │                     │                     │
```

### 5.3 控制命令流程

```
┌─────────────┐     ┌──────────────┐     ┌────────────────┐
│  Web前端    │     │   PC引擎     │     │  路由器代理    │
└─────┬───────┘     └───────┬──────┘     └────────┬───────┘
      │                     │                     │
      │  用户点击控制按钮   │                     │
      ├─────────────────────┤                     │
      │                     │                     │
      │  发送控制命令       │                     │
      ├────────────────────►│                     │
      │                     │                     │
      │                     │ 验证命令格式和参数  │
      │                     ├─────────────────────┤
      │                     │                     │
      │                     │ 转换为gRPC请求      │
      │                     ├─────────────────────┤
      │                     │                     │
      │                     │ SendControlCommand  │
      │                     ├────────────────────►│
      │                     │                     │
      │                     │                     │ 执行请求操作
      │                     │                     ├─────────────
      │                     │                     │
      │                     │ ControlResponse     │
      │                     │◄────────────────────┤
      │                     │                     │
      │     命令结果通知    │                     │
      │◄────────────────────┤                     │
      │                     │                     │
      │  更新UI反映结果     │                     │
      ├─────────────────────┤                     │
      │                     │                     │
```

## 6. 开发和调试指南

### 6.1 环境搭建

#### 6.1.1 路由器环境

* **硬件要求**：
  - 支持Monitor模式的WiFi网卡
  - 足够的CPU和内存资源运行抓包程序

* **软件要求**：
  - Linux操作系统（如OpenWrt）
  - 安装 `tcpdump` 工具
  - 安装 `iw` 工具
  - 正确配置的无线接口（如"ath1"）

* **设置Monitor模式**：
  ```bash
  iw dev wlan0 interface add mon0 type monitor
  ip link set mon0 up
  ```

#### 6.1.2 开发环境

* **PC端开发环境**：
  - Go 1.20+
  - Wails框架 (v2.5+)
  - Node.js (v18+)与npm/yarn
  - `tshark` 命令行工具 (Wireshark CLI)
  - Protobuf编译器 (`protoc`)
  - Go Protobuf插件 (`protoc-gen-go`, `protoc-gen-go-grpc`)

* **开发工具**：
  - Visual Studio Code或GoLand
  - Wireshark (用于验证解析结果)
  - Chrome开发者工具 (用于Web前端调试)

### 6.2 构建与运行

#### 6.2.1 路由器代理

* **生成Protobuf代码**：
  ```bash
  cd router_agent
  protoc --go_out=. --go_opt=paths=source_relative \
         --go-grpc_out=. --go-grpc_opt=paths=source_relative \
         capture_agent.proto
  ```

* **交叉编译**：
  ```bash
  cd router_agent
  GOOS=linux GOARCH=arm64 go build -o router_agent_arm64 .
  ```

* **部署与运行**：
  ```bash
  scp router_agent_arm64 USER@ROUTER_IP:/path/on/router/
  ssh USER@ROUTER_IP "chmod +x /path/on/router/router_agent_arm64 && /path/on/router/router_agent_arm64"
  ```

#### 6.2.2 PC端分析引擎与Web前端

* **安装依赖**：
  ```bash
  cd desktop_app/WifiPcapAnalyzer
  go mod tidy
  cd frontend
  npm install
  ```

* **开发模式运行**：
  ```bash
  cd desktop_app/WifiPcapAnalyzer
  wails dev
  ```

* **构建生产版本**：
  ```bash
  cd desktop_app/WifiPcapAnalyzer
  wails build
  ```

### 6.3 调试技巧

#### 6.3.1 路由器代理调试

* **检查网卡模式**：
  ```bash
  iw dev
  ```

* **手动测试tcpdump**：
  ```bash
  tcpdump -i ath1 -U -w - | less
  ```

* **查看gRPC服务日志**：
  ```bash
  /path/on/router/router_agent_arm64 2>&1 | tee router_agent.log
  ```

#### 6.3.2 PC端分析引擎调试

* **PCAP文件测试**：
  使用配置文件中的`test_pcap_file`选项，允许从文件加载PCAP数据，而不是实时抓包。

* **查看tshark输出**：
  ```bash
  tshark -r capture.pcap -T fields -E header=y -E separator=, -E quote=d -e frame.time_epoch -e radiotap.channel.freq -e wlan.sa | head
  ```

* **日志级别调整**：
  修改配置文件中的`log_level`参数为"debug"以获取详细日志。

#### 6.3.3 Web前端调试

* **React开发者工具**：
  使用Chrome扩展检查组件层次和状态。

* **WebSocket流量检查**：
  通过浏览器开发者工具的Network面板监视WebSocket消息。

* **模拟数据**：
  创建模拟数据源以测试UI渲染，避免依赖后端服务。

### 6.4 常见问题与解决方案

#### 6.4.1 路由器代理问题

* **tcpdump启动失败**：
  - 检查无线接口是否存在
  - 确保有足够权限运行tcpdump
  - 验证无线网卡是否支持Monitor模式

* **"Unimplemented" gRPC错误**：
  - 确保路由器代理和PC引擎使用相同的Protobuf定义
  - 检查service和package名称是否一致

#### 6.4.2 PC端分析引擎问题

* **tshark字段解析错误**：
  - 验证字段名是否正确（使用`tshark -G fields`查看所有字段）
  - 确保使用正确的tshark版本

* **性能指标显示为"N/A"**：
  - 检查是否正确初始化`lastCalcTime`
  - 确保帧解析提取了必要的字段（如Duration/ID, 传输层负载长度）

#### 6.4.3 Web前端问题

* **WebSocket连接失败**：
  - 确认PC引擎正在运行并监听WebSocket端口
  - 检查WebSocket URL是否正确

* **UI渲染问题**：
  - 检查数据结构是否与前端期望的类型定义匹配
  - 验证CSS样式是否正确应用

### 6.5 性能优化建议

#### 6.5.1 路由器端优化

* **BPF过滤器**：
  使用精确的BPF过滤器减少捕获的帧数量，如：
  ```
  type mgt subtype beacon or type mgt subtype probe-resp or type data
  ```

* **缓冲区调整**：
  适当增加tcpdump的缓冲区大小以防止丢包。

#### 6.5.2 PC端优化

* **批处理更新**：
  累积多个帧的处理结果，进行批量状态更新。

* **过滤重复数据**：
  对于频繁出现的信息（如Beacon帧），实施智能采样。

#### 6.5.3 Web前端优化

* **虚拟化列表**：
  当BSS或STA数量很大时，使用virtualized list (如`react-window`)。

* **防抖动更新**：
  实施更新防抖以避免过于频繁的UI重新渲染。
  - 使用Recharts库展示性能指标图表
  - 展示信道占用率和吞吐量等历史数据

#### 3.3.4 UI/UX设计规范

* **整体风格**：
  - 高端、简约、专业
  - 自然、稳定，适合企业Wi-Fi测试客户

* **配色方案**：
  - 石墨灰（#1F242B）：主背景、暗元素
  - 薄雾白（#F5F7F9）：主内容区背景、亮元素
  - 科技蓝（#1E90FF）：交互元素、强调部分、关键图表颜色

* **排版与布局**：
  - 基于网格的布局
  - 主要字体：SF Pro（备用：Helvetica Neue, Arial等）
  - 圆角：8px
  - 轻质阴影
  - 符合WCAG AA对比度标准

#### 3.3.5 数据交互

* **从PC引擎接收数据**：
  ```typescript
  // 数据结构示例
  interface WebSocketData {
    type: string;  // 如 "snapshot"
    data: {
      bsss: BSS[];
      stas: STA[];
    }
  }
  ```

* **向PC引擎发送控制命令**：
  ```typescript
  // 命令结构示例
  interface ControlCommand {
    action: string;  // 如 "start_capture"
    payload?: {
      interface?: string;
      channel?: number;
      bandwidth?: string;
    }
  }
  ```