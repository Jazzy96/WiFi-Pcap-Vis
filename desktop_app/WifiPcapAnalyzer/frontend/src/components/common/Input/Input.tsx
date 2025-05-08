import React, { InputHTMLAttributes, SelectHTMLAttributes } from 'react';
import styles from './Input.module.css';

interface CommonInputProps {
  label?: string;
  error?: string;
  containerClassName?: string;
}

interface TextInputProps extends InputHTMLAttributes<HTMLInputElement>, CommonInputProps {
  type: 'text' | 'number' | 'password' | 'email'; // Add other relevant input types
}

interface SelectInputProps extends SelectHTMLAttributes<HTMLSelectElement>, CommonInputProps {
  type: 'select';
  options: { value: string | number; label: string }[];
}

type InputProps = TextInputProps | SelectInputProps;

const Input: React.FC<InputProps> = (props) => {
  const { label, error, containerClassName, className, ...rest } = props;

  const inputClasses = `
    ${styles.input}
    ${error ? styles.errorInput : ''}
    ${className || ''}
  `;

  const renderInput = () => {
    if (props.type === 'select') {
      const { options, ...selectProps } = props as SelectInputProps;
      return (
        <select {...selectProps} className={inputClasses.trim()}>
          {options.map(option => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      );
    }
    return <input {...(rest as TextInputProps)} className={inputClasses.trim()} />;
  };

  return (
    <div className={`${styles.inputContainer} ${containerClassName || ''}`.trim()}>
      {label && <label htmlFor={props.id || props.name} className={styles.label}>{label}</label>}
      {renderInput()}
      {error && <span className={styles.errorMessage}>{error}</span>}
    </div>
  );
};

export default Input;