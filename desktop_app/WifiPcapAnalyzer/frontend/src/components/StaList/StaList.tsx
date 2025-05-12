import React from 'react';
import { useAppState, useAppDispatch } from '../../contexts/DataContext'; // Import useAppDispatch
import { STA } from '../../types/data';
import styles from './StaList.module.css'; // Import CSS Modules
import Card from '../common/Card/Card'; // Import Card component

// Helper to format capabilities for display, can be moved to a utils file if used elsewhere
const formatCapabilities = (caps: any) => {
  if (!caps) return 'N/A';
  // Basic JSON stringify, can be enhanced for better readability or specific formatting
  // Using <pre> for basic formatting
  return <pre className={styles.capabilitiesPre}>{JSON.stringify(caps, null, 2)}</pre>;
};

interface StaCardItemProps {
  sta: STA;
  onClick: () => void; // Add onClick handler for selecting performance target
  isSelectedForPerformance: boolean; // To highlight if selected
}

// Component to render a single STA within a Card
const StaCardItem: React.FC<StaCardItemProps> = ({ sta, onClick, isSelectedForPerformance }) => {
  return (
    <div
      className={`${styles.staCard} ${isSelectedForPerformance ? styles.selectedForPerformance : ''}`}
      onClick={onClick}
    >
      {/* STA MAC地址 */}
      <div className={styles.staMacAddress}>{sta.mac_address}</div>
      
      {/* STA主要信息网格 */}
      <div className={styles.staInfoGrid}>
        {/* 信号强度 */}
        <div className={styles.staGridCell}>
          <div className={styles.staGridLabel}>Signal</div>
          <div className={styles.staGridValue}>{sta.signal_strength !== null ? `${sta.signal_strength} dBm` : 'N/A'}</div>
        </div>
        
        {/* 信道利用率 */}
        <div className={styles.staGridCell}>
          <div className={styles.staGridLabel}>Util</div>
          <div className={styles.staGridValue}>{sta.util !== undefined ? `${sta.util.toFixed(1)}%` : 'N/A'}</div>
        </div>
        
        {/* 吞吐量 */}
        <div className={styles.staGridCell}>
          <div className={styles.staGridLabel}>Throughput</div>
          <div className={styles.staGridValue}>{sta.thrpt !== undefined ? `${(sta.thrpt/1000000).toFixed(2)} Mbps` : 'N/A'}</div>
        </div>
        
        {/* 比特率 */}
        <div className={styles.staGridCell}>
          <div className={styles.staGridLabel}>Bitrate</div>
          <div className={styles.staGridValue}>{sta.bitrate !== undefined ? `${sta.bitrate.toFixed(1)} Mbps` : 'N/A'}</div>
        </div>
        
        {/* 最后一次见到的时间 */}
        <div className={styles.lastSeenRow}>
          <strong>Last Seen:</strong> {new Date(sta.last_seen).toLocaleTimeString()}
        </div>
        
        {/* HT能力 */}
        {sta.ht_capabilities && (
          <div className={styles.staCapabilities}>
            <strong>HT Capabilities:</strong>
            {formatCapabilities(sta.ht_capabilities)}
          </div>
        )}
        
        {/* VHT能力 */}
        {sta.vht_capabilities && (
          <div className={styles.staCapabilities}>
            <strong>VHT Capabilities:</strong>
            {formatCapabilities(sta.vht_capabilities)}
          </div>
        )}
      </div>
    </div>
  );
};


export const StaList: React.FC = () => {
  const { bssList, selectedBssidForStaList, selectedPerformanceTarget } = useAppState();
  const dispatch = useAppDispatch();

  if (!selectedBssidForStaList) {
    return (
      <div className={styles.staListContainer}>
        <h2 className={styles.staListTitle}>Associated Stations</h2>
        <p className={styles.staListStatus}>Select a BSS from the list to see its associated stations.</p>
      </div>
    );
  }

  const selectedBss = bssList.find(bss => bss.bssid === selectedBssidForStaList);

  if (!selectedBss) {
     return (
      <div className={styles.staListContainer}>
        <h2 className={styles.staListTitle}>Associated Stations</h2>
        <p className={styles.staListStatus}>
          Selected BSS (<code>{selectedBssidForStaList}</code>) not found.
        </p>
      </div>
    );
  }

  const stationsToShow: STA[] = Object.values(selectedBss.associated_stas || {});
  const titleText = `Stations of ${selectedBss.ssid || selectedBss.bssid} (${stationsToShow.length})`;

  const handleStaClick = (staMac: string) => {
    dispatch({ type: 'SET_SELECTED_PERFORMANCE_TARGET', payload: { type: 'sta', id: staMac } });
  };

  return (
    <div className={styles.staListContainer}>
      <h2 className={styles.staListTitle}>{titleText}</h2>
      
      {stationsToShow.length === 0 ? (
        <p className={styles.staListStatus}>No stations currently associated with this BSS.</p>
      ) : (
        <div className={styles.staList}> {/* This div will contain the list of cards */}
          {stationsToShow.map(sta => (
            <StaCardItem
              key={sta.mac_address}
              sta={sta}
              onClick={() => handleStaClick(sta.mac_address)}
              isSelectedForPerformance={selectedPerformanceTarget?.type === 'sta' && selectedPerformanceTarget.id === sta.mac_address}
            />
          ))}
        </div>
      )}
    </div>
  );
};