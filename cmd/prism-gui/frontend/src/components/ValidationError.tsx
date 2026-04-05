import React from 'react';
import { Alert } from '../lib/cloudscape-shim';

interface ValidationErrorProps {
  message: string;
  visible: boolean;
}

export const ValidationError: React.FC<ValidationErrorProps> = ({ message, visible }) => {
  if (!visible) return null;

  return (
    <Alert
      type="error"
      data-testid="validation-error"
    >
      {message}
    </Alert>
  );
};
