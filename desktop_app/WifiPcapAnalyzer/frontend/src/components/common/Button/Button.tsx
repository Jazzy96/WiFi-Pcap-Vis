import React, { ButtonHTMLAttributes, ReactNode } from 'react';
import styles from './Button.module.css';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  variant?: 'primary' | 'secondary' | 'none';
  // Add other props like 'size', 'iconOnly', etc. as needed
}

const Button: React.FC<ButtonProps> = ({
  children,
  variant = 'primary',
  className,
  ...props
}) => {
  const buttonClasses = `
    ${styles.button}
    ${variant !== 'none' ? styles[variant] : ''}
    ${className || ''}
  `;

  return (
    <button className={buttonClasses.trim()} {...props}>
      {children}
    </button>
  );
};

export default Button;