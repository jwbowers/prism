import { createContext, useContext } from 'react';
import { SafePrismAPI } from '../lib/api';

const ApiContext = createContext<SafePrismAPI | null>(null);

export { ApiContext };

export function useApi(): SafePrismAPI {
  const api = useContext(ApiContext);
  if (!api) {
    throw new Error('useApi must be used within an ApiProvider');
  }
  return api;
}
