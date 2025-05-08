// Defines the structure for BSS and STA data

export interface STA { // Renamed from Station to STA
  mac_address: string; // Match backend field name
  associated_bssid?: string; // Match backend field name
  signal_strength: number | null; // Match backend field name
  last_seen: string; // Match backend field name (ISO string)
  ht_capabilities?: any; // Placeholder, match backend structure if needed
  vht_capabilities?: any; // Placeholder, match backend structure if needed
  // Performance Metrics
  rx_bytes: number;
  tx_bytes: number;
  rx_packets: number;
  tx_packets: number;
  rx_retries: number;
  tx_retries: number;
  rx_bitrate_mbps: number;
  tx_bitrate_mbps: number;
  throughput_ul_mbps: number; // Uplink throughput
  throughput_dl_mbps: number; // Downlink throughput
  historical_throughput_ul?: { timestamp: number; value: number }[];
  historical_throughput_dl?: { timestamp: number; value: number }[];
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
  // Performance Metrics
  channel_utilization_percent: number;
  total_throughput_mbps: number; // Combined UL/DL throughput for the BSS
  historical_channel_utilization?: { timestamp: number; value: number }[];
  historical_total_throughput?: { timestamp: number; value: number }[];
}

// Structure of the data expected from the WebSocket server
// Assuming the backend sends the new performance metrics within BSS and STA objects
export interface WebSocketData {
  type: string; // e.g., "snapshot"
  data: {
    bsss: BSS[]; // Now includes performance metrics
    stas: STA[]; // Now includes performance metrics
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