import React, { HTMLAttributes, ReactNode } from 'react';
import styles from './Table.module.css';

interface TableColumn<T> {
  key: keyof T | 'actions'; // 'actions' for a column with action buttons
  header: ReactNode;
  render: (item: T) => ReactNode;
  width?: string;
}

interface TableProps<T> extends HTMLAttributes<HTMLTableElement> {
  columns: TableColumn<T>[];
  data: T[];
  // Add other props like 'onRowClick', 'isLoading', 'emptyMessage' etc.
  caption?: string;
}

const Table = <T extends {}>({
  columns,
  data,
  className,
  caption,
  ...props
}: TableProps<T>) => {
  const tableClasses = `
    ${styles.table}
    ${className || ''}
  `;

  return (
    <div className={styles.tableContainer}>
      <table className={tableClasses.trim()} {...props}>
        {caption && <caption className={styles.tableCaption}>{caption}</caption>}
        <thead className={styles.tableHead}>
          <tr>
            {columns.map((col) => (
              <th key={String(col.key)} style={{ width: col.width }}>
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className={styles.tableBody}>
          {data.length === 0 ? (
            <tr>
              <td colSpan={columns.length} className={styles.emptyMessage}>
                No data available.
              </td>
            </tr>
          ) : (
            data.map((item, rowIndex) => (
              <tr key={rowIndex}>
                {columns.map((col) => (
                  <td key={`${String(col.key)}-${rowIndex}`}>
                    {col.render(item)}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
};

export default Table;