import React, { useState, useEffect } from 'react'; // Import useEffect
import { useAppState, useAppDispatch } from '../../contexts/DataContext'; // Import useAppDispatch
// Remove WebSocket command sender
// import { sendControlCommand } from '../../contexts/DataContext';
// Import Wails Go functions
import { StartCapture, StopCapture } from '../../../wailsjs/go/main/App';
// Keep ControlCommand type if still useful for local state, or remove if not needed
// import { ControlCommand } from '../../types/data';
import './ControlPanel.css';

export const ControlPanel: React.FC = () => {
  // Remove isConnected from context usage
  const { isCapturing: isCapturingFromContext, isPanelCollapsed } = useAppState(); // Get capture and panel collapse state from context
  const dispatch = useAppDispatch(); // Get dispatch function
  // const [isCollapsed, setIsCollapsed] = useState(false); // Local state removed, use global isPanelCollapsed
  // const [isCapturing, setIsCapturing] = useState(false); // Remove local state, use context state
  const fiveGhzChannels: number[] = [36, 40, 44, 48, 52, 56, 60, 64, 100, 104, 108, 112, 116, 120, 124, 128, 132, 136, 140, 144, 149, 153, 157, 161, 165];
  const [channel, setChannel] = useState<string>('149'); // Default 5GHz channel
  const [bandwidth, setBandwidth] = useState<string>('20'); // Default bandwidth

  const toggleCollapse = () => {
    dispatch({ type: 'SET_PANEL_COLLAPSED', payload: !isPanelCollapsed });
  };

  // Removed handleSendCommand as logic is moved into handleCaptureToggle and potentially others

  const handleCaptureToggle = () => {
    if (isCapturingFromContext) {
      // Send StopCapture command via Wails
      StopCapture()
        .then(() => {
          console.log("StopCapture command sent successfully via Wails.");
          // Dispatch immediately for UI feedback, actual state confirmed by event
          dispatch({ type: 'SET_IS_CAPTURING', payload: false });
        })
        .catch(err => {
          console.error("Error sending StopCapture command via Wails:", err);
          alert(`Error stopping capture: ${err}`);
        });
    } else {
      // Send StartCapture command via Wails
      const ch = parseInt(channel, 10);
      const bw = `${bandwidth}MHz`; // Assuming Go backend expects MHz suffix, adjust if needed

      // Validate inputs before sending
      if (isNaN(ch) || !fiveGhzChannels.includes(ch)) {
        alert(`Invalid channel for capture. Please select a valid 5GHz channel.`);
        return;
      }
      if (!['20', '40', '80', '160'].includes(bandwidth)) {
        alert('Invalid bandwidth for capture. Choose from 20, 40, 80, 160 MHz.');
        return;
      }

      const interfaceName = "ath1"; // Still hardcoded, consider making configurable
      const bpfFilter = ""; // Add BPF filter input if needed

      StartCapture(interfaceName, ch, bw, bpfFilter)
        .then(() => {
          console.log(`StartCapture command sent successfully via Wails for interface ${interfaceName}, channel ${ch}, bandwidth ${bw}.`);
          // Dispatch immediately for UI feedback, actual state confirmed by event
          dispatch({ type: 'SET_IS_CAPTURING', payload: true });
        })
        .catch(err => {
          console.error("Error sending StartCapture command via Wails:", err);
          alert(`Error starting capture: ${err}`);
        });
    }
  };

  // Placeholder handlers for Set Channel/Bandwidth if needed later
  const handleSetChannel = () => {
      console.warn("Set Channel functionality not implemented via Wails yet.");
      // If implemented in Go:
      // const ch = parseInt(channel, 10);
      // window.go.main.App.SetChannel(ch).then(...).catch(...);
  }
   const handleSetBandwidth = () => {
      console.warn("Set Bandwidth functionality not implemented via Wails yet.");
      // If implemented in Go:
      // const bw = `${bandwidth}MHz`;
      // window.go.main.App.SetBandwidth(bw).then(...).catch(...);
  }

  return (
    <div className={`control-panel ${isPanelCollapsed ? 'collapsed' : ''}`}>
      <div className="panel-header">
        <h2>Control Panel</h2>
        <button onClick={toggleCollapse} className="collapse-button">
          {isPanelCollapsed ? '>' : '<'}
        </button>
      </div>
      <div className="panel-content">
        {!isPanelCollapsed && (
          <>
            {/* Remove WebSocket status indicator */}
            {/* <div className={`status-indicator ${isConnected ? 'connected' : 'disconnected'}`}>
              {isConnected ? 'WebSocket Connected' : 'WebSocket Disconnected'}
            </div> */}

            <div className="control-group">
              <label htmlFor="channel-select">Channel (5GHz):</label>
              <select
                id="channel-select"
                value={channel}
                onChange={(e) => setChannel(e.target.value)}
                disabled={isCapturingFromContext} // Disable only based on capture state
              >
                {fiveGhzChannels.map((ch) => (
                  <option key={ch} value={ch}>
                    {ch}
                  </option>
                ))}
              </select>
              {/* Update button handler and disabled state */}
              <button onClick={handleSetChannel} disabled={isCapturingFromContext}>
                Set Channel (NYI)
              </button>
            </div>

            <div className="control-group">
              <label htmlFor="bandwidth-select">Bandwidth (MHz):</label>
              <select
                id="bandwidth-select"
                value={bandwidth}
                onChange={(e) => setBandwidth(e.target.value)}
                disabled={isCapturingFromContext} // Disable only based on capture state
              >
                <option value="20">20 MHz</option>
                <option value="40">40 MHz</option>
                <option value="80">80 MHz</option>
                <option value="160">160 MHz</option>
              </select>
              {/* Update button handler and disabled state */}
              <button onClick={handleSetBandwidth} disabled={isCapturingFromContext}>
                Set Bandwidth (NYI)
              </button>
            </div>

            <div className="control-group action-buttons">
              <button
                className={`action-button ${isCapturingFromContext ? 'stop' : 'start'}`} // Use context state for class
                onClick={handleCaptureToggle}
                // disabled={!isConnected} // Remove isConnected check
              >
                {isCapturingFromContext ? 'Stop Capture' : 'Start Capture'} {/* Use context state for text */}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
};