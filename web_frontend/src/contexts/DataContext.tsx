import React, { createContext, useContext, useReducer, ReactNode, useEffect } from 'react';
import { BSS, STA, WebSocketData } from '../types/data';
import { connectWebSocket, addMessageListener, removeMessageListener, getWebSocketState, sendMessage } from '../services/websocketService';

interface AppState {
  bssList: BSS[];
  staList: STA[];
  isConnected: boolean;
  lastMessageTimestamp: number | null;
  selectedBssidForStaList: string | null; // State for selected BSSID for STA list
  isCapturing: boolean; // New state for capture status
}

type Action =
  | { type: 'SET_DATA'; payload: WebSocketData }
  | { type: 'SET_CONNECTED'; payload: boolean }
  | { type: 'UPDATE_BSS'; payload: BSS }
  | { type: 'ADD_BSS'; payload: BSS }
  | { type: 'REMOVE_BSS'; payload: string }
  | { type: 'SET_SELECTED_BSSID_FOR_STA_LIST'; payload: string | null }
  | { type: 'SET_IS_CAPTURING'; payload: boolean }; // New action for capture status

const initialState: AppState = {
  bssList: [],
  staList: [],
  isConnected: false,
  lastMessageTimestamp: null,
  selectedBssidForStaList: null,
  isCapturing: false, // Initialize capture status
};

const AppStateContext = createContext<AppState | undefined>(undefined);
const AppDispatchContext = createContext<React.Dispatch<Action> | undefined>(undefined);

const appReducer = (state: AppState, action: Action): AppState => {
  switch (action.type) {
    case 'SET_DATA':
      console.log("SET_DATA action received. Payload:", JSON.stringify(action.payload, null, 2));
      console.log("Type of action.payload:", typeof action.payload);
      if (action.payload) {
        console.log("action.payload.type:", action.payload.type);
        console.log("action.payload.data exists:", !!action.payload.data);
        if (action.payload.data) {
          console.log("action.payload.data.bsss exists:", !!action.payload.data.bsss);
          console.log("action.payload.data.stas exists:", !!action.payload.data.stas);
        }
      }

      // Only update lists if capturing is active
      if (state.isCapturing && action.payload && action.payload.data && action.payload.type === 'snapshot') {
        console.log("SET_DATA: Processing snapshot data while capturing.");
        return {
          ...state,
          // Update lists only if capturing
          bssList: action.payload.data.bsss || state.bssList, // Keep old data if not capturing or payload invalid
          staList: action.payload.data.stas || state.staList, // Keep old data if not capturing or payload invalid
          lastMessageTimestamp: Date.now(),
        };
      } else if (!state.isCapturing && action.payload && action.payload.type === 'snapshot') {
          console.log("SET_DATA: Received snapshot data but not capturing. Ignoring list update.");
          // Still update timestamp maybe? Or just return state.
          return { ...state, lastMessageTimestamp: Date.now() }; // Update timestamp but not lists
      }
      console.warn("Received SET_DATA action with unexpected payload structure, type, or not capturing. Actual payload:", action.payload);
      return state; // Return current state if not capturing or payload invalid
    case 'SET_CONNECTED':
      return {
        ...state,
        isConnected: action.payload,
      };
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
    default:
      return state;
  }
};

export const DataProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [state, dispatch] = useReducer(appReducer, initialState);

  useEffect(() => {
    const handleOpen = () => {
      dispatch({ type: 'SET_CONNECTED', payload: true });
      console.log("DataProvider: WebSocket connected");
    };

    const handleClose = () => {
      dispatch({ type: 'SET_CONNECTED', payload: false });
      console.log("DataProvider: WebSocket disconnected");
    };

    const handleError = (event: Event) => {
      dispatch({ type: 'SET_CONNECTED', payload: false });
      console.error("DataProvider: WebSocket error", event);
    };

    const handleMessage = (data: WebSocketData) => {
      // Assuming the backend sends the full list each time.
      // If the backend sends deltas, the reducer logic would need to be more complex.
      dispatch({ type: 'SET_DATA', payload: data });
    };

    connectWebSocket(handleOpen, handleClose, handleError);
    addMessageListener(handleMessage);

    return () => {
      removeMessageListener(handleMessage);
      // Potentially close WebSocket connection if component unmounts,
      // though typically you want it open for the app's lifetime.
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

// Utility function to send control commands via WebSocket
export const sendControlCommand = (command: import('../types/data').ControlCommand) => {
  if (getWebSocketState() === WebSocket.OPEN) {
    sendMessage(command);
  } else {
    console.error("Cannot send command: WebSocket is not connected.");
    // Optionally, queue the command or notify the user.
  }
};