import React, { useState } from 'react';
import { useAppState, useAppDispatch } from '../../contexts/DataContext';
import { StartCapture, StopCapture } from '../../../wailsjs/go/main/App';
import styles from './ControlPanel.module.css'; // Import CSS Modules
import Button from '../common/Button/Button';
import Input from '../common/Input/Input';

export const ControlPanel: React.FC = () => {
  const { isCapturing: isCapturingFromContext, isPanelCollapsed } = useAppState();
  const dispatch = useAppDispatch();
  const fiveGhzChannels: number[] = [36, 40, 44, 48, 52, 56, 60, 64, 100, 104, 108, 112, 116, 120, 124, 128, 132, 136, 140, 144, 149, 153, 157, 161, 165];
  const [channel, setChannel] = useState<string>('149');
  const [bandwidth, setBandwidth] = useState<string>('20');

  const toggleCollapse = () => {
    dispatch({ type: 'SET_PANEL_COLLAPSED', payload: !isPanelCollapsed });
  };

  const handleCaptureToggle = () => {
    if (isCapturingFromContext) {
      StopCapture()
        .then(() => {
          console.log("StopCapture command sent successfully via Wails.");
          dispatch({ type: 'SET_IS_CAPTURING', payload: false });
        })
        .catch(err => {
          console.error("Error sending StopCapture command via Wails:", err);
          alert(`Error stopping capture: ${err}`);
        });
    } else {
      const ch = parseInt(channel, 10);
      const bw = `${bandwidth}MHz`; 

      if (isNaN(ch) || !fiveGhzChannels.includes(ch)) {
        alert(`Invalid channel for capture. Please select a valid 5GHz channel.`);
        return;
      }
      if (!['20', '40', '80', '160'].includes(bandwidth)) {
        alert('Invalid bandwidth for capture. Choose from 20, 40, 80, 160 MHz.');
        return;
      }

      const interfaceName = "ath1"; // Still hardcoded
      const bpfFilter = ""; 

      StartCapture(interfaceName, ch, bw, bpfFilter)
        .then(() => {
          console.log(`StartCapture command sent successfully via Wails for interface ${interfaceName}, channel ${ch}, bandwidth ${bw}.`);
          dispatch({ type: 'SET_IS_CAPTURING', payload: true });
        })
        .catch(err => {
          console.error("Error sending StartCapture command via Wails:", err);
          alert(`Error starting capture: ${err}`);
        });
    }
  };

  const handleSetChannel = () => {
      console.warn("Set Channel functionality not implemented via Wails yet.");
      // Example: window.go.main.App.SetChannel(parseInt(channel, 10)).then(...).catch(...);
  }
   const handleSetBandwidth = () => {
      console.warn("Set Bandwidth functionality not implemented via Wails yet.");
      // Example: window.go.main.App.SetBandwidth(`${bandwidth}MHz`).then(...).catch(...);
  }

  return (
    <div className={`${styles.controlPanel} ${isPanelCollapsed ? styles.collapsed : ''}`}>
      <div className={styles.panelHeader}>
        <h2 className={styles.panelTitle}>Control Panel</h2>
        <Button
          onClick={toggleCollapse}
          className={styles.collapseButton}
          aria-label={isPanelCollapsed ? 'Expand Panel' : 'Collapse Panel'}
          variant="secondary" // Or a more appropriate variant
        >
          {isPanelCollapsed ? '❯' : '❮'}
        </Button>
      </div>
      <div className={styles.panelContent}>
        {!isPanelCollapsed && (
          <>
            <Input
              type="select"
              label="Channel (5GHz):"
              id="channel-select"
              value={channel}
              onChange={(e) => setChannel(e.target.value)}
              disabled={isCapturingFromContext}
              options={fiveGhzChannels.map(ch => ({ value: ch, label: ch.toString() }))}
              containerClassName={styles.controlGroup}
              className={styles.selectInput} // Add specific class if needed for select
            />
            {/* The "Set" button for channel is NYI, so we can omit it or use a disabled Button */}
            {/* For now, let's keep it similar to before but using the new Button component */}
            <div className={styles.inputRow}> {/* This div might need to be adjusted or removed depending on final layout */}
                 <Button onClick={handleSetChannel} disabled={isCapturingFromContext} variant="secondary" className={styles.setButton}>
                   Set (NYI)
                 </Button>
            </div>


            <Input
              type="select"
              label="Bandwidth (MHz):"
              id="bandwidth-select"
              value={bandwidth}
              onChange={(e) => setBandwidth(e.target.value)}
              disabled={isCapturingFromContext}
              options={[
                { value: '20', label: '20 MHz' },
                { value: '40', label: '40 MHz' },
                { value: '80', label: '80 MHz' },
                { value: '160', label: '160 MHz' },
              ]}
              containerClassName={styles.controlGroup}
              className={styles.selectInput} // Add specific class if needed for select
            />
             <div className={styles.inputRow}> {/* This div might need to be adjusted or removed depending on final layout */}
                <Button onClick={handleSetBandwidth} disabled={isCapturingFromContext} variant="secondary" className={styles.setButton}>
                  Set (NYI)
                </Button>
            </div>

            <div className={styles.actionButtons}>
              <Button
                variant={isCapturingFromContext ? 'secondary' : 'primary'} // Example: stop is secondary, start is primary
                className={`${styles.actionButton} ${isCapturingFromContext ? styles.stop : styles.start}`}
                onClick={handleCaptureToggle}
              >
                {isCapturingFromContext ? 'Stop Capture' : 'Start Capture'}
              </Button>
            </div>
          </>
        )}
      </div>
    </div>
  );
};