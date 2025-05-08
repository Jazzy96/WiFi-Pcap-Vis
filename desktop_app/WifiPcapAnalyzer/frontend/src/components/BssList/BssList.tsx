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

  const cardTitle = `${bss.ssid || '(Hidden)'} (${bss.bssid})`;

  return (
    <Card
      title={cardTitle}
      className={`${styles.bssItem} ${isExpanded ? styles.expanded : ''} ${isSelectedForStaList ? styles.selectedForSta : ''}`}
      onClick={handleItemClick}
      // role="button" // Card itself is not a button, but clickable
      // tabIndex={0} // Consider accessibility for custom clickable elements
      // onKeyDown={(e) => (e.key === 'Enter' || e.key === ' ') && handleItemClick()}
      // aria-expanded={isExpanded}
      // aria-selected={isSelectedForStaList} // This might be better on a specific element if the whole card isn't "selectable" in ARIA terms
    >
      <div className={styles.bssSummary}>
        {/* <div className={`${styles.bssField} ${styles.bssid}`}><strong>BSSID:</strong> {bss.bssid}</div> */}
        {/* <div className={`${styles.bssField} ${styles.ssid}`}><strong>SSID:</strong> {bss.ssid || '(Hidden)'}</div> */}
        <div className={`${styles.bssField} ${styles.signal}`}><strong>Signal:</strong> {bss.signal_strength !== null ? `${bss.signal_strength} dBm` : 'N/A'}</div>
        <div className={`${styles.bssField} ${styles.channel}`}><strong>Ch:</strong> {bss.channel}</div>
        <div className={`${styles.bssField} ${styles.stationsSummary}`}><strong>STAs:</strong> {stationCount}</div>
        <div className={`${styles.bssField} ${styles.channelUtilization}`}><strong>Util:</strong> {bss.channel_utilization_percent !== undefined ? `${bss.channel_utilization_percent.toFixed(1)}%` : 'N/A'}</div>
        <div className={`${styles.bssField} ${styles.throughput}`}><strong>Thrpt:</strong> {bss.total_throughput_mbps !== undefined ? `${bss.total_throughput_mbps.toFixed(2)} Mbps` : 'N/A'}</div>
        {/* Expand indicator removed based on feedback */}
        {/* <span className={`${styles.expandIndicator} ${isExpanded ? styles.expanded : ''}`}>{isExpanded ? '▼' : '▶'}</span> */}
      </div>
      {isExpanded && (
        <div className={styles.bssDetails}>
          <div className={styles.bssField}><strong>Bandwidth:</strong> {bss.bandwidth}</div>
          <div className={`${styles.bssField} ${styles.fullWidthField}`}><strong>Security:</strong> {bss.security}</div>
          <div className={styles.bssField}><strong>Last Seen:</strong> {new Date(bss.last_seen).toLocaleTimeString()}</div>
          {/* Detailed performance metrics for expanded view */}
          <div className={`${styles.bssField} ${styles.fullWidthField}`}><strong>Channel Utilization:</strong> {bss.channel_utilization_percent !== undefined ? `${bss.channel_utilization_percent.toFixed(1)}%` : 'N/A'}</div>
          <div className={`${styles.bssField} ${styles.fullWidthField}`}><strong>Total Throughput:</strong> {bss.total_throughput_mbps !== undefined ? `${bss.total_throughput_mbps.toFixed(2)} Mbps` : 'N/A'}</div>
          
          {bss.ht_capabilities && (
            <div className={`${styles.bssField} ${styles.htCaps}`}>
              <strong>HT Cap: </strong>
              {bss.ht_capabilities.channel_width_40mhz ? '40MHz, ' : '20MHz, '}
              {bss.ht_capabilities.short_gi_20mhz ? 'SGI_20 ' : ''}
              {bss.ht_capabilities.short_gi_40mhz ? 'SGI_40 ' : ''}
              {/* MCS: {bss.ht_capabilities.supported_mcs_set} */}
            </div>
          )}
          {bss.vht_capabilities && (
            <div className={`${styles.bssField} ${styles.vhtCaps}`}>
              <strong>VHT Cap: </strong>
              {bss.vht_capabilities.channel_width_160mhz ? '160MHz, ' : (bss.vht_capabilities.channel_width_80plus80mhz ? '80+80MHz, ' : (bss.vht_capabilities.channel_width_80mhz ? '80MHz, ' : ''))}
              {bss.vht_capabilities.short_gi_80mhz ? 'SGI_80 ' : ''}
              {bss.vht_capabilities.short_gi_160mhz ? 'SGI_160 ' : ''}
              {/* SU Beamformer: {bss.vht_capabilities.su_beamformer_capable ? 'Yes' : 'No'}, MU Beamformer: {bss.vht_capabilities.mu_beamformer_capable ? 'Yes' : 'No'} */}
            </div>
          )}
        </div>
      )}
    </Card>
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