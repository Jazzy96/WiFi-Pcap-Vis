import React, { createContext, useContext, useReducer, ReactNode, useEffect } from 'react';
import { BSS, STA } from '../types/data'; // Keep BSS, STA types if still relevant
// Remove WebSocket imports
// import { connectWebSocket, addMessageListener, removeMessageListener, getWebSocketState, sendMessage } from '../services/websocketService';
import { EventsOn } from '../../wailsjs/runtime'; // Import Wails runtime
import { state_manager } from '../../wailsjs/go/models'; // Import generated Go models

interface AppState {
  bssList: BSS[]; // Consider if BSS/STA types need update based on wailsjs/go/models.ts
  staList: STA[]; // Consider if BSS/STA types need update based on wailsjs/go/models.ts
  // isConnected: boolean; // No longer needed for Wails
  lastMessageTimestamp: number | null;
  selectedBssidForStaList: string | null; // State for selected BSSID for STA list
  isCapturing: boolean; // New state for capture status
  isPanelCollapsed: boolean; // New state for panel collapse
  selectedPerformanceTarget: { type: 'bss'; id: string } | { type: 'sta'; id: string } | null; // For PerformanceDetailPanel
  isConnected: boolean; // 新增连接状态
}

type Action =
  | { type: 'SET_SNAPSHOT_DATA'; payload: state_manager.Snapshot } // Use Wails snapshot type
  // | { type: 'SET_CONNECTED'; payload: boolean } // No longer needed
  | { type: 'UPDATE_BSS'; payload: BSS } // Keep if manual updates are needed, otherwise snapshot handles it
  | { type: 'ADD_BSS'; payload: BSS } // Keep if manual updates are needed
  | { type: 'REMOVE_BSS'; payload: string }
  | { type: 'SET_SELECTED_BSSID_FOR_STA_LIST'; payload: string | null }
  | { type: 'SET_IS_CAPTURING'; payload: boolean } // New action for capture status
  | { type: 'SET_PANEL_COLLAPSED'; payload: boolean } // New action for panel collapse
  | { type: 'SET_SELECTED_PERFORMANCE_TARGET'; payload: { type: 'bss'; id: string } | { type: 'sta'; id: string } | null } // For PerformanceDetailPanel
  | { type: 'SET_IS_CONNECTED'; payload: boolean }; // 新增连接状态action

const initialState: AppState = {
  bssList: [],
  staList: [],
  // isConnected: false,
  lastMessageTimestamp: null,
  selectedBssidForStaList: null,
  isCapturing: false, // Initialize capture status
  isPanelCollapsed: false, // Default to not collapsed
  selectedPerformanceTarget: null, // Initialize selected performance target
  isConnected: false, // 初始化连接状态为未连接
};

const AppStateContext = createContext<AppState | undefined>(undefined);
const AppDispatchContext = createContext<React.Dispatch<Action> | undefined>(undefined);

// Function to transform backend STA data to frontend format
const transformStaData = (sta: state_manager.STAInfo): STA => {
  // Convert HTCapabilities to HTCabilities format
  const htCapabilities = sta.ht_capabilities ? {
    channel_width_40mhz: sta.ht_capabilities.channel_width_40mhz,
    short_gi_20mhz: sta.ht_capabilities.short_gi_20mhz,
    short_gi_40mhz: sta.ht_capabilities.short_gi_40mhz,
    // Convert the array to string format as expected by frontend
    supported_mcs_set: sta.ht_capabilities.supported_mcs_set ? sta.ht_capabilities.supported_mcs_set.join(',') : undefined,
  } : undefined;

  // Convert VHTCapabilities to VHTCabilities format
  const vhtCapabilities = sta.vht_capabilities ? {
    channel_width_160mhz: sta.vht_capabilities.channel_width_160mhz,
    channel_width_80plus80mhz: sta.vht_capabilities.channel_width_80plus80mhz,
    channel_width_80mhz: sta.vht_capabilities.channel_width_80mhz,
    short_gi_80mhz: sta.vht_capabilities.short_gi_80mhz,
    short_gi_160mhz: sta.vht_capabilities.short_gi_160mhz,
    su_beamformer_capable: sta.vht_capabilities.su_beamformer_capable,
    mu_beamformer_capable: sta.vht_capabilities.mu_beamformer_capable
  } : undefined;

  return {
    mac_address: sta.mac_address,
    associated_bssid: sta.associated_bssid,
    signal_strength: sta.signal_strength,
    last_seen: new Date(sta.last_seen).toISOString(),
    ht_capabilities: htCapabilities,
    vht_capabilities: vhtCapabilities,
    // Convert backend throughput values (bps) to frontend format (Mbps)
    throughput_ul_mbps: sta.uplink_throughput / 1000000,
    throughput_dl_mbps: sta.downlink_throughput / 1000000,
    // Include other metrics
    historical_throughput_ul: sta.historical_uplink_throughput,
    historical_throughput_dl: sta.historical_downlink_throughput,
    historical_channel_utilization: sta.historical_channel_utilization,
    util: sta.util,
    thrpt: sta.thrpt,
    bitrate: sta.bitrate,
    // 使用后端提供的累积统计数据，而不是硬编码为0
    rx_bytes: sta.rx_bytes || 0,
    tx_bytes: sta.tx_bytes || 0, 
    rx_packets: sta.rx_packets || 0,
    tx_packets: sta.tx_packets || 0,
    rx_retries: sta.rx_retries || 0,
    tx_retries: sta.tx_retries || 0,
    // 使用站点比特率作为发送和接收比特率的估计
    rx_bitrate_mbps: sta.bitrate || 0,
    tx_bitrate_mbps: sta.bitrate || 0
  };
};

// Function to transform backend BSS data to frontend format, including its associated STAs
const transformBssData = (bss: state_manager.BSSInfo): BSS => {
  // Convert associated STAs
  const associatedStas: { [mac: string]: STA } = {};
  
  // Process each associated STA in the BSS
  if (bss.associated_stas) {
    Object.entries(bss.associated_stas).forEach(([mac, sta]) => {
      if (sta) {
        // Use the same transformStaData function to ensure consistency
        associatedStas[mac] = transformStaData(sta);
      }
    });
  }

  // Convert HTCapabilities to HTCabilities format
  const htCapabilities = bss.ht_capabilities ? {
    channel_width_40mhz: bss.ht_capabilities.channel_width_40mhz,
    short_gi_20mhz: bss.ht_capabilities.short_gi_20mhz,
    short_gi_40mhz: bss.ht_capabilities.short_gi_40mhz,
    // Convert the array to string format as expected by frontend
    supported_mcs_set: bss.ht_capabilities.supported_mcs_set ? bss.ht_capabilities.supported_mcs_set.join(',') : undefined,
  } : undefined;

  // Convert VHTCapabilities to VHTCabilities format
  const vhtCapabilities = bss.vht_capabilities ? {
    channel_width_160mhz: bss.vht_capabilities.channel_width_160mhz,
    channel_width_80plus80mhz: bss.vht_capabilities.channel_width_80plus80mhz,
    channel_width_80mhz: bss.vht_capabilities.channel_width_80mhz,
    short_gi_80mhz: bss.vht_capabilities.short_gi_80mhz,
    short_gi_160mhz: bss.vht_capabilities.short_gi_160mhz,
    su_beamformer_capable: bss.vht_capabilities.su_beamformer_capable,
    mu_beamformer_capable: bss.vht_capabilities.mu_beamformer_capable
  } : undefined;

  return {
    bssid: bss.bssid,
    ssid: bss.ssid,
    channel: bss.channel,
    bandwidth: bss.bandwidth,
    security: bss.security,
    signal_strength: bss.signal_strength,
    last_seen: new Date(bss.last_seen).toISOString(),
    ht_capabilities: htCapabilities,
    vht_capabilities: vhtCapabilities,
    associated_stas: associatedStas,
    // Convert metrics
    channel_utilization_percent: bss.channel_utilization,
    total_throughput_mbps: bss.throughput / 1000000, // Convert from bps to Mbps
    historical_channel_utilization: bss.historical_channel_utilization,
    historical_total_throughput: bss.historical_throughput,
    util: bss.util,
    thrpt: bss.thrpt
  };
};

const appReducer = (state: AppState, action: Action): AppState => {
  switch (action.type) {
    case 'SET_SNAPSHOT_DATA':
      // console.log("SET_SNAPSHOT_DATA received. Payload:", action.payload); // Less verbose log
      // Assuming snapshot is always the full state when received
      // Note: The types in action.payload (state_manager.BSSInfo, state_manager.STAInfo)
      // might be slightly different from the original frontend types (BSS, STA).
      // You might need to map the data or update the frontend types.
      // For now, assuming direct assignment works or types are compatible.
      // Also, Wails might automatically convert int64 (LastSeen) to number in JS.
      if (state.isCapturing && action.payload) {
         // Transform BSS list including associated STAs
         const bssListFromSnapshot = action.payload.bsss.map(transformBssData);
         
         // Transform standalone STA list
         const staListFromSnapshot = action.payload.stas.map(transformStaData);

        return {
          ...state,
          bssList: bssListFromSnapshot || state.bssList,
          staList: staListFromSnapshot || state.staList,
          lastMessageTimestamp: Date.now(),
        };
      } else if (!state.isCapturing) {
          // console.log("SET_SNAPSHOT_DATA: Received snapshot data but not capturing. Ignoring list update.");
          return { ...state, lastMessageTimestamp: Date.now() }; // Update timestamp only
      }
       console.warn("Received SET_SNAPSHOT_DATA action with invalid payload or while not capturing.");
      return state;
    // case 'SET_CONNECTED': // Removed
    //   return {
    //     ...state,
    //     isConnected: action.payload,
    //   };
    case 'UPDATE_BSS':
      return {
        ...state,
        bssList: state.bssList.map(bss =>
          bss.bssid === action.payload.bssid ? action.payload : bss
        ),
        lastMessageTimestamp: Date.now(),
      };
    case 'ADD_BSS':
      // Avoid duplicates if BSS already exists
      if (state.bssList.find(bss => bss.bssid === action.payload.bssid)) {
        return { // If exists, update it
          ...state,
          bssList: state.bssList.map(bss =>
            bss.bssid === action.payload.bssid ? action.payload : bss
          ),
          lastMessageTimestamp: Date.now(),
        };
      }
      return {
        ...state,
        bssList: [...state.bssList, action.payload],
        lastMessageTimestamp: Date.now(),
      };
    case 'REMOVE_BSS':
      return {
        ...state,
        bssList: state.bssList.filter(bss => bss.bssid !== action.payload),
        lastMessageTimestamp: Date.now(),
      };
    case 'SET_SELECTED_BSSID_FOR_STA_LIST':
      return {
        ...state,
        selectedBssidForStaList: action.payload,
      };
    case 'SET_IS_CAPTURING':
      console.log("Setting isCapturing to:", action.payload); // Log capture state change
      return {
        ...state,
        isCapturing: action.payload,
        // Optionally clear lists when stopping capture? Or keep last state? Keeping last state for now.
      };
    case 'SET_PANEL_COLLAPSED': // Handle new action
      return {
        ...state,
        isPanelCollapsed: action.payload,
      };
    case 'SET_SELECTED_PERFORMANCE_TARGET':
      return {
        ...state,
        selectedPerformanceTarget: action.payload,
      };
    case 'SET_IS_CONNECTED': // 新增连接状态处理
      console.log("Setting isConnected to:", action.payload);
      return {
        ...state,
        isConnected: action.payload,
      };
    default:
      return state;
  }
};

export const DataProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [state, dispatch] = useReducer(appReducer, initialState);

  // Effect for Wails event listeners
  useEffect(() => {
    console.log("Setting up Wails event listeners...");

    const cleanupSnapshot = EventsOn('state_snapshot', (snapshot: state_manager.Snapshot) => {
      // console.log("Received state_snapshot event:", snapshot);
      dispatch({ type: 'SET_SNAPSHOT_DATA', payload: snapshot });
    });

    // Listen for capture status events
    const cleanupCaptureStatus = EventsOn('capture_status', (status: string) => {
      console.log("Received capture_status event:", status);
      dispatch({ type: 'SET_IS_CAPTURING', payload: status === 'started' });
    });

    // Listen for connection status events
    const cleanupConnectionStatus = EventsOn('connection_status', (status: string) => {
      console.log("Received connection_status event:", status);
      dispatch({ type: 'SET_IS_CONNECTED', payload: status === 'connected' });
    });

    // Listen for error events
    const cleanupError = EventsOn('error', (errorMsg: string) => {
      console.error("Received error event:", errorMsg);
      // Optionally show an error toast/notification or update error state
    });

    // Return cleanup function
    return () => {
      cleanupSnapshot();
      cleanupCaptureStatus();
      cleanupConnectionStatus();
      cleanupError();
      console.log("Cleaned up Wails event listeners.");
    };
  }, []);

  return (
    <AppStateContext.Provider value={state}>
      <AppDispatchContext.Provider value={dispatch}>
        {children}
      </AppDispatchContext.Provider>
    </AppStateContext.Provider>
  );
};

export const useAppState = () => {
  const context = useContext(AppStateContext);
  if (context === undefined) {
    throw new Error('useAppState must be used within a DataProvider');
  }
  return context;
};

export const useAppDispatch = () => {
  const context = useContext(AppDispatchContext);
  if (context === undefined) {
    throw new Error('useAppDispatch must be used within a DataProvider');
  }
  return context;
};

// Remove the old WebSocket-based sendControlCommand function
// export const sendControlCommand = (command: import('../types/data').ControlCommand) => { ... };
// Commands should now be sent using window.go.main.App.* directly from components like ControlPanel.