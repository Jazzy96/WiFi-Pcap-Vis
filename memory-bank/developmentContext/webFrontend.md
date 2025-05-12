# Web Frontend Development Context

This document outlines the development details for the Web Frontend Visualization UI component of the WiFi PCAP Visualizer project.

**Last Updated: 2025-05-08 04:04:00** - Reflects new UI/UX redesign architecture.

## 1. Overview

The Web Frontend is responsible for:
*   Connecting to the PC-side Real-time Analysis Engine via WebSocket.
*   Receiving and displaying BSS (Basic Service Set) and STA (Station) information in real-time, adhering to the new UI/UX design specifications.
*   Allowing users to send control commands (e.g., start/stop capture, set channel/bandwidth) back to the PC-side engine.

## 2. Technology Stack

*   **Framework:** React (with TypeScript)
    *   Initialized using `create-react-app --template typescript`.
*   **State Management:** React Context API (`useContext` and `useReducer`)
    *   Centralized in `src/contexts/DataContext.tsx`.
*   **Real-time Communication:** Native WebSocket API
    *   Managed in `src/services/websocketService.ts`.
*   **Styling:**
    *   **Primary:** CSS Modules (`*.module.css`) for component-level styles to ensure encapsulation and avoid naming conflicts.
    *   **Global Styles:** Standard CSS files (`src/index.css`, `src/App.css`) for global definitions, CSS variables, and base layout.
    *   **CSS Variables:** Extensively used for theming, color palette, typography, and consistent spacing, defined primarily in `src/index.css`.
*   **UI Components:** Custom-built React components, with a focus on creating a library of common/atomic components.
*   **Charting Library:** Recharts
    *   Selected for its good React integration, rich chart types, and customization options.
    *   Used in `PerformanceDetailPanel.tsx` for displaying historical performance data.

## 3. Project Structure (within `desktop_app/WifiPcapAnalyzer/frontend/`)

```
desktop_app/WifiPcapAnalyzer/frontend/
├── public/
│   ├── index.html
│   ├── fonts/  // For web fonts like SF Pro (if used)
│   └── icons/  // For SVG icons (if not embedded as React components)
│   └── ... (other static assets)
├── src/
│   ├── assets/
│   │   └── fonts/ // Alternative location for font files if managed by build process
│   ├── components/
│   │   ├── common/ // Reusable atomic UI components
│   │   │   ├── Button/
│   │   │   │   ├── Button.tsx
│   │   │   │   └── Button.module.css
│   │   │   ├── Input/
│   │   │   │   ├── Input.tsx
│   │   │   │   └── Input.module.css
│   │   │   ├── Card/
│   │   │   │   ├── Card.tsx
│   │   │   │   └── Card.module.css
│   │   │   ├── Table/
│   │   │   │   ├── Table.tsx
│   │   │   │   └── Table.module.css
│   │   │   ├── Tabs/
│   │   │   │   ├── Tabs.tsx
│   │   │   │   └── Tabs.module.css
│   │   │   └── Icon/
│   │   │       ├── Icon.tsx
│   │   │       └── icons/ // Directory for SVG files if converted to React components
│   │   ├── layout/ (Optional, for complex layout structures)
│   │   │   ├── MainLayout.tsx
│   │   │   └── ...
│   │   ├── BssList/
│   │   │   ├── BssList.tsx
│   │   │   └── BssList.module.css (or BssList.css if not yet migrated)
│   │   ├── StaList/
│   │   │   ├── StaList.tsx
│   │   │   └── StaList.module.css (or StaList.css)
│   │   └── ControlPanel/
│   │       ├── ControlPanel.tsx
│   │       └── ControlPanel.module.css (or ControlPanel.css)
│   ├── contexts/
│   │   └── DataContext.tsx
│   ├── services/
│   │   └── websocketService.ts
│   ├── styles/ // Optional: For additional global style partials if index.css becomes too large
│   │   └── _variables.css
│   │   └── _base.css
│   ├── types/
│   │   └── data.ts
│   ├── App.css
│   ├── App.tsx
│   ├── index.css  // Global styles, CSS Variables, Font definitions
│   ├── index.tsx
│   └── ... (other CRA files)
├── package.json
├── tsconfig.json
└── README.md
```

## 4. Core Components and Logic (Post UI/UX Redesign)

### 4.1. Styling Strategy
*   **Global Styles (`src/index.css`):**
    *   Defines CSS variables for the new color palette (`--graphite-gray: #1F242B;`, `--misty-white: #F5F7F9;`, `--tech-blue: #1E90FF;`).
    *   Defines font variables (`--font-primary: "SF Pro", ...;`, `--font-fallback: ...;`).
    *   Defines base layout variables (`--border-radius-base: 8px;`, `--shadow-light: ...;`).
    *   Applies base HTML element styling (body, headings, etc.) using these variables.
    *   Includes `@font-face` rules for SF Pro if self-hosted.
*   **Application Layout (`src/App.css`, `src/components/layout/`):**
    *   Styles for the main application shell, sidebars, content areas, using global CSS variables.
*   **Component Styles (`src/components/**/*.module.css`):**
    *   Each component will have its own CSS Module file for scoped styles.
    *   Components will heavily utilize global CSS variables for consistency.
    *   Existing CSS files (e.g., `ControlPanel.css`) will be migrated to CSS Modules or strictly follow BEM and use CSS variables.

### 4.2. Common Components (`src/components/common/`)
This new directory will house a set of generic, reusable UI components built according to the new UI/UX specifications.
*   **`Button.tsx`**: A versatile button component supporting different states (primary, secondary, disabled) and styles (e.g., icon buttons).
*   **`Input.tsx`**: Styled input fields (text, number, select if generic).
*   **`Card.tsx`**: A container component for displaying grouped information, with defined padding, border-radius, and shadow.
*   **`Table.tsx`**: A flexible table component for displaying BSS/STA lists, potentially with sorting and filtering capabilities.
*   **`Tabs.tsx`**: For tabbed navigation if needed in the UI.
*   **`Icon.tsx`**: A component to render SVG icons, allowing easy management and styling (e.g., color via `currentColor` or props). SVG icons will be stored in `src/components/common/Icon/icons/` and possibly converted to React components using SVGR.

### 4.3. Feature Components (Refactored)
*   **`ControlPanel.tsx`**: Will be refactored to use the new common `Button`, `Input`, etc., components. Its layout will adhere to the new grid-based design.
*   **`BssList.tsx` / `StaList.tsx`**: Will be refactored to use the common `Table` or a list of `Card` components for displaying data. Each item will be styled according to the new specifications.

### 4.4. `websocketService.ts` (`src/services/`)
*   No major architectural changes anticipated, but will continue to manage WebSocket connection and message handling.

### 4.5. `data.ts` (`src/types/`)
*   Interfaces (`BSS`, `STA`, `WebSocketData`, `ControlCommand`) will be reviewed and updated if the backend data structure changes or if new UI elements require different data shapes.

### 4.6. `DataContext.tsx` (`src/contexts/`)
*   Will continue to manage global application state (BSS/STA lists, connection status).
*   The structure of data stored (e.g., `bssList`, `staList`) will align with the types defined in `data.ts`.

## 5. WebSocket Message Format
*   Remains as previously defined unless backend changes necessitate updates. Frontend components will consume data based on the `WebSocketData` interface in `src/types/data.ts`.

## 6. Font and Resource Management
*   **SF Pro Font:**
    *   **Primary Method:** Attempt to use local "SF Pro" if available (macOS).
    *   **Fallback/Cross-Platform:** Host WOFF2 versions of SF Pro (ensure licensing) in `public/fonts/` or `src/assets/fonts/` and declare via `@font-face` in `src/index.css`.
    *   CSS variables `--font-primary` and `--font-fallback` will manage font stacks.
*   **SVG Icons:**
    *   Store SVG icons in `src/components/common/Icon/icons/`.
    *   Use SVGR (often integrated with CRA) to import them as React components.
    *   The `Icon.tsx` component will serve as a centralized way to render these icons, allowing for props like `size`, `color` (via `fill="currentColor"` and CSS).

## 7. Build and Run
*   No changes to the existing `npm start` and `npm run build` processes.
*   Ensure ESLint and Prettier are configured for code quality and consistency.

## 8. Maintainability and Scalability
*   **Modularity:** Emphasis on small, reusable common components and CSS Modules.
*   **Clear Structure:** Logical organization of files and directories.
*   **CSS Variables:** Centralized theme and style control.
*   **TypeScript:** Strong typing for better code quality and refactoring.
*   **Documentation:** Key components and architectural decisions will be documented here and potentially using Storybook for common components.

## 9. UI/UX Specification Adherence
*   **Color Palette:** Graphite Gray (`#1F242B`), Misty White (`#F5F7F9`), Tech Blue (`#1E90FF`) will be defined as CSS variables and used throughout.
*   **Typography:** SF Pro (with fallbacks) applied via CSS variables.
*   **Layout:** Grid-based layout principles will be applied to `App.tsx` and feature components.
*   **Rounded Corners:** `8px` border-radius applied via a CSS variable (`--border-radius-base`).
*   **Shadows:** Lightweight shadows applied via a CSS variable (`--shadow-light`).
*   **WCAG AA Contrast:** Color combinations will be chosen to meet WCAG AA contrast ratios.

## 10. Previous Implementation Details (Pre-Redesign) - For Reference
*(This section includes the content from the previously read version of this document for historical context. The sections above describe the new architectural direction.)*

### 10.1. Original Overview (from previous version)
The Web Frontend is responsible for:
*   Connecting to the PC-side Real-time Analysis Engine via WebSocket.
*   Receiving and displaying BSS (Basic Service Set) and STA (Station) information in real-time.
*   Allowing users to send control commands (e.g., start/stop capture, set channel/bandwidth) back to the PC-side engine.

### 10.2. Original Technology Stack (from previous version)
*   **Framework:** React (with TypeScript)
    *   Initialized using `create-react-app --template typescript`.
*   **State Management:** React Context API (`useContext` and `useReducer`)
    *   Centralized in `src/contexts/DataContext.tsx`.
*   **Real-time Communication:** Native WebSocket API
    *   Managed in `src/services/websocketService.ts`.
*   **Styling:** CSS Modules (implicit via CRA naming conventions like `Component.module.css`, though standard CSS files like `App.css`, `BssList.css`, `ControlPanel.css` were used directly in this iteration for simplicity).
*   **UI Components:** Custom-built React components. No external UI component library was used in this initial setup.

### 10.3. Original Project Structure (within `web_frontend/`) (from previous version)
```
web_frontend/
├── public/
│   └── index.html
│   └── ... (other static assets)
├── src/
│   ├── components/
│   │   ├── BssList/
│   │   │   ├── BssList.tsx
│   │   │   └── BssList.css
│   │   └── ControlPanel/
│   │       ├── ControlPanel.tsx
│   │       └── ControlPanel.css
│   ├── contexts/
│   │   └── DataContext.tsx
│   ├── services/
│   │   └── websocketService.ts
│   ├── types/
│   │   └── data.ts
│   ├── App.css
│   ├── App.tsx
│   ├── index.css
│   ├── index.tsx
│   └── ... (other CRA files)
├── package.json
├── tsconfig.json
└── README.md
```

### 10.4. Original Core Components and Logic (from previous version)
#### 10.4.1. `websocketService.ts` (`src/services/`)
*   Manages the WebSocket connection lifecycle (connect, open, message, close, error).
*   Provides functions to send messages (`sendMessage`) and manage message listeners (`addMessageListener`, `removeMessageListener`).
*   Defines the WebSocket URL (default: `ws://localhost:8080/ws`).

#### 10.4.2. `data.ts` (`src/types/`)
*   Defines TypeScript interfaces for `BSS`, `Station`, `WebSocketData` (expected data structure from server), and `ControlCommand` (structure for commands sent to server).

#### 10.4.3. `DataContext.tsx` (`src/contexts/`)
*   Implements a React Context (`AppStateContext`, `AppDispatchContext`) for global state management.
*   Uses a reducer (`appReducer`) to handle state updates (e.g., `SET_DATA`, `SET_CONNECTED`).
*   The `DataProvider` component initializes the WebSocket connection and sets up message handling.
*   Exports `useAppState` and `useAppDispatch` hooks for easy state access.
*   Exports `sendControlCommand` utility function to dispatch commands via WebSocket, including a check for connection status.

#### 10.4.4. `BssList.tsx` and `BssList.css` (`src/components/BssList/`)
*   Displays a list of BSSs received from the WebSocket.
*   Allows selecting a BSS to view its associated STAs.
*   `BssItem` sub-component renders individual BSS details.
*   `StaListItem` sub-component renders individual STA details.
*   Uses `useAppState` to access BSS data.

#### 10.4.5. `ControlPanel.tsx` and `ControlPanel.css` (`src/components/ControlPanel/`)
*   Provides UI elements (inputs, select, buttons) for users to:
    *   Set channel and bandwidth.
    *   Start and stop packet capture.
*   Uses `useAppState` to check WebSocket connection status.
*   Uses the `sendControlCommand` function (imported from `DataContext`) to send commands.

#### 10.4.6. `App.tsx` and `App.css` (`src/`)
*   Main application component.
*   Wraps the entire application with `DataProvider`.
*   Sets up the overall layout, including a header, main content area (for `ControlPanel` and `BssList`), and a footer.
*   `App.css` provides global styles and layout structure.

### 10.5. Original WebSocket Message Format (Assumptions) (from previous version)
#### 10.5.1. Data from PC Engine to Web Frontend
*   The frontend expects a JSON message from the server that conforms to the `WebSocketData` interface:
    ```json
    {
      "bssList": [
        {
          "bssid": "AA:BB:CC:DD:EE:FF",
          "ssid": "MyWiFi",
          "channel": 6,
          "bandwidth": "20MHz",
          "security": "WPA2-PSK",
          "associatedStations": [
            {
              "mac": "11:22:33:44:55:66",
              "aid": 1,
              "state": "Associated",
              "capabilities": ["ShortPreamble"],
              "lastSeen": 1678886400000,
              "signalStrength": -55,
              "rxBytes": 10240,
              "txBytes": 5120
            }
          ],
          "stationCount": 1,
          "lastSeen": 1678886400100,
          "signalStrength": -50
        }
        // ... more BSS objects
      ]
      // Potentially other global state information
    }
    ```
*   Currently, `DataContext.tsx` assumes the server sends the complete `bssList` on each update.

#### 10.5.2. Control Commands from Web Frontend to PC Engine
*   The frontend sends JSON messages conforming to the `ControlCommand` interface:
    *   **Start Capture:**
        ```json
        { "action": "start_capture" }
        ```
    *   **Stop Capture:**
        ```json
        { "action": "stop_capture" }
        ```
    *   **Set Channel:**
        ```json
        { "action": "set_channel", "payload": { "channel": 11 } }
        ```
    *   **Set Bandwidth:**
        ```json
        { "action": "set_bandwidth", "payload": { "bandwidth": "40" } } // e.g., "20", "40", "80"
        ```

### 10.6. Original How to Build and Run (from previous version)
1.  **Navigate to the frontend directory:**
    ```bash
    cd web_frontend
    ```
2.  **Install dependencies (if not already done by `create-react-app`):**
    ```bash
    npm install
    ```
3.  **Start the development server:**
    ```bash
    npm start
    ```
    This will typically open the application in a web browser at `http://localhost:3000`.

4.  **Build for production:**
    ```bash
    npm run build
    ```

### 10.7. Original Challenges and Solutions (During Initial Setup) (from previous version)
*   **`create-react-app` Path:** The `create-react-app` command was executed from the workspace root, and it created the `web_frontend` directory inside `router_agent/` instead of directly under the workspace root as initially planned. This was noted, and subsequent file paths were adjusted accordingly.
*   **TypeScript Import Error:** An incorrect import path for `sendControlCommand` in `ControlPanel.tsx` caused a TypeScript error. This was resolved by:
    1.  Confirming `sendControlCommand` was correctly exported from `DataContext.tsx`.
    2.  Correcting the import path in `ControlPanel.tsx` to point to `../../contexts/DataContext` instead of `../../services/websocketService`.

### 10.8. Original Recent Fixes and Updates (As of 2025-05-07) (from previous version)
This section details fixes applied to address issues identified during testing or development.

#### 10.8.1. UI Alignment Error in Control Panel (Issue 2.1)
*   **Description:** Buttons within the `ControlPanel` (initially "Start Capture" and "Stop Capture", later also "Set Channel" and "Set Bandwidth") were misaligned and could overflow the panel boundaries, especially on narrower screens.
*   **File Modified:** `web_frontend/src/components/ControlPanel/ControlPanel.css`
*   **Changes:**
    *   Initially, `flex-wrap: wrap;` was added to the `.action-buttons` CSS rule. This addressed the "Start Capture" and "Stop Capture" buttons.
    *   Subsequently, to address the "Set Channel" and "Set Bandwidth" buttons (which are in separate `.control-group` divs), `flex-wrap: wrap;` was also added to the `.control-group` CSS rule. This ensures that all control groups allow their items (labels, selects, buttons) to wrap to the next line if there isn't enough horizontal space, preventing overflow and improving alignment for all buttons within the panel.

#### 10.8.2. Channel List Mismatch in Control Panel (Issue 2.2)
*   **Description:** The channel selection in `ControlPanel` was previously a number input field with validation for 2.4GHz channels (1-14). This did not match the user's 5GHz router interface and its supported channels.
*   **File Modified:** `web_frontend/src/components/ControlPanel/ControlPanel.tsx`
*   **Changes:**
    *   Defined an array `fiveGhzChannels` containing the specified 5GHz channels: `[36, 40, ..., 165]`.
    *   Changed the `channel` state's default value from `'1'` to `'149'` (user's current channel).
    *   Replaced the `<input type="number">` for channel selection with a `<select>` dropdown.
    *   Populated the dropdown with `<option>` elements generated from the `fiveGhzChannels` array.
    *   Updated the channel validation logic in `handleSendCommand` to check if the selected channel exists in the `fiveGhzChannels` list.
    *   Updated the label for the channel selection to "Channel (5GHz):".

#### 10.8.3. Runtime Error in BssList Component (Issue 2.3)
*   **Description:** The `BssList` component was throwing a `TypeError: Cannot read properties of undefined (reading 'length')` upon opening the Web UI. This was likely due to `bssList` being `undefined` at some point during initial rendering or data fetching.
*   **Files Modified:**
    *   `web_frontend/src/contexts/DataContext.tsx`
    *   `web_frontend/src/components/BssList/BssList.tsx`
*   **Changes:**
    *   **In `DataContext.tsx`:**
        *   Modified the `SET_DATA` case in `appReducer` to ensure `state.bssList` is always an array, even if `action.payload.bssList` is `undefined`. Changed `bssList: action.payload.bssList` to `bssList: action.payload.bssList || []`.
    *   **In `BssList.tsx`:**
        *   Added an early return condition at the beginning of the `BssList` component. It now checks if `appState` (from `useAppState()`) or `appState.bssList` is `undefined`. If so, it renders a "Initializing data context..." message and returns, preventing further execution until the context and `bssList` are properly initialized. This provides a more robust guard against accessing `bssList` before it's ready.

#### 10.8.4. Update `start_capture` Command Payload (2025-05-07)
*   **Description:** The PC-side analysis engine now requires the `start_capture` command to include an `interface` name, `channel`, and `bandwidth` in its `payload`. The Web Frontend was updated to meet this requirement.
*   **Files Modified:**
    *   `web_frontend/src/components/ControlPanel/ControlPanel.tsx`
    *   `web_frontend/src/types/data.ts`
*   **Changes:**
    *   **In `ControlPanel.tsx`:**
        *   The `handleSendCommand` function was modified. When `action` is `start_capture`, the `payload` is now constructed with:
            *   `interface`: Hardcoded to `"ath1"` as per initial requirements.
            *   `channel`: Taken from the `channel` state (user-selected or default).
            *   `bandwidth`: Taken from the `bandwidth` state (user-selected or default).
        *   Validation for `channel` and `bandwidth` was also added for the `start_capture` case.
    *   **In `types/data.ts`:**
        *   The `ControlCommand` interface's `payload` type definition was updated to include an optional `interface?: string;` field.
*   **Outcome:** The Web Frontend now sends the required `interface`, `channel`, and `bandwidth` in the `payload` of the `start_capture` command, aligning with PC-side engine expectations. The `interface` is currently hardcoded to `"ath1"`.

### 10.9. Original Future Considerations / Potential Improvements (from previous version)
*   **Granular State Updates:** Instead of replacing the entire `bssList` on each WebSocket message, implement more granular updates (add, update, remove individual BSS/STA) for better performance, especially with large datasets. The reducer in `DataContext.tsx` has placeholders for `UPDATE_BSS`, `ADD_BSS`, `REMOVE_BSS` that can be expanded.
*   **Virtualization:** For displaying potentially large lists of BSSs or STAs, consider using a virtualization library like `react-window` or `react-virtualized` to improve rendering performance.
*   **Error Handling and User Feedback:** Enhance error handling for WebSocket connection issues and provide more user-friendly feedback (e.g., toasts, status messages).
*   **Advanced Filtering/Sorting:** Add UI controls for filtering and sorting the BSS/STA lists.
*   **Visualizations:** Integrate charting libraries (e.g., Chart.js, Recharts) to visualize data like signal strength over time or channel utilization.
*   **Component Styling:** Transition to CSS Modules for more robust and scoped styling if the project grows.
*   **Testing:** Add unit and integration tests for components and services.
## Production Build Optimization

*   **Objective:** Optimize the frontend build for production by removing `console` and `debugger` calls.
*   **Changes:**
    *   Updated `vite.config.ts` to remove `console` and `debugger` calls in production builds.
*   **WebSocket Port Configuration:** Make the WebSocket URL and port configurable, perhaps via an environment variable or a settings UI, rather than hardcoding `ws://localhost:8080/ws`.