import React from 'react';
import { useAppState } from '../../contexts/DataContext';
import { BSS, STA } from '../../types/data';
import styles from './PerformanceDetailPanel.module.css';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

// Helper function to format timestamp for chart
const formatTimestamp = (timestamp: number) => {
  return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
};

export const PerformanceDetailPanel: React.FC = () => {
  const { selectedPerformanceTarget, bssList, staList } = useAppState();

  if (!selectedPerformanceTarget) {
    return (
      <div className={`${styles.panel} ${styles.noSelection}`}>
        Select a BSS or STA to view detailed performance metrics.
      </div>
    );
  }

  let targetData: BSS | STA | undefined;
  let panelTitle = 'Performance Details';
  let isBss = false;

  if (selectedPerformanceTarget.type === 'bss') {
    targetData = bssList.find(bss => bss.bssid === selectedPerformanceTarget.id);
    if (targetData) {
      panelTitle = `BSS Performance: ${targetData.ssid || targetData.bssid}`;
      isBss = true;
    }
  } else {
    targetData = staList.find(sta => sta.mac_address === selectedPerformanceTarget.id);
    if (targetData) {
      panelTitle = `STA Performance: ${targetData.mac_address}`;
    }
  }

  if (!targetData) {
    return (
      <div className={`${styles.panel} ${styles.noSelection}`}>
        Selected target data not found.
      </div>
    );
  }

  const bssTarget = targetData as BSS; // Type assertion for BSS specific fields
  const staTarget = targetData as STA; // Type assertion for STA specific fields

  return (
    <div className={styles.panel}>
      <h2 className={styles.panelTitle}>{panelTitle}</h2>

      <div className={styles.metricsGrid}>
        {isBss && bssTarget && (
          <>
            <div className={styles.metricItem}>
              <div className={styles.metricLabel}>Channel Utilization</div>
              <div className={styles.metricValue}>{bssTarget.channel_utilization_percent !== undefined ? `${bssTarget.channel_utilization_percent.toFixed(1)}%` : 'N/A'}</div>
            </div>
            <div className={styles.metricItem}>
              <div className={styles.metricLabel}>Total Throughput</div>
              <div className={styles.metricValue}>{bssTarget.total_throughput_mbps !== undefined ? `${bssTarget.total_throughput_mbps.toFixed(2)} Mbps` : 'N/A'}</div>
            </div>
            <div className={styles.metricItem}>
              <div className={styles.metricLabel}>Signal Strength</div>
              <div className={styles.metricValue}>{bssTarget.signal_strength !== null ? `${bssTarget.signal_strength} dBm` : 'N/A'}</div>
            </div>
             <div className={styles.metricItem}>
              <div className={styles.metricLabel}>Associated STAs</div>
              <div className={styles.metricValue}>{Object.keys(bssTarget.associated_stas || {}).length}</div>
            </div>
          </>
        )}
        {!isBss && staTarget && (
          <>
            <div className={styles.metricItem}>
              <div className={styles.metricLabel}>Signal Strength</div>
              <div className={styles.metricValue}>{staTarget.signal_strength !== null ? `${staTarget.signal_strength} dBm` : 'N/A'}</div>
            </div>
            <div className={styles.metricItem}>
              <div className={styles.metricLabel}>UL Throughput</div>
              <div className={styles.metricValue}>{staTarget.throughput_ul_mbps !== undefined ? `${staTarget.throughput_ul_mbps.toFixed(2)} Mbps` : 'N/A'}</div>
            </div>
            <div className={styles.metricItem}>
              <div className={styles.metricLabel}>DL Throughput</div>
              <div className={styles.metricValue}>{staTarget.throughput_dl_mbps !== undefined ? `${staTarget.throughput_dl_mbps.toFixed(2)} Mbps` : 'N/A'}</div>
            </div>
             <div className={styles.metricItem}>
              <div className={styles.metricLabel}>TX Bitrate</div>
              <div className={styles.metricValue}>{staTarget.tx_bitrate_mbps !== undefined ? `${staTarget.tx_bitrate_mbps} Mbps` : 'N/A'}</div>
            </div>
             <div className={styles.metricItem}>
              <div className={styles.metricLabel}>RX Bitrate</div>
              <div className={styles.metricValue}>{staTarget.rx_bitrate_mbps !== undefined ? `${staTarget.rx_bitrate_mbps} Mbps` : 'N/A'}</div>
            </div>
          </>
        )}
      </div>

      {isBss && bssTarget?.historical_channel_utilization && bssTarget.historical_channel_utilization.length > 0 && (
        <div className={styles.chartContainer}>
          <h3 className={styles.chartTitle}>Channel Utilization History (%)</h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={bssTarget.historical_channel_utilization}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--color-grid-line)" />
              <XAxis dataKey="timestamp" tickFormatter={formatTimestamp} stroke="var(--color-axis-line)" />
              <YAxis stroke="var(--color-axis-line)" domain={[0, 100]}/>
              <Tooltip 
                contentStyle={{ backgroundColor: 'var(--color-tooltip-background)', border: '1px solid var(--color-tooltip-border)', borderRadius: 'var(--border-radius-small)'}} 
                labelStyle={{ color: 'var(--color-tooltip-label)' }}
                itemStyle={{ color: 'var(--color-tooltip-item)' }}
              />
              <Legend />
              <Line type="monotone" dataKey="value" name="Utilization" stroke="var(--color-accent-blue)" strokeWidth={2} dot={{ r: 2 }} activeDot={{ r: 5 }} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {isBss && bssTarget?.historical_total_throughput && bssTarget.historical_total_throughput.length > 0 && (
        <div className={styles.chartContainer}>
          <h3 className={styles.chartTitle}>Total Throughput History (Mbps)</h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={bssTarget.historical_total_throughput}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--color-grid-line)" />
              <XAxis dataKey="timestamp" tickFormatter={formatTimestamp} stroke="var(--color-axis-line)" />
              <YAxis stroke="var(--color-axis-line)" />
              <Tooltip 
                contentStyle={{ backgroundColor: 'var(--color-tooltip-background)', border: '1px solid var(--color-tooltip-border)', borderRadius: 'var(--border-radius-small)'}} 
                labelStyle={{ color: 'var(--color-tooltip-label)' }}
                itemStyle={{ color: 'var(--color-tooltip-item)' }}
              />
              <Legend />
              <Line type="monotone" dataKey="value" name="Throughput" stroke="var(--color-accent-green)" strokeWidth={2} dot={{ r: 2 }} activeDot={{ r: 5 }} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {!isBss && staTarget?.historical_throughput_ul && staTarget.historical_throughput_ul.length > 0 && (
         <div className={styles.chartContainer}>
          <h3 className={styles.chartTitle}>Uplink Throughput History (Mbps)</h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={staTarget.historical_throughput_ul}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--color-grid-line)" />
              <XAxis dataKey="timestamp" tickFormatter={formatTimestamp} stroke="var(--color-axis-line)" />
              <YAxis stroke="var(--color-axis-line)" />
              <Tooltip 
                contentStyle={{ backgroundColor: 'var(--color-tooltip-background)', border: '1px solid var(--color-tooltip-border)', borderRadius: 'var(--border-radius-small)'}} 
                labelStyle={{ color: 'var(--color-tooltip-label)' }}
                itemStyle={{ color: 'var(--color-tooltip-item)' }}
              />
              <Legend />
              <Line type="monotone" dataKey="value" name="UL Throughput" stroke="var(--color-accent-orange)" strokeWidth={2} dot={{ r: 2 }} activeDot={{ r: 5 }} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {!isBss && staTarget?.historical_throughput_dl && staTarget.historical_throughput_dl.length > 0 && (
         <div className={styles.chartContainer}>
          <h3 className={styles.chartTitle}>Downlink Throughput History (Mbps)</h3>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={staTarget.historical_throughput_dl}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--color-grid-line)" />
              <XAxis dataKey="timestamp" tickFormatter={formatTimestamp} stroke="var(--color-axis-line)" />
              <YAxis stroke="var(--color-axis-line)" />
              <Tooltip 
                contentStyle={{ backgroundColor: 'var(--color-tooltip-background)', border: '1px solid var(--color-tooltip-border)', borderRadius: 'var(--border-radius-small)'}} 
                labelStyle={{ color: 'var(--color-tooltip-label)' }}
                itemStyle={{ color: 'var(--color-tooltip-item)' }}
              />
              <Legend />
              <Line type="monotone" dataKey="value" name="DL Throughput" stroke="var(--color-accent-purple)" strokeWidth={2} dot={{ r: 2 }} activeDot={{ r: 5 }} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
};