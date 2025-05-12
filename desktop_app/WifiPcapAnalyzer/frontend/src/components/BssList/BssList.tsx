import React, { useState } from 'react';
import { useAppState, useAppDispatch } from '../../contexts/DataContext';
import { BSS } from '../../types/data';
import styles from './BssList.module.css'; // Import CSS Modules
import Card from '../common/Card/Card'; // Import Card component
// import Button from '../common/Button/Button'; // If needed for actions within the card

interface BssItemProps {
  bss: BSS;
  isSelectedForStaList: boolean;
  isExpanded: boolean;
  onToggleExpand: (bssid: string) => void;
}

const BssItem: React.FC<BssItemProps> = ({ bss, isSelectedForStaList, isExpanded, onToggleExpand }) => {
  const dispatch = useAppDispatch();
  const stationCount = Object.keys(bss.associated_stas || {}).length;

  const handleItemClick = () => {
    onToggleExpand(bss.bssid);
    dispatch({ type: 'SET_SELECTED_BSSID_FOR_STA_LIST', payload: bss.bssid });
    // Also set this BSS as the selected performance target
    dispatch({ type: 'SET_SELECTED_PERFORMANCE_TARGET', payload: { type: 'bss', id: bss.bssid } });
  };

  return (
    <div 
      className={`${styles.bssItem} ${isExpanded ? styles.expanded : ''} ${isSelectedForStaList ? styles.selectedForSta : ''}`}
      onClick={handleItemClick}
    >
      {/* BSS头部信息 - BSSID和SSID */}
      <div className={styles.bssHeader}>
        <div className={styles.bssId}>{bss.bssid}</div>
        <div className={styles.bssSsid}>{bss.ssid || '(Hidden)'}</div>
      </div>
      
      {/* BSS主要信息 - 表格式网格布局 */}
      <div className={styles.bssInfoGrid}>
        {/* 信号强度 */}
        <div className={styles.bssGridCell}>
          <div className={styles.bssGridLabel}>Signal</div>
          <div className={styles.bssGridValue}>{bss.signal_strength !== null ? `${bss.signal_strength} dBm` : 'N/A'}</div>
        </div>
        
        {/* 信道 */}
        <div className={styles.bssGridCell}>
          <div className={styles.bssGridLabel}>Channel</div>
          <div className={styles.bssGridValue}>{bss.channel}</div>
        </div>
        
        {/* 关联的STA数 */}
        <div className={styles.bssGridCell}>
          <div className={styles.bssGridLabel}>STAs</div>
          <div className={styles.bssGridValue}>{stationCount}</div>
        </div>
        
        {/* 信道利用率 */}
        <div className={styles.bssGridCell}>
          <div className={styles.bssGridLabel}>Utilization</div>
          <div className={styles.bssGridValue}>{bss.util !== undefined ? `${bss.util.toFixed(1)}%` : 'N/A'}</div>
        </div>
        
        {/* 吞吐量 */}
        <div className={styles.bssGridCell}>
          <div className={styles.bssGridLabel}>Throughput</div>
          <div className={styles.bssGridValue}>{bss.thrpt !== undefined ? `${(bss.thrpt/1000000).toFixed(2)} Mbps` : 'N/A'}</div>
        </div>
        
        {/* 最后一次见到的时间 */}
        <div className={styles.lastSeenRow}>
          <strong>Last Seen:</strong> {new Date(bss.last_seen).toLocaleTimeString()}
        </div>
      </div>
      
      {/* 展开时显示详细信息 */}
      {isExpanded && (
        <div className={styles.bssDetails}>
          <div className={styles.bssDetailsGrid}>
            <div className={styles.bssDetailRow}>
              <div className={styles.bssDetailLabel}>Bandwidth:</div>
              <div className={styles.bssDetailValue}>{bss.bandwidth}</div>
            </div>
            
            <div className={styles.bssDetailRow}>
              <div className={styles.bssDetailLabel}>Security:</div>
              <div className={styles.bssDetailValue}>
                {bss.security === "WPA2-PSK" ? "WPA2 Personal" :
                 bss.security === "WPA3-SAE" ? "WPA3 Personal" :
                 bss.security === "RSN/WPA2/WPA3" ? "WPA2/WPA3 Mixed" :
                 bss.security === "Open" ? "Open Network" :
                 bss.security === "WEP" ? "WEP (Insecure)" :
                 `${bss.security} (Raw Value)`}
                <span style={{color: 'gray', fontSize: '0.8em', marginLeft: '4px'}}>[{bss.security}]</span>
              </div>
            </div>
            
            <div className={styles.bssDetailRow}>
              <div className={styles.bssDetailLabel}>Ch. Util:</div>
              <div className={styles.bssDetailValue}>{bss.util !== undefined ? `${bss.util.toFixed(1)}%` : 'N/A'}</div>
            </div>
            
            <div className={styles.bssDetailRow}>
              <div className={styles.bssDetailLabel}>Throughput:</div>
              <div className={styles.bssDetailValue}>{bss.thrpt !== undefined ? `${(bss.thrpt/1000000).toFixed(2)} Mbps` : 'N/A'}</div>
            </div>
            
            {/* HT能力 */}
            {bss.ht_capabilities && (
              <div className={styles.bssCapabilities}>
                <strong>HT Capabilities: </strong>
                {bss.ht_capabilities.channel_width_40mhz ? '40MHz, ' : '20MHz, '}
                {bss.ht_capabilities.short_gi_20mhz ? 'SGI_20 ' : ''}
                {bss.ht_capabilities.short_gi_40mhz ? 'SGI_40 ' : ''}
              </div>
            )}
            
            {/* VHT能力 */}
            {bss.vht_capabilities && (
              <div className={styles.bssCapabilities}>
                <strong>VHT Capabilities: </strong>
                {bss.vht_capabilities.channel_width_160mhz ? '160MHz, ' : 
                 (bss.vht_capabilities.channel_width_80plus80mhz ? '80+80MHz, ' : 
                 (bss.vht_capabilities.channel_width_80mhz ? '80MHz, ' : ''))}
                {bss.vht_capabilities.short_gi_80mhz ? 'SGI_80 ' : ''}
                {bss.vht_capabilities.short_gi_160mhz ? 'SGI_160 ' : ''}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export const BssList: React.FC = () => {
  const appState = useAppState();
  const [expandedBssid, setExpandedBssid] = useState<string | null>(null);

  if (!appState || typeof appState.bssList === 'undefined') {
    return <div className={styles.bssListStatus}>Initializing data context...</div>;
  }

  const { bssList, selectedBssidForStaList } = appState;

  if (bssList.length === 0) {
    return <div className={styles.bssListStatus}>No BSS data available.</div>;
  }

  const sortedBssList = [...bssList].sort((a, b) => {
    const signalA = a.signal_strength ?? -Infinity;
    const signalB = b.signal_strength ?? -Infinity;
    if (signalB !== signalA) {
      return signalB - signalA;
    }
    const staCountA = Object.keys(a.associated_stas || {}).length;
    const staCountB = Object.keys(b.associated_stas || {}).length;
    if (staCountB !== staCountA) {
      return staCountB - staCountA;
    }
    return a.bssid.localeCompare(b.bssid); // Secondary sort by BSSID for stability
  });

  const handleToggleExpand = (bssid: string) => {
    setExpandedBssid(prev => (prev === bssid ? null : bssid));
  };

  return (
    <div className={styles.bssListWrapperInternal}>
      <h2 className={styles.bssListTitle}>BSS List ({sortedBssList.length})</h2>
      <div className={styles.bssList}>
        {sortedBssList.map((bss) => (
          <BssItem
            key={bss.bssid}
            bss={bss}
            isSelectedForStaList={bss.bssid === selectedBssidForStaList}
            isExpanded={bss.bssid === expandedBssid}
            onToggleExpand={handleToggleExpand}
          />
        ))}
      </div>
    </div>
  );
};