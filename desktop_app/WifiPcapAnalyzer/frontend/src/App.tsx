import React from 'react';
import './App.css';
import { DataProvider, useAppState } from './contexts/DataContext'; // Import useAppState
import { ControlPanel } from './components/ControlPanel/ControlPanel';
import { BssList } from './components/BssList/BssList';
import { StaList } from './components/StaList/StaList'; // Import StaList
import { PerformanceDetailPanel } from './components/PerformanceDetailPanel/PerformanceDetailPanel'; // Import PerformanceDetailPanel

// InnerApp component to access context after DataProvider is set up
const InnerApp: React.FC = () => {
  const { selectedPerformanceTarget } = useAppState(); // 移除 isPanelCollapsed

  // 判断是否显示性能面板
  const performancePanelVisible = !!selectedPerformanceTarget;
  
  return (
    <div className="App">
      {/* Header removed based on feedback */}
      <main className={`App-main ${!performancePanelVisible ? 'no-performance-panel' : ''}`}>
        <div className="control-panel-container">
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
