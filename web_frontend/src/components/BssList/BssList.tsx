import React, { useState } from 'react';
import { useAppState, useAppDispatch } from '../../contexts/DataContext'; // Import useAppDispatch
import { BSS } from '../../types/data';
import './BssList.css';

interface BssItemProps {
  bss: BSS;
  isSelectedForStaList: boolean;
  isExpanded: boolean; // Controlled from parent
  onToggleExpand: (bssid: string) => void; // Function to toggle expansion in parent
}

const BssItem: React.FC<BssItemProps> = ({ bss, isSelectedForStaList, isExpanded, onToggleExpand }) => {
  // const [isExpanded, setIsExpanded] = useState(false); // State moved to parent
  const dispatch = useAppDispatch();

  const stationCount = Object.keys(bss.associated_stas || {}).length;

  const handleItemClick = () => {
    onToggleExpand(bss.bssid); // Tell parent to toggle expansion for this BSSID
    // Also set this BSSID for STA list display when clicked
    dispatch({ type: 'SET_SELECTED_BSSID_FOR_STA_LIST', payload: bss.bssid });
  };

  return (
    <div
      className={`bss-item ${isExpanded ? 'expanded' : ''} ${isSelectedForStaList ? 'selected-for-sta' : ''}`}
      onClick={handleItemClick}
    >
      <div className="bss-summary">
        <div className="bss-field bssid"><strong>BSSID:</strong> {bss.bssid}</div>
        <div className="bss-field ssid"><strong>SSID:</strong> {bss.ssid || '(Hidden)'}</div>
        <div className="bss-field signal"><strong>Signal:</strong> {bss.signal_strength !== null ? `${bss.signal_strength} dBm` : 'N/A'}</div>
        <div className="bss-field channel"><strong>Ch:</strong> {bss.channel}</div>
        <div className="bss-field stations-summary"><strong>STAs:</strong> {stationCount}</div>
        <span className={`expand-indicator ${isExpanded ? 'expanded' : ''}`}>{isExpanded ? '▼' : '▶'}</span>
      </div>
      {isExpanded && (
        <div className="bss-details">
          <div className="bss-field bandwidth"><strong>Bandwidth:</strong> {bss.bandwidth}</div>
          <div className="bss-field security"><strong>Security:</strong> {bss.security}</div>
          <div className="bss-field last-seen"><strong>Last Seen:</strong> {new Date(bss.last_seen).toLocaleTimeString()}</div>
          
          {bss.ht_capabilities && (
            <div className="bss-field ht-caps">
              <strong>HT Cap: </strong>
              {bss.ht_capabilities.channel_width_40mhz ? '40MHz, ' : '20MHz, '}
              {bss.ht_capabilities.short_gi_20mhz ? 'SGI_20 ' : ''}
              {bss.ht_capabilities.short_gi_40mhz ? 'SGI_40 ' : ''}
              {/* Consider adding MCS set summary if needed, e.g., Max MCS */}
            </div>
          )}
          {bss.vht_capabilities && (
            <div className="bss-field vht-caps">
              <strong>VHT Cap: </strong>
              {bss.vht_capabilities.channel_width_160mhz ? '160MHz, ' : (bss.vht_capabilities.channel_width_80plus80mhz ? '80+80MHz, ' : (bss.vht_capabilities.channel_width_80mhz ? '80MHz, ' : ''))}
              {bss.vht_capabilities.short_gi_80mhz ? 'SGI_80 ' : ''}
              {bss.vht_capabilities.short_gi_160mhz ? 'SGI_160 ' : ''}
              {/* Consider adding MU-MIMO, SU/MU Beamformer status if important */}
            </div>
          )}
          {/* Removed STA list rendering from here */}
        </div>
      )}
    </div>
  );
};

export const BssList: React.FC = () => {
  const appState = useAppState();
  const [expandedBssid, setExpandedBssid] = useState<string | null>(null); // State for the single expanded BSS

  if (!appState || typeof appState.bssList === 'undefined') {
    return <div className="bss-list-status">Initializing data context...</div>;
  }

  const { bssList, isConnected, selectedBssidForStaList } = appState;

  if (!isConnected && bssList.length === 0) {
    return <div className="bss-list-status">Connecting to WebSocket or no data received yet...</div>;
  }

  if (bssList.length === 0) {
    return <div className="bss-list-status">No BSS data available.</div>;
  }

  const sortedBssList = [...bssList].sort((a, b) => {
    const signalA = a.signal_strength ?? -Infinity;
    const signalB = b.signal_strength ?? -Infinity;
    if (signalB !== signalA) {
      return signalB - signalA;
    }
    const staCountA = Object.keys(a.associated_stas || {}).length;
    const staCountB = Object.keys(b.associated_stas || {}).length;
    return staCountB - staCountA;
  });

  const handleToggleExpand = (bssid: string) => {
    setExpandedBssid(prev => (prev === bssid ? null : bssid)); // Toggle: if same, collapse (null), else expand new one
  };

  return (
    <div className="bss-list-wrapper-internal">
      <h2>BSS List ({sortedBssList.length})</h2>
      <div className="bss-list">
        {sortedBssList.map((bss) => (
          <BssItem
            key={bss.bssid}
            bss={bss}
            isSelectedForStaList={bss.bssid === selectedBssidForStaList}
            isExpanded={bss.bssid === expandedBssid} // Pass expanded state down
            onToggleExpand={handleToggleExpand} // Pass toggle function down
          />
        ))}
      </div>
    </div>
  );
};