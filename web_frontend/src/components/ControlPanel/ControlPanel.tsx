import React, { useState } from 'react';
import { useAppState, useAppDispatch } from '../../contexts/DataContext'; // Import useAppDispatch
import { sendControlCommand } from '../../contexts/DataContext';
import { ControlCommand } from '../../types/data';
import './ControlPanel.css';

export const ControlPanel: React.FC = () => {
  const { isConnected, isCapturing: isCapturingFromContext } = useAppState(); // Get capture state from context
  const dispatch = useAppDispatch(); // Get dispatch function
  const [isCollapsed, setIsCollapsed] = useState(false); // State for collapsing
  // const [isCapturing, setIsCapturing] = useState(false); // Remove local state, use context state
  const fiveGhzChannels: number[] = [36, 40, 44, 48, 52, 56, 60, 64, 100, 104, 108, 112, 116, 120, 124, 128, 132, 136, 140, 144, 149, 153, 157, 161, 165];
  const [channel, setChannel] = useState<string>('149'); // Default 5GHz channel
  const [bandwidth, setBandwidth] = useState<string>('20'); // Default bandwidth

  const toggleCollapse = () => {
    setIsCollapsed(!isCollapsed);
  };

  const handleSendCommand = (action: ControlCommand['action']) => {
    if (!isConnected) {
      alert('WebSocket is not connected. Cannot send command.');
      return;
    }

    let payload: ControlCommand['payload'] = {};
    if (action === 'set_channel') {
      const ch = parseInt(channel, 10);
      if (isNaN(ch) || !fiveGhzChannels.includes(ch)) {
        alert(`Invalid channel. Please select a valid 5GHz channel from the list.`);
        return;
      }
      payload = { channel: ch };
    } else if (action === 'set_bandwidth') {
      if (!['20', '40', '80', '160'].includes(bandwidth)) {
        alert('Invalid bandwidth. Choose from 20, 40, 80, 160 MHz.');
        return;
      }
      payload = { bandwidth: bandwidth };
    } else if (action === 'start_capture') {
      const ch = parseInt(channel, 10);
      // It's good practice to validate channel and bandwidth for start_capture as well
      if (isNaN(ch) || !fiveGhzChannels.includes(ch)) {
        alert(`Invalid channel for capture. Please select a valid 5GHz channel from the list.`);
        return;
      }
      if (!['20', '40', '80', '160'].includes(bandwidth)) {
        alert('Invalid bandwidth for capture. Choose from 20, 40, 80, 160 MHz.');
        return;
      }
      payload = {
        interface: "ath1", // Hardcoded as per requirement
        channel: ch,
        bandwidth: bandwidth
      };
    }

    sendControlCommand({ action, payload });
    console.log(`Command sent: ${action}`, payload);

    // Dispatch capture state change to context
    if (action === 'start_capture') {
      dispatch({ type: 'SET_IS_CAPTURING', payload: true });
    } else if (action === 'stop_capture') {
      dispatch({ type: 'SET_IS_CAPTURING', payload: false });
    }
  };

  const handleCaptureToggle = () => {
    // Use isCapturingFromContext to decide which command to send
    if (isCapturingFromContext) {
      handleSendCommand('stop_capture');
    } else {
      handleSendCommand('start_capture');
    }
  };

  return (
    <div className={`control-panel ${isCollapsed ? 'collapsed' : ''}`}>
      <button onClick={toggleCollapse} className="collapse-button">
        {isCollapsed ? '>' : '<'}
      </button>
      <div className="panel-content">
        <h2>Control Panel</h2>
        {!isCollapsed && (
          <>
            <div className={`status-indicator ${isConnected ? 'connected' : 'disconnected'}`}>
              {isConnected ? 'WebSocket Connected' : 'WebSocket Disconnected'}
            </div>

            <div className="control-group">
              <label htmlFor="channel-select">Channel (5GHz):</label>
              <select
                id="channel-select"
                value={channel}
                onChange={(e) => setChannel(e.target.value)}
                disabled={!isConnected || isCapturingFromContext} // Disable if capturing (use context state)
              >
                {fiveGhzChannels.map((ch) => (
                  <option key={ch} value={ch}>
                    {ch}
                  </option>
                ))}
              </select>
              <button onClick={() => handleSendCommand('set_channel')} disabled={!isConnected || isCapturingFromContext}>
                Set Channel
              </button>
            </div>

            <div className="control-group">
              <label htmlFor="bandwidth-select">Bandwidth (MHz):</label>
              <select
                id="bandwidth-select"
                value={bandwidth}
                onChange={(e) => setBandwidth(e.target.value)}
                disabled={!isConnected || isCapturingFromContext} // Disable if capturing (use context state)
              >
                <option value="20">20 MHz</option>
                <option value="40">40 MHz</option>
                <option value="80">80 MHz</option>
                <option value="160">160 MHz</option>
              </select>
              <button onClick={() => handleSendCommand('set_bandwidth')} disabled={!isConnected || isCapturingFromContext}>
                Set Bandwidth
              </button>
            </div>

            <div className="control-group action-buttons">
              <button
                className={`action-button ${isCapturingFromContext ? 'stop' : 'start'}`} // Use context state for class
                onClick={handleCaptureToggle}
                disabled={!isConnected}
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