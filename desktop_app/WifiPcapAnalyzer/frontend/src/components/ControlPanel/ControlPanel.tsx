import React, { useState, useEffect } from 'react';
import { useAppState, useAppDispatch } from '../../contexts/DataContext';
import { StartCapture, StopCapture } from '../../../wailsjs/go/main/App';
import { GetAppConfig } from '../../../wailsjs/go/main/App';
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import styles from './ControlPanel.module.css'; // Import CSS Modules
import Button from '../common/Button/Button';
import Input from '../common/Input/Input';

// 声明ConnectToAgent和IsConnected函数类型，因为这些是新添加的后端函数
declare const window: {
  go: {
    main: {
      App: {
        ConnectToAgent(serverAddr: string): Promise<void>;
        IsConnected(): Promise<boolean>;
        DisconnectFromAgent(): Promise<void>;
      }
    }
  }
} & Window;

export const ControlPanel: React.FC = () => {
  const { isCapturing: isCapturingFromContext, isConnected: isConnectedFromContext } = useAppState();
  const dispatch = useAppDispatch();
  const fiveGhzChannels: number[] = [36, 40, 44, 48, 52, 56, 60, 64, 100, 104, 108, 112, 116, 120, 124, 128, 132, 136, 140, 144, 149, 153, 157, 161, 165];
  const [channel, setChannel] = useState<string>('149');
  const [bandwidth, setBandwidth] = useState<string>('20');
  const [interfaceName, setInterfaceName] = useState<string>('phy1-mon0');
  
  // 服务器连接配置
  const [serverIP, setServerIP] = useState<string>('192.168.6.171');
  const [serverPort, setServerPort] = useState<string>('50051');
  const [isConnecting, setIsConnecting] = useState<boolean>(false);
  const [isDisconnecting, setIsDisconnecting] = useState<boolean>(false);

  // 组件挂载时，从配置中获取默认的服务器地址
  useEffect(() => {
    GetAppConfig().then(config => {
      if (config.grpc_server_address) {
        const parts = config.grpc_server_address.split(':');
        if (parts.length === 2) {
          setServerIP(parts[0]);
          setServerPort(parts[1]);
        }
      }
    }).catch(err => {
      console.error("Error getting app config:", err);
    });

    // 检查当前连接状态
    window.go.main.App.IsConnected().then((connected: boolean) => {
      dispatch({ type: 'SET_IS_CONNECTED', payload: connected });
    }).catch(err => {
      console.error("Error checking connection status:", err);
    });
  }, [dispatch]);

  // 连接到Agent服务器
  const handleConnect = () => {
    setIsConnecting(true);
    const serverAddr = `${serverIP}:${serverPort}`;
    
    window.go.main.App.ConnectToAgent(serverAddr)
      .then(() => {
        console.log(`Successfully connected to gRPC server at ${serverAddr}`);
        dispatch({ type: 'SET_IS_CONNECTED', payload: true });
      })
      .catch(err => {
        console.error(`Error connecting to gRPC server at ${serverAddr}:`, err);
        alert(`连接失败: ${err}`);
        dispatch({ type: 'SET_IS_CONNECTED', payload: false });
      })
      .finally(() => {
        setIsConnecting(false);
      });
  };

  // 断开与Agent服务器的连接
  const handleDisconnect = () => {
    setIsDisconnecting(true);
    window.go.main.App.DisconnectFromAgent()
      .then(() => {
        console.log("Successfully disconnected from gRPC server.");
        dispatch({ type: 'SET_IS_CONNECTED', payload: false });
      })
      .catch(err => {
        console.error("Error disconnecting from gRPC server:", err);
        alert(`断开连接失败: ${err}`);
        // 根据实际情况，可能需要保留连接状态为true，或者尝试重新获取状态
      })
      .finally(() => {
        setIsDisconnecting(false);
      });
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
        {/* Agent连接配置区域 */}
        <div className={styles.infoSection}>
          <div className={styles.infoHeader}>Agent连接配置</div>
          <div className={styles.paramsGrid}>
            <div className={styles.paramRow}>
              <div className={styles.paramLabel}>服务器IP:</div>
              <div className={styles.paramValue}>
                <Input
                  type="text"
                  id="server-ip-input"
                  value={serverIP}
                  onChange={(e) => setServerIP(e.target.value)}
                  disabled={isConnecting || isConnectedFromContext || isDisconnecting}
                  className={styles.textInput}
                />
              </div>
            </div>
            
            <div className={styles.paramRow}>
              <div className={styles.paramLabel}>端口:</div>
              <div className={styles.paramValue}>
                <Input
                  type="text"
                  id="server-port-input"
                  value={serverPort}
                  onChange={(e) => setServerPort(e.target.value)}
                  disabled={isConnecting || isConnectedFromContext || isDisconnecting}
                  className={styles.textInput}
                />
              </div>
            </div>
            
            <div className={styles.paramRow}>
              <div className={styles.statusLabel}>连接状态:</div>
              <div className={`${styles.statusValue} ${isConnectedFromContext ? styles.connected : styles.disconnected}`}>
                {isConnectedFromContext ? '已连接' : '未连接'}
              </div>
            </div>
            
            <div className={styles.configButtonContainer}>
              {!isConnectedFromContext ? (
                <Button 
                  onClick={handleConnect} 
                  disabled={isConnecting || isCapturingFromContext} // 抓包时不能连接
                  variant="primary" 
                  className={styles.connectButton}
                >
                  {isConnecting ? '连接中...' : '连接'}
                </Button>
              ) : (
                <Button 
                  onClick={handleDisconnect} 
                  disabled={isDisconnecting || isCapturingFromContext} // 抓包时不能断开
                  variant="danger" // 使用danger变体表示断开
                  className={styles.connectButton} // 可以复用connectButton样式或创建新的
                >
                  {isDisconnecting ? '断开中...' : '断开连接'}
                </Button>
              )}
            </div>
          </div>
        </div>
        
        {/* 信道配置区域 */}
        <div className={styles.infoSection}>
          <div className={styles.infoHeader}>信道配置</div>
          <div className={styles.paramsGrid}>
            <div className={styles.paramRow}>
              <div className={styles.paramLabel}>接口名称:</div>
              <div className={styles.paramValue}>
                <Input
                  type="text"
                  id="interface-input"
                  value={interfaceName}
                  onChange={(e) => setInterfaceName(e.target.value)}
                  disabled={isCapturingFromContext}
                  className={styles.textInput}
                />
              </div>
            </div>
            
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
                disabled={isCapturingFromContext || !isConnectedFromContext} 
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
              disabled={!isConnectedFromContext || isCapturingFromContext && isDisconnecting} // 只有在连接成功时才能启用抓包按钮
            >
              {isCapturingFromContext ? '停止抓包' : '开始抓包'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};