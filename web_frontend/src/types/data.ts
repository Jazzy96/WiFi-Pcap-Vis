// Defines the structure for BSS and STA data

export interface STA { // Renamed from Station to STA
  mac_address: string; // Match backend field name
  associated_bssid?: string; // Match backend field name
  signal_strength: number | null; // Match backend field name
  last_seen: string; // Match backend field name (ISO string)
  // Add capabilities if needed, matching backend structure
  ht_capabilities?: any; // Placeholder, match backend structure if needed
  vht_capabilities?: any; // Placeholder, match backend structure if needed
  // rxBytes: number; // Remove if not sent by backend
  // txBytes: number; // Remove if not sent by backend
}

export interface BSS {
  bssid: string;
  ssid: string;
  channel: number;
  bandwidth: string;
  security: string;
  signal_strength: number | null; // Match backend field name
  last_seen: string; // Match backend field name (ISO string)
  ht_capabilities?: any; // Placeholder, match backend structure
  vht_capabilities?: any; // Placeholder, match backend structure
  associated_stas: { [mac: string]: STA }; // Match backend structure (map)
  // stationCount: number; // Remove if not sent directly by backend
}

// Structure of the data expected from the WebSocket server
export interface WebSocketData {
  type: string; // e.g., "snapshot"
  data: {
    bsss: BSS[];
    stas: STA[];
  };
}

// Structure for control commands sent to the WebSocket server
export interface ControlCommand {
  action: 'start_capture' | 'stop_capture' | 'set_channel' | 'set_bandwidth';
  payload?: {
    interface?: string; // Added for start_capture command
    channel?: number;
    bandwidth?: string; // e.g., "20", "40", "80", "160"
    // other potential payload properties
  };
}