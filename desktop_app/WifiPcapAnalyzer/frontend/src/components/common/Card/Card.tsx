import React, { HTMLAttributes, ReactNode } from 'react';
import styles from './Card.module.css';

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  title?: string;
  // Add other props like 'footerContent', 'headerActions', etc. as needed
}

const Card: React.FC<CardProps> = ({
  children,
  title,
  className,
  ...props
}) => {
  const cardClasses = `
    ${styles.card}
    ${className || ''}
  `;

  return (
    <div className={cardClasses.trim()} {...props}>
      {title && <h3 className={styles.cardTitle}>{title}</h3>}
      <div className={styles.cardContent}>
        {children}
      </div>
    </div>
  );
};

export default Card;