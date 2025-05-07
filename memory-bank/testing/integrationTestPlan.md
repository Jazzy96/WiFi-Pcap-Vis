# Integration Test Plan for WiFi PCAP Visualizer

## 1. Objective
Verify the end-to-end functionality of the three core components: Router-Side Capture Agent, PC-Side Real-time Analysis Engine, and Web Frontend Visualization UI. This includes component startup, inter-component communication (gRPC and WebSocket), data capture initiation, data flow and processing, real-time data display, and capture termination.

## 2. Test Environment Prerequisites

### 2.1. Router-Side Agent
*   **IP Address:** `192.168.110.1` (Verify and use actual router IP)
*   **Application:** `router_agent_server_executable` compiled and deployed.
*   **Wireless Interface:** (e.g., `ath1`) configured in Monitor Mode.
*   **Utilities:** `tcpdump` and `iw` installed.
*   **gRPC Port:** Default `:50051` (or as configured).

### 2.2. PC-Side Analysis Engine
*   **IP Address:** `192.168.110.30` (Verify and use actual PC IP)
*   **Application:** `pc_analyzer_engine` compiled.
*   **Configuration:** `pc_analyzer/config/config.json` updated with:
    *   `router_agent_address`: Router IP and gRPC port (e.g., `192.168.110.1:50051`).
    *   `websocket_port`: Default `8080` (or as configured).
*   **WebSocket Endpoint:** `ws://localhost:8080/ws` (or as configured).

### 2.3. Web Frontend
*   **Access:** Via browser on the PC at `http://localhost:3000` (default for `npm start`).
*   **WebSocket Configuration:** `web_frontend/src/services/websocketService.ts` should point to the PC Analysis Engine's WebSocket server (e.g., `ws://localhost:8080/ws`).

## 3. Test Scenarios and Execution Steps

### Scenario 1: Successful Component Startup
*   **Step 1.1: Start Router-Side Capture Agent**
    *   **Action:** SSH to router, navigate to agent directory, run `./router_agent_server_executable`.
    *   **Expected:** Agent starts, listens on gRPC port (e.g., `:50051`). No errors.
    *   **Observe:** Router console output.
*   **Step 1.2: Start PC-Side Analysis Engine**
    *   **Action:** On PC, navigate to `pc_analyzer/`, run `./pc_analyzer_engine -config config/config.json`.
    *   **Expected:** Engine starts, WebSocket server listens (e.g., `:8080`), attempts gRPC connection. No critical errors.
    *   **Observe:** PC console output.
*   **Step 1.3: Start Web Frontend**
    *   **Action:** On PC, navigate to `web_frontend/`, run `npm start`.
    *   **Expected:** Dev server starts, browser opens `http://localhost:3000`. Page loads correctly.
    *   **Observe:** Browser console (no critical errors), page layout.

### Scenario 2: Web Frontend to PC Engine WebSocket Connection
*   **Prerequisite:** Scenario 1 successful.
*   **Step 2.1: Observe Connection**
    *   **Action:** Check Web Frontend UI and PC Engine console.
    *   **Expected:** Frontend indicates WebSocket connection success. PC Engine logs new WebSocket client.
    *   **Observe:** Frontend UI, PC console, Browser DevTools (Network > WS).

### Scenario 3: PC Engine to Router Agent gRPC Connection
*   **Prerequisite:** Scenario 1 successful.
*   **Step 3.1: Observe Connection**
    *   **Action:** Check PC Engine console and Router Agent console.
    *   **Expected:** PC Engine logs successful gRPC connection. Router Agent may log new client.
    *   **Observe:** PC console, Router console.

### Scenario 4: Initiate Capture via Web Frontend
*   **Prerequisites:** Scenarios 1, 2, 3 successful.
*   **Step 4.1: Configure Parameters (Optional)**
    *   **Action:** In Web Frontend, set interface (e.g., `ath1`), channel (e.g., `6`), bandwidth (e.g., `HT20`).
    *   **Expected:** UI elements function correctly.
*   **Step 4.2: Send "Start Capture" Command**
    *   **Action:** Click "Start Capture" button on Web Frontend.
    *   **Expected:**
        *   Frontend sends `start_capture` WebSocket message.
        *   PC Engine relays `START_CAPTURE` gRPC command.
        *   Router Agent starts `tcpdump` and streams `CaptureData` via gRPC.
    *   **Observe:** Browser DevTools (WS message), PC console (WS & gRPC logs), Router console (`tcpdump` start).

### Scenario 5: Data Parsing and Real-time Display
*   **Prerequisite:** Scenario 4 successful, active WiFi nearby.
*   **Step 5.1: Observe Data Flow and Display**
    *   **Action:** Monitor Web Frontend's BSS/STA lists.
    *   **Expected:**
        *   PC Engine parses frames, sends structured data via WebSocket.
        *   Frontend updates BSS/STA lists in real-time with correct information (SSID, BSSID, Channel, Security, associated STAs).
    *   **Observe:** Frontend UI (real-time updates), PC console (parsing logs), Browser DevTools (WS data).

### Scenario 6: Terminate Capture via Web Frontend
*   **Prerequisite:** Scenarios 4 & 5 ongoing.
*   **Step 6.1: Send "Stop Capture" Command**
    *   **Action:** Click "Stop Capture" button on Web Frontend.
    *   **Expected:**
        *   Frontend sends `stop_capture` WebSocket message.
        *   PC Engine relays `STOP_CAPTURE` gRPC command.
        *   Router Agent stops `tcpdump`.
        *   Data flow ceases; Frontend UI reflects stopped state.
    *   **Observe:** Browser DevTools (WS message), PC console (WS & gRPC logs), Router console (`tcpdump` stop), Frontend UI (data stops updating).

### Scenario 7 (Optional): Test Channel/Bandwidth Change Commands
*   **Prerequisites:** Scenarios 1, 2, 3 successful. Capture not active or stopped.
*   **Step 7.1: Send "Set Channel" Command**
    *   **Action:** On Frontend, input new channel, click "Set Channel" (if available).
    *   **Expected:** Frontend sends `set_channel` command. PC Engine relays. Router Agent (if implemented) executes `iw` command. Subsequent captures use new channel.
    *   **Observe:** Router console for `iw` command logs.
*   **Step 7.2: Send "Set Bandwidth" Command**
    *   **Action:** On Frontend, select new bandwidth, click "Set Bandwidth" (if available).
    *   **Expected:** Frontend sends `set_bandwidth` command. PC Engine relays. Router Agent (if implemented) executes `iw` command.
    *   **Observe:** Router console for `iw` command logs.

## 4. Results Collection Guidelines
For each step, record:
*   **Scenario/Step ID:**
*   **Action Performed:**
*   **Expected Result:**
*   **Actual Result:**
*   **Status (Success/Fail):**
*   **Error Messages/Logs (if any):** (Router, PC, Browser consoles)
*   **Screenshots (Optional):**
*   **Notes:**

**Results will be documented in `memory-bank/testing/integrationTestResults.md`.**