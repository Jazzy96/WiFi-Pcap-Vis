import React from 'react';
import './App.css';
import { DataProvider, useAppState } from './contexts/DataContext'; // Import useAppState
import { ControlPanel } from './components/ControlPanel/ControlPanel';
import { BssList } from './components/BssList/BssList';
import { StaList } from './components/StaList/StaList'; // Import StaList

// InnerApp component to access context after DataProvider is set up
const InnerApp: React.FC = () => {
  const { isPanelCollapsed } = useAppState(); // Get panel collapse state

  const mainStyle = {
    gridTemplateColumns: isPanelCollapsed ? '60px 2fr 3fr' : 'minmax(240px, 0.8fr) 2fr 3fr',
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
