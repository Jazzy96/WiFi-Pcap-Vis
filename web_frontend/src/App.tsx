import React from 'react';
import './App.css';
import { DataProvider } from './contexts/DataContext';
import { ControlPanel } from './components/ControlPanel/ControlPanel';
import { BssList } from './components/BssList/BssList';
import { StaList } from './components/StaList/StaList'; // Import StaList

function App() {
  return (
    <DataProvider>
      <div className="App">
        <header className="App-header">
          <h1>WiFi PCAP Visualizer</h1>
        </header>
        <main className="App-main">
          <div className="control-panel-container">
            <ControlPanel />
          </div>
          <div className="bss-list-container">
            <BssList />
          </div>
          <div className="main-content-area">
            <StaList /> {/* Render StaList here */}
          </div>
        </main>
        <footer className="App-footer">
          <p>Real-time 802.11 Data Visualization</p>
        </footer>
      </div>
    </DataProvider>
  );
}

export default App;
