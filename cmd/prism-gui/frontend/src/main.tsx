import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { ApiContext } from './hooks/use-api';
import { SafePrismAPI } from './lib/api';

const api = new SafePrismAPI();

const root = ReactDOM.createRoot(document.getElementById('root')!);
root.render(
  <React.StrictMode>
    <ApiContext.Provider value={api}>
      <App />
    </ApiContext.Provider>
  </React.StrictMode>
);
