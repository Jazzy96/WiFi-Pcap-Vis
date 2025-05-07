import React from 'react';
import { useAppState } from '../../contexts/DataContext';
import { STA, BSS } from '../../types/data';
import './StaList.css'; // We will create this CSS file next

interface StaListItemProps {
  sta: STA;
}

const StaListItemDetails: React.FC<StaListItemProps> = ({ sta }) => {
  return (
    <div className="sta-list-item-details">
      <div className="sta-field mac"><strong>MAC:</strong> {sta.mac_address}</div>
      <div className="sta-field signal"><strong>Signal:</strong> {sta.signal_strength !== null ? `${sta.signal_strength} dBm` : 'N/A'}</div>
      <div className="sta-field last-seen"><strong>Last Seen:</strong> {new Date(sta.last_seen).toLocaleTimeString()}</div>
      {/* Add other STA details as needed, e.g., capabilities */}
      {sta.ht_capabilities && <div className="sta-field ht-caps"><strong>HT Caps:</strong> {JSON.stringify(sta.ht_capabilities)}</div>}
      {sta.vht_capabilities && <div className="sta-field vht-caps"><strong>VHT Caps:</strong> {JSON.stringify(sta.vht_capabilities)}</div>}
    </div>
  );
};

export const StaList: React.FC = () => {
  const { bssList, selectedBssidForStaList, staList: allStas } = useAppState();

  // Handle the initial state where no BSS is selected
  if (!selectedBssidForStaList) {
    return (
      <div className="sta-list-container">
        <h2>Associated Stations</h2>
        <p className="sta-list-status">Select a BSS from the list to see its associated stations.</p>
      </div>
    );
  }

  // Find the selected BSS *after* confirming selectedBssidForStaList is not null
  const selectedBss = bssList.find(bss => bss.bssid === selectedBssidForStaList);

  // Handle the case where the selected BSSID exists but the BSS object is not found
  // (This might happen briefly during updates or if data is inconsistent)
  if (!selectedBss) {
     return (
      <div className="sta-list-container">
        <h2>Associated Stations</h2>
        {/* Display the BSSID correctly */}
        <p className="sta-list-status">Selected BSS (<code>{selectedBssidForStaList}</code>) not found in current data.</p>
      </div>
    );
  }

  // Now we know selectedBss exists
  const stationsToShow: STA[] = Object.values(selectedBss.associated_stas || {});
  const title = `Stations in ${selectedBss.ssid || selectedBss.bssid} (${stationsToShow.length})`;


  if (stationsToShow.length === 0) {
    return (
      <div className="sta-list-container">
        <h2>{title}</h2>
        <p className="sta-list-status">No stations currently associated with this BSS.</p>
      </div>
    );
  }

  return (
    <div className="sta-list-container">
      <h2>{title}</h2>
      <div className="sta-list">
        {stationsToShow.map(sta => (
          <StaListItemDetails key={sta.mac_address} sta={sta} />
        ))}
      </div>
    </div>
  );
};