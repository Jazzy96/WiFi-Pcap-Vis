import React from 'react';
import './App.css';
import { DataProvider, useAppState } from './contexts/DataContext'; // Import useAppState
import { ControlPanel } from './components/ControlPanel/ControlPanel';
import { BssList } from './components/BssList/BssList';
import { StaList } from './components/StaList/StaList'; // Import StaList
import { PerformanceDetailPanel } from './components/PerformanceDetailPanel/PerformanceDetailPanel'; // Import PerformanceDetailPanel

// InnerApp component to access context after DataProvider is set up
const InnerApp: React.FC = () => {
  const { isPanelCollapsed, selectedPerformanceTarget } = useAppState(); // Get panel collapse state and selectedPerformanceTarget

  // Adjust grid based on whether the performance panel should be shown
  const performancePanelVisible = !!selectedPerformanceTarget;
  let gridTemplateColumnsValue = '';

  if (isPanelCollapsed) {
    gridTemplateColumnsValue = performancePanelVisible ? '60px 1.5fr 2fr 2.5fr' : '60px 2fr 3fr';
  } else {
    gridTemplateColumnsValue = performancePanelVisible ? 'minmax(240px, 0.6fr) 1.5fr 2fr 2.5fr' : 'minmax(240px, 0.8fr) 2fr 3fr';
  }
  
  const mainStyle = {
    gridTemplateColumns: gridTemplateColumnsValue,
  };

  return (
    <div className="App">
      {/* Header removed based on feedback */}
      <main className="App-main" style={mainStyle}>
        <div className={`control-panel-container ${isPanelCollapsed ? 'collapsed' : ''}`}>
          <ControlPanel />
        </div>
        <div className="bss-list-container">
            <BssList />
          </div>
          <div className="sta-list-container-wrapper"> {/* Ensure class name matches App.css */}
            <StaList /> {/* Render StaList here */}
          </div>
          {performancePanelVisible && (
            <div className="performance-detail-panel-container">
              <PerformanceDetailPanel />
            </div>
          )}
        </main>
      </div>
  );
};

function App() {
  return (
    <DataProvider>
      <InnerApp />
    </DataProvider>
  );
}

export default App;
