import React, { useState } from 'react';
import { useAppState, useAppDispatch } from '../../contexts/DataContext';
import { StartCapture, StopCapture } from '../../../wailsjs/go/main/App';
import styles from './ControlPanel.module.css'; // Import CSS Modules
import Button from '../common/Button/Button';
import Input from '../common/Input/Input';

export const ControlPanel: React.FC = () => {
  const { isCapturing: isCapturingFromContext } = useAppState();
  const dispatch = useAppDispatch();
  const fiveGhzChannels: number[] = [36, 40, 44, 48, 52, 56, 60, 64, 100, 104, 108, 112, 116, 120, 124, 128, 132, 136, 140, 144, 149, 153, 157, 161, 165];
  const [channel, setChannel] = useState<string>('149');
  const [bandwidth, setBandwidth] = useState<string>('20');

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
      const bpfFilter = "type mgt or type data"; // 捕获管理帧和数据帧，以便正确计算吞吐量

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

  const handleSetConfig = () => {
      console.warn("Set Channel and Bandwidth functionality not implemented via Wails yet.");
      // 这里可以合并之前的两个处理函数
      // 示例: window.go.main.App.SetConfig(parseInt(channel, 10), `${bandwidth}MHz`).then(...).catch(...);
  }

  return (
    <div className={styles.controlPanel}>
      <h2 className={styles.panelTitle}>控制面板</h2>
      
      <div className={styles.panelContent}>
        {/* 信道配置区域 */}
        <div className={styles.infoSection}>
          <div className={styles.infoHeader}>信道配置</div>
          <div className={styles.paramsGrid}>
            <div className={styles.paramRow}>
              <div className={styles.paramLabel}>信道 (5GHz):</div>
              <div className={styles.paramValue}>
                <Input
                  type="select"
                  id="channel-select"
                  value={channel}
                  onChange={(e) => setChannel(e.target.value)}
                  disabled={isCapturingFromContext}
                  options={fiveGhzChannels.map(ch => ({ value: ch, label: ch.toString() }))}
                  className={styles.selectInput}
                />
              </div>
            </div>

            <div className={styles.paramRow}>
              <div className={styles.paramLabel}>带宽 (MHz):</div>
              <div className={styles.paramValue}>
                <Input
                  type="select"
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
                  className={styles.selectInput}
                />
              </div>
            </div>
            
            <div className={styles.configButtonContainer}>
              <Button 
                onClick={handleSetConfig} 
                disabled={isCapturingFromContext} 
                variant="secondary" 
                className={styles.configButton}
              >
                设置
              </Button>
            </div>
          </div>
        </div>

        {/* 抓包控制区域 */}
        <div className={styles.infoSection}>
          <div className={styles.infoHeader}>抓包控制</div>
          <div className={styles.captureStatus}>
            <div className={styles.statusLabel}>状态:</div>
            <div className={`${styles.statusValue} ${isCapturingFromContext ? styles.capturing : styles.notCapturing}`}>
              {isCapturingFromContext ? '抓包中' : '未抓包'}
            </div>
          </div>
          
          <div className={styles.actionButtons}>
            <Button
              variant="none"
              className={`${styles.actionButton} ${isCapturingFromContext ? styles.stop : styles.start}`}
              onClick={handleCaptureToggle}
            >
              {isCapturingFromContext ? '停止抓包' : '开始抓包'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};