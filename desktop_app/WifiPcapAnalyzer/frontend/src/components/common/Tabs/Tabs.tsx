import React, { useState, ReactNode } from 'react';
import styles from './Tabs.module.css';

interface TabProps {
  label: string;
  children: ReactNode;
  disabled?: boolean;
}

interface TabsProps {
  children: React.ReactElement<TabProps>[] | React.ReactElement<TabProps>;
  defaultActiveLabel?: string;
  onTabChange?: (label: string) => void;
}

const Tab: React.FC<TabProps> = ({ children }) => {
  return <div className={styles.tabPanel}>{children}</div>;
};

const Tabs: React.FC<TabsProps> = ({ children, defaultActiveLabel, onTabChange }) => {
  const tabsArray = React.Children.toArray(children) as React.ReactElement<TabProps>[];
  const [activeTab, setActiveTab] = useState(defaultActiveLabel || (tabsArray.length > 0 ? tabsArray[0].props.label : ''));

  const handleTabClick = (label: string, disabled?: boolean) => {
    if (disabled) return;
    setActiveTab(label);
    if (onTabChange) {
      onTabChange(label);
    }
  };

  return (
    <div className={styles.tabsContainer}>
      <div className={styles.tabList} role="tablist">
        {tabsArray.map((tab) => (
          <button
            key={tab.props.label}
            role="tab"
            aria-selected={activeTab === tab.props.label}
            aria-disabled={tab.props.disabled}
            disabled={tab.props.disabled}
            onClick={() => handleTabClick(tab.props.label, tab.props.disabled)}
            className={`${styles.tabButton} ${activeTab === tab.props.label ? styles.active : ''} ${tab.props.disabled ? styles.disabled : ''}`}
          >
            {tab.props.label}
          </button>
        ))}
      </div>
      {tabsArray.map((tab) => {
        if (tab.props.label === activeTab) {
          return React.cloneElement(tab, { key: tab.props.label });
        }
        return null;
      })}
    </div>
  );
};

export { Tabs, Tab };