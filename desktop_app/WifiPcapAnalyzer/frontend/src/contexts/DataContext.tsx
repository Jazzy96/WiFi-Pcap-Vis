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
  | { type: 'SET_SELECTED_PERFORMANCE_TARGET'; payload: { type: 'bss'; id: string } | { type: 'sta'; id: string } | null }; // For PerformanceDetailPanel

const initialState: AppState = {
  bssList: [],
  staList: [],
  // isConnected: false,
  lastMessageTimestamp: null,
  selectedBssidForStaList: null,
  isCapturing: false, // Initialize capture status
  isPanelCollapsed: false, // Default to not collapsed
  selectedPerformanceTarget: null, // Initialize selected performance target
};

const AppStateContext = createContext<AppState | undefined>(undefined);
const AppDispatchContext = createContext<React.Dispatch<Action> | undefined>(undefined);

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
         // Type assertion might be needed if types differ significantly
         const bssListFromSnapshot = action.payload.bsss as unknown as BSS[];
         const staListFromSnapshot = action.payload.stas as unknown as STA[];

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
      // console.log("Received Wails event 'state_snapshot':", snapshot);
      dispatch({ type: 'SET_SNAPSHOT_DATA', payload: snapshot });
    });

    const cleanupStatus = EventsOn('capture_status', (status: string) => {
      console.log("Received Wails event 'capture_status':", status);
      dispatch({ type: 'SET_IS_CAPTURING', payload: status === 'started' });
    });

     const cleanupError = EventsOn('error', (errorMessage: string) => {
        console.error("Received Wails event 'error':", errorMessage);
        // Optionally dispatch an action to show the error in the UI
        // dispatch({ type: 'SET_ERROR_MESSAGE', payload: errorMessage });
    });

    // Cleanup function to remove listeners when component unmounts
    return () => {
      console.log("Cleaning up Wails event listeners...");
      // Wails runtime automatically handles cleanup for listeners
      // returned by EventsOn when the component unmounts if using
      // the function reference directly. If using an intermediate
      // function, you might need manual cleanup, but the direct way is preferred.
      // cleanupSnapshot(); // Example if manual cleanup were needed
      // cleanupStatus();
      // cleanupError();
    };
  }, []); // Empty dependency array ensures this runs only once on mount

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