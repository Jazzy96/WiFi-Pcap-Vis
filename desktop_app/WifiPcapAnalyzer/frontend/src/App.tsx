import React from 'react';
import './App.css';
import { DataProvider, useAppState } from './contexts/DataContext'; // Import useAppState
import { ControlPanel } from './components/ControlPanel/ControlPanel';
import { BssList } from './components/BssList/BssList';
import { StaList } from './components/StaList/StaList'; // Import StaList

// InnerApp component to access context after DataProvider is set up
const InnerApp: React.FC = () => {
  const { isPanelCollapsed } = useAppState(); // Get panel collapse state

  return (
    <div className="App">
      <header className="App-header">
        <h1>WiFi PCAP Visualizer</h1>
      </header>
      <main className="App-main">
        <div className={`control-panel-container ${isPanelCollapsed ? 'collapsed' : ''}`}>
          <ControlPanel />
        </div>
        <div className="bss-list-container">
            <BssList />
          </div>
          <div className="main-content-area">
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
