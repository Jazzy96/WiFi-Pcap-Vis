import React, { SVGProps } from 'react';
import styles from './Icon.module.css';

// It's good practice to define the names of icons you'll support.
// This can be expanded as you add more SVG files to the icons/ directory.
export type IconName = 'close' | 'settings' | 'info' | 'warning' | 'error' | 'placeholder'; // Add more as needed

interface IconProps extends SVGProps<SVGSVGElement> {
  name: IconName;
  size?: number | string;
  // 'color' prop can be used if you want to override CSS 'currentColor' behavior
  // color?: string; 
}

const Icon: React.FC<IconProps> = ({
  name,
  size = 24, // Default size
  className,
  style,
  ...props
}) => {
  // This is a dynamic import. For this to work well with bundlers like Webpack/Vite
  // and for SVGR to transform them into React components, specific configuration
  // might be needed in vite.config.ts or webpack.config.js.
  // For Vite with SVGR: `import { ReactComponent as IconName } from './icons/icon-name.svg';`
  // However, a simpler approach for now is to manually create a map or switch statement.

  // Placeholder for actual icon components.
  // These would ideally be dynamically imported or mapped from actual SVG components.
  const IconsMap: Record<IconName, React.FC<SVGProps<SVGSVGElement>>> = {
    close: (svgProps) => ( // Example: Close Icon (X)
      <svg viewBox="0 0 24 24" fill="currentColor" {...svgProps}>
        <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z" />
      </svg>
    ),
    settings: (svgProps) => ( // Example: Settings Icon (Gear)
      <svg viewBox="0 0 24 24" fill="currentColor" {...svgProps}>
        <path d="M19.43 12.98c.04-.32.07-.64.07-.98s-.03-.66-.07-.98l2.11-1.65c.19-.15.24-.42.12-.64l-2-3.46c-.12-.22-.39-.3-.61-.22l-2.49 1c-.52-.4-1.08-.73-1.69-.98l-.38-2.65C14.46 2.18 14.25 2 14 2h-4c-.25 0-.46.18-.49.42l-.38 2.65c-.61.25-1.17.59-1.69.98l-2.49-1c-.23-.09-.49 0-.61.22l-2 3.46c-.13.22-.07.49.12.64l2.11 1.65c-.04.32-.07.65-.07.98s.03.66.07.98l-2.11 1.65c-.19.15-.24.42-.12.64l2 3.46c.12.22.39.3.61.22l2.49-1c.52.4 1.08.73 1.69.98l.38 2.65c.03.24.24.42.49.42h4c.25 0 .46-.18.49-.42l.38-2.65c.61-.25 1.17-.59 1.69-.98l2.49 1c.23.09.49 0 .61.22l2-3.46c.12-.22.07-.49-.12-.64l-2.11-1.65zM12 15.5c-1.93 0-3.5-1.57-3.5-3.5s1.57-3.5 3.5-3.5 3.5 1.57 3.5 3.5-1.57 3.5-3.5 3.5z" />
      </svg>
    ),
    // Add other icons here as React components
    info: (svgProps) => <svg {...svgProps}><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/></svg>,
    warning: (svgProps) => <svg {...svgProps}><path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/></svg>,
    error: (svgProps) => <svg {...svgProps}><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v4z"/></svg>, // Placeholder, same as info for now
    placeholder: (svgProps) => ( // A simple placeholder
      <svg viewBox="0 0 24 24" fill="currentColor" {...svgProps}>
        <rect width="18" height="18" x="3" y="3" rx="2" ry="2" stroke="currentColor" strokeWidth="2" fill="none" />
      </svg>
    ),
  };

  const SelectedIcon = IconsMap[name] || IconsMap.placeholder;

  return (
    <SelectedIcon
      className={`${styles.icon} ${className || ''}`.trim()}
      style={{ width: size, height: size, ...style }}
      {...props}
    />
  );
};

export default Icon;