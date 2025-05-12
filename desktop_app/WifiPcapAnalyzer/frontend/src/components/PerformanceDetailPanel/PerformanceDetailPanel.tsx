import React from 'react';
import { useAppState } from '../../contexts/DataContext';
import { BSS, STA } from '../../types/data';
import styles from './PerformanceDetailPanel.module.css';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

// 开发模式标志 - 在实际应用中可以通过环境变量或构建配置设置
// 这里设为true以便在开发环境中显示模拟数据
const IS_DEVELOPMENT_MODE = true; // 设置为true以便在测试时显示模拟数据图表

// Helper function to format timestamp for chart
const formatTimestamp = (timestamp: number) => {
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
};

// Helper function to convert array of numbers to chart-compatible format
const transformHistoricalData = (data: number[] | undefined): { timestamp: number; value: number }[] => {
  if (!data || data.length === 0) return [];
  
  const now = Date.now();
  // Assume each data point is approximately 1 second apart (adjust based on actual data collection interval)
  const interval = 1000; 
  
  return data.map((value, index) => {
    // Calculate timestamps by working backwards from current time
    // Most recent value has the highest index
    const timestamp = now - (data.length - 1 - index) * interval;
    return { timestamp, value };
  });
};

// 添加生成随机历史数据的函数（仅用于开发/演示）
const generateMockThroughputData = (count: number = 30, maxValue: number = 50): number[] => {
  // 生成带有一定趋势的随机数据
  const result: number[] = [];
  let currentValue = Math.random() * maxValue * 0.5; // 起始值在最大值的一半以内

  for (let i = 0; i < count; i++) {
    // 添加一些随机波动，但保持一定趋势
    const trend = Math.sin(i * 0.2) * maxValue * 0.3; // 正弦波趋势
    const randomFactor = (Math.random() - 0.5) * maxValue * 0.2; // 随机波动
    
    currentValue = Math.max(0, Math.min(maxValue, currentValue + trend + randomFactor));
    // 将bps值转换为较大的数值，以模拟实际吞吐量数据
    result.push(currentValue * 1000000); // 转换为bps
  }
  
  return result;
};

// 根据当前值生成趋势一致的历史数据，确保最后一个值等于当前值
const generateConsistentHistoryData = (currentValue: number, count: number = 30): number[] => {
  // 如果当前值是undefined或null，生成默认的模拟数据
  if (currentValue === undefined || currentValue === null) {
    return generateMockThroughputData(count, 100);
  }

  const result: number[] = [];
  // 从当前值开始，向前模拟历史数据
  // 添加一些随机波动，但保持趋势接近当前值
  let value = currentValue;
  
  // 反向生成数据（从现在向过去）
  for (let i = 0; i < count; i++) {
    // 最后一个点的值应该等于当前值
    if (i === 0) {
      result.push(value);
      continue;
    }
    
    // 生成随机波动，但保持趋势
    const randomFactor = (Math.random() - 0.5) * 0.2; // 小波动
    // 随机值比例，越往前波动越大
    const variationFactor = 1 + randomFactor * (i / count * 2);
    value = value * variationFactor;
    
    // 确保值不会变得太小或太大
    value = Math.max(value * 0.5, Math.min(value * 1.5, value));
    
    result.push(value);
  }
  
  // 反转数组，使得时间顺序从过去到现在
  return result.reverse();
};

// Helper function to calculate Y axis domain for channel utilization charts
const calculateYAxisDomain = (data: number[] | undefined): [number, number] => {
  if (!data || data.length === 0) return [0, 10]; // 如果没有数据，默认0-10%范围
  
  // 查找最大值
  const maxValue = Math.max(...data);
  
  // 根据最大值计算合适的上限
  let upperBound;
  
  if (maxValue < 10) {
    upperBound = Math.ceil((maxValue * 1.1) / 5) * 5;
  } else {
    // 如果最大值大于50%，上限设为最大值+10%，并向上取整到10的倍数
    upperBound = Math.ceil((maxValue * 1.1) / 10) * 10;
  }
  
  // 确保上限不超过100%
  upperBound = Math.min(upperBound, 100);
  
  return [0, upperBound];
};

// Helper function to calculate Y axis domain for throughput charts
const calculateThroughputYAxisDomain = (data: number[] | undefined): [number, number] => {
  if (!data || data.length === 0) return [0, 1]; // 如果没有数据，默认0-1Mbps范围
  
  // 将bps转换为Mbps后再计算最大值
  const maxValueMbps = Math.max(...data.map(v => v / 1000000));
  
  // 根据最大值计算合适的上限
  let upperBound;
  
  if (maxValueMbps < 1) {
    // 如果最大值小于1Mbps，上限设为1Mbps
    upperBound = 1;
  } else if (maxValueMbps < 10) {
    // 如果最大值在1-10Mbps之间，上限设为最大值+20%，并向上取整到1的倍数
    upperBound = Math.ceil(maxValueMbps * 1.2);
  } else if (maxValueMbps < 100) {
    // 如果最大值在10-100Mbps之间，上限设为最大值+20%，并向上取整到5的倍数
    upperBound = Math.ceil((maxValueMbps * 1.2) / 5) * 5;
  } else {
    // 如果最大值大于100Mbps，上限设为最大值+20%，并向上取整到20的倍数
    upperBound = Math.ceil((maxValueMbps * 1.2) / 20) * 20;
  }
  
  return [0, upperBound];
};

// Helper function to format capabilities
const formatCapabilities = (caps: any) => {
  if (!caps) return '无';
  const items = [];
  
  if (caps.channel_width_40mhz) items.push('40MHz');
  if (caps.channel_width_80mhz) items.push('80MHz');
  if (caps.channel_width_160mhz) items.push('160MHz');
  if (caps.channel_width_80plus80mhz) items.push('80+80MHz');
  if (caps.short_gi_20mhz) items.push('SGI_20');
  if (caps.short_gi_40mhz) items.push('SGI_40');
  if (caps.short_gi_80mhz) items.push('SGI_80');
  if (caps.short_gi_160mhz) items.push('SGI_160');
  
  return items.length > 0 ? items.join(', ') : '基本能力';
};

export const PerformanceDetailPanel: React.FC = () => {
  const { selectedPerformanceTarget, bssList } = useAppState();

  // 开发/演示环境下使用模拟数据
  const useMockData = IS_DEVELOPMENT_MODE;

  if (!selectedPerformanceTarget) {
    return (
      <div className={`${styles.panel} ${styles.noSelection}`}>
        选择一个BSS或STA以查看详细性能指标
      </div>
    );
  }

  let targetData: BSS | STA | undefined;
  let panelTitle = '性能详情';
  let isBss = false;

  if (selectedPerformanceTarget.type === 'bss') {
    targetData = bssList.find(bss => bss.bssid === selectedPerformanceTarget.id);
    if (targetData) {
      panelTitle = `BSS性能详情: ${targetData.ssid || targetData.bssid}`;
      isBss = true;
    }
  } else {
    // 从所有BSS的关联STA中查找目标STA
    for (const bss of bssList) {
      const sta = Object.values(bss.associated_stas || {}).find(
        s => s.mac_address === selectedPerformanceTarget.id
      );
      if (sta) {
        targetData = sta;
        panelTitle = `STA性能详情: ${sta.mac_address}`;
        break;
      }
    }
  }

  if (!targetData) {
    return (
      <div className={`${styles.panel} ${styles.noSelection}`}>
        未找到所选目标数据
      </div>
    );
  }

  // 类型断言
  const bssTarget = targetData as BSS;
  const staTarget = targetData as STA;
  
  // 根据可见图表数量确定布局类名
  const getChartGridClassName = () => {
    // Since all charts are now displayed in a single column,
    // we just return the base chartGrid class
    return styles.chartGrid;
  };

  return (
    <div className={styles.panel}>
      <h2 className={styles.panelTitle}>{panelTitle}</h2>
      
      <div className={styles.scrollContainer}>
        {/* 基本指标信息部分 */}
        <div className={styles.infoSection}>
          <div className={styles.infoHeader}>基本指标</div>
          <div className={styles.metricsGrid}>
            {isBss ? (
              // BSS指标
              <>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>信道</div>
                  <div className={styles.metricValue}>{bssTarget.channel}</div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>带宽</div>
                  <div className={styles.metricValue}>{bssTarget.bandwidth}</div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>安全类型</div>
                  <div className={styles.metricValue}>{bssTarget.security}</div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>信号强度</div>
                  <div className={`${styles.metricValue} ${styles.highlight}`}>
                    {bssTarget.signal_strength !== null ? `${bssTarget.signal_strength} dBm` : 'N/A'}
                  </div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>关联站点数</div>
                  <div className={styles.metricValue}>{Object.keys(bssTarget.associated_stas || {}).length}</div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>最后更新</div>
                  <div className={styles.metricValue}>{new Date(bssTarget.last_seen).toLocaleTimeString()}</div>
                </div>
              </>
            ) : (
              // STA指标
              <>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>关联的BSS</div>
                  <div className={styles.metricValue}>{staTarget.associated_bssid || 'N/A'}</div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>信号强度</div>
                  <div className={`${styles.metricValue} ${styles.highlight}`}>
                    {staTarget.signal_strength !== null ? `${staTarget.signal_strength} dBm` : 'N/A'}
                  </div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>比特率</div>
                  <div className={styles.metricValue}>
                    {staTarget.bitrate !== undefined ? `${staTarget.bitrate.toFixed(1)} Mbps` : 'N/A'}
                  </div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>最后更新</div>
                  <div className={styles.metricValue}>{new Date(staTarget.last_seen).toLocaleTimeString()}</div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>HT能力</div>
                  <div className={styles.metricValue}>
                    {staTarget.ht_capabilities ? formatCapabilities(staTarget.ht_capabilities) : 'N/A'}
                  </div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>VHT能力</div>
                  <div className={styles.metricValue}>
                    {staTarget.vht_capabilities ? formatCapabilities(staTarget.vht_capabilities) : 'N/A'}
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
        
        {/* 性能指标部分 */}
        <div className={styles.infoSection}>
          <div className={styles.infoHeader}>性能指标</div>
          <div className={styles.metricsGrid}>
            {isBss ? (
              // BSS性能指标
              <>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>信道利用率</div>
                  <div className={`${styles.metricValue} ${styles.highlight}`}>
                    {bssTarget.util !== undefined ? `${bssTarget.util.toFixed(1)}%` : 'N/A'}
                  </div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>总吞吐量</div>
                  <div className={`${styles.metricValue} ${styles.highlight}`}>
                    {bssTarget.thrpt !== undefined ? `${(bssTarget.thrpt/1000000).toFixed(2)} Mbps` : 'N/A'}
                  </div>
                </div>
              </>
            ) : (
              // STA性能指标
              <>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>信道利用率</div>
                  <div className={`${styles.metricValue} ${styles.highlight}`}>
                    {staTarget.util !== undefined ? `${staTarget.util.toFixed(1)}%` : 'N/A'}
                  </div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>吞吐量</div>
                  <div className={`${styles.metricValue} ${styles.highlight}`}>
                    {staTarget.thrpt !== undefined ? `${(staTarget.thrpt/1000000).toFixed(2)} Mbps` : 'N/A'}
                  </div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>上行吞吐量</div>
                  <div className={styles.metricValue}>
                    {staTarget.throughput_ul_mbps !== undefined ? `${staTarget.throughput_ul_mbps.toFixed(2)} Mbps` : 'N/A'}
                  </div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>下行吞吐量</div>
                  <div className={styles.metricValue}>
                    {staTarget.throughput_dl_mbps !== undefined ? `${staTarget.throughput_dl_mbps.toFixed(2)} Mbps` : 'N/A'}
                  </div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>发送比特率</div>
                  <div className={styles.metricValue}>
                    {staTarget.tx_bitrate_mbps !== undefined ? `${staTarget.tx_bitrate_mbps.toFixed(1)} Mbps` : 'N/A'}
                  </div>
                </div>
                <div className={styles.metricItem}>
                  <div className={styles.metricLabel}>接收比特率</div>
                  <div className={styles.metricValue}>
                    {staTarget.rx_bitrate_mbps !== undefined ? `${staTarget.rx_bitrate_mbps.toFixed(1)} Mbps` : 'N/A'}
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
        
        {/* 流量统计表格 - 只有STA才显示 */}
        {!isBss && staTarget && (
          <div className={styles.infoSection}>
            <div className={styles.infoHeader}>流量统计</div>
            <table className={styles.detailsTable}>
              <thead>
                <tr>
                  <th>类型</th>
                  <th>字节数</th>
                  <th>包数</th>
                  <th>重传次数</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>发送(TX)</td>
                  <td>{staTarget.tx_bytes?.toLocaleString() || 'N/A'}</td>
                  <td>{staTarget.tx_packets?.toLocaleString() || 'N/A'}</td>
                  <td>{staTarget.tx_retries?.toLocaleString() || 'N/A'}</td>
                </tr>
                <tr>
                  <td>接收(RX)</td>
                  <td>{staTarget.rx_bytes?.toLocaleString() || 'N/A'}</td>
                  <td>{staTarget.rx_packets?.toLocaleString() || 'N/A'}</td>
                  <td>{staTarget.rx_retries?.toLocaleString() || 'N/A'}</td>
                </tr>
              </tbody>
            </table>
          </div>
        )}
        
        {/* 历史图表部分 - 修改条件判断，添加模拟数据支持 */}
        {(isBss && (
            (useMockData) || 
            ((bssTarget?.historical_channel_utilization && bssTarget.historical_channel_utilization.length > 0) || 
            (bssTarget?.historical_total_throughput && bssTarget.historical_total_throughput.length > 0))
          )) || 
         (!isBss && (
            (useMockData) ||
            ((staTarget?.historical_throughput_ul && staTarget.historical_throughput_ul.length > 0) || 
            (staTarget?.historical_throughput_dl && staTarget.historical_throughput_dl.length > 0) ||
            (staTarget?.historical_channel_utilization && staTarget.historical_channel_utilization.length > 0))
          )) ? (
          <div className={styles.infoSection}>
            <div className={styles.infoHeader}>历史性能图表</div>
            <div className={getChartGridClassName()}>
              {/* BSS图表 */}
              {isBss && (useMockData || (bssTarget?.historical_channel_utilization && bssTarget.historical_channel_utilization.length > 0)) && (
                <div className={styles.chartContainer}>
                  <div className={styles.chartTitle}>信道利用率历史 (%)</div>
                  <ResponsiveContainer width="100%" height={200}>
                    <LineChart data={transformHistoricalData(
                      useMockData && (!bssTarget?.historical_channel_utilization || bssTarget.historical_channel_utilization.length === 0) 
                        ? generateMockThroughputData(30, 60).map(v => v / 1000000) // 模拟数据转换为0-100范围
                        : bssTarget.historical_channel_utilization
                    )}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
                      <XAxis dataKey="timestamp" tickFormatter={formatTimestamp} stroke="#666" />
                      <YAxis 
                        stroke="#666" 
                        domain={calculateYAxisDomain(
                          useMockData && (!bssTarget?.historical_channel_utilization || bssTarget.historical_channel_utilization.length === 0)
                            ? generateMockThroughputData(30, 60).map(v => v / 1000000)
                            : bssTarget.historical_channel_utilization
                        )}
                      />
                      <Tooltip 
                        formatter={(value) => [`${value}%`, '利用率']}
                        labelFormatter={formatTimestamp}
                      />
                      <Line type="monotone" dataKey="value" name="利用率" stroke="#1E90FF" strokeWidth={2} dot={{ r: 2 }} activeDot={{ r: 4 }} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              )}

              {isBss && (useMockData || (bssTarget?.historical_total_throughput && bssTarget.historical_total_throughput.length > 0)) && (
                <div className={styles.chartContainer}>
                  <div className={styles.chartTitle}>总吞吐量历史 (Mbps)</div>
                  <ResponsiveContainer width="100%" height={200}>
                    <LineChart data={transformHistoricalData(
                      useMockData && (!bssTarget?.historical_total_throughput || bssTarget.historical_total_throughput.length === 0)
                        ? generateMockThroughputData(30, 80)
                        : bssTarget.historical_total_throughput
                    ).map(item => ({
                      ...item, 
                      value: item.value / 1000000 // 转换为Mbps
                    }))}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
                      <XAxis dataKey="timestamp" tickFormatter={formatTimestamp} stroke="#666" />
                      <YAxis 
                        stroke="#666" 
                        domain={calculateThroughputYAxisDomain(
                          useMockData && (!bssTarget?.historical_total_throughput || bssTarget.historical_total_throughput.length === 0)
                            ? generateMockThroughputData(30, 80)
                            : bssTarget.historical_total_throughput
                        )}
                      />
                      <Tooltip 
                        formatter={(value) => [`${value} Mbps`, '吞吐量']}
                        labelFormatter={formatTimestamp}
                      />
                      <Line type="monotone" dataKey="value" name="吞吐量" stroke="#2E8B57" strokeWidth={2} dot={{ r: 2 }} activeDot={{ r: 4 }} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              )}

              {/* STA图表 */}
              {!isBss && (useMockData || (staTarget?.historical_throughput_ul && staTarget.historical_throughput_ul.length > 0)) && (
                <div className={styles.chartContainer}>
                  <div className={styles.chartTitle}>上行吞吐量历史 (Mbps)</div>
                  <ResponsiveContainer width="100%" height={200}>
                    <LineChart data={transformHistoricalData(
                      useMockData && (!staTarget?.historical_throughput_ul || staTarget.historical_throughput_ul.length === 0)
                        ? generateMockThroughputData(30, 40)
                        : staTarget.historical_throughput_ul
                    ).map(item => ({
                      ...item, 
                      value: item.value / 1000000 // Convert from bps to Mbps
                    }))}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
                      <XAxis dataKey="timestamp" tickFormatter={formatTimestamp} stroke="#666" />
                      <YAxis 
                        stroke="#666" 
                        domain={calculateThroughputYAxisDomain(
                          useMockData && (!staTarget?.historical_throughput_ul || staTarget.historical_throughput_ul.length === 0)
                            ? generateMockThroughputData(30, 40)
                            : staTarget.historical_throughput_ul
                        )}
                      />
                      <Tooltip 
                        formatter={(value) => [`${value} Mbps`, '上行吞吐量']}
                        labelFormatter={formatTimestamp}
                      />
                      <Line type="monotone" dataKey="value" name="上行" stroke="#FF8C00" strokeWidth={2} dot={{ r: 2 }} activeDot={{ r: 4 }} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              )}

              {!isBss && (useMockData || (staTarget?.historical_throughput_dl && staTarget.historical_throughput_dl.length > 0)) && (
                <div className={styles.chartContainer}>
                  <div className={styles.chartTitle}>下行吞吐量历史 (Mbps)</div>
                  <ResponsiveContainer width="100%" height={200}>
                    <LineChart data={transformHistoricalData(
                      useMockData && (!staTarget?.historical_throughput_dl || staTarget.historical_throughput_dl.length === 0)
                        ? generateMockThroughputData(30, 80)
                        : staTarget.historical_throughput_dl
                    ).map(item => ({
                      ...item, 
                      value: item.value / 1000000 // Convert from bps to Mbps
                    }))}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
                      <XAxis dataKey="timestamp" tickFormatter={formatTimestamp} stroke="#666" />
                      <YAxis 
                        stroke="#666" 
                        domain={calculateThroughputYAxisDomain(
                          useMockData && (!staTarget?.historical_throughput_dl || staTarget.historical_throughput_dl.length === 0)
                            ? generateMockThroughputData(30, 80)
                            : staTarget.historical_throughput_dl
                        )}
                      />
                      <Tooltip 
                        formatter={(value) => [`${value} Mbps`, '下行吞吐量']}
                        labelFormatter={formatTimestamp}
                      />
                      <Line type="monotone" dataKey="value" name="下行" stroke="#9370DB" strokeWidth={2} dot={{ r: 2 }} activeDot={{ r: 4 }} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              )}
              
              {/* STA信道利用率图表 */}
              {!isBss && (useMockData || (staTarget?.historical_channel_utilization && staTarget.historical_channel_utilization.length > 0)) && (
                <div className={styles.chartContainer}>
                  <div className={styles.chartTitle}>信道利用率历史 (%)</div>
                  <ResponsiveContainer width="100%" height={200}>
                    <LineChart data={transformHistoricalData(
                      useMockData && (!staTarget?.historical_channel_utilization || staTarget.historical_channel_utilization.length === 0)
                        ? generateMockThroughputData(30, 40).map(v => v / 1000000) // 模拟数据转换为0-100范围
                        : staTarget.historical_channel_utilization
                    )}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
                      <XAxis dataKey="timestamp" tickFormatter={formatTimestamp} stroke="#666" />
                      <YAxis 
                        stroke="#666" 
                        domain={calculateYAxisDomain(
                          useMockData && (!staTarget?.historical_channel_utilization || staTarget.historical_channel_utilization.length === 0)
                            ? generateMockThroughputData(30, 40).map(v => v / 1000000)
                            : staTarget.historical_channel_utilization
                        )}
                      />
                      <Tooltip 
                        formatter={(value) => [`${value}%`, '利用率']}
                        labelFormatter={formatTimestamp}
                      />
                      <Line type="monotone" dataKey="value" name="利用率" stroke="#1E90FF" strokeWidth={2} dot={{ r: 2 }} activeDot={{ r: 4 }} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              )}
            </div>
          </div>
        ) : null}
      </div>
    </div>
  );
};