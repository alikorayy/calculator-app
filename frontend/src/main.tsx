import '@fontsource/figtree/400.css';
import '@fontsource/figtree/500.css';
import '@fontsource/figtree/600.css';
import '@fontsource/figtree/700.css';
import './app/tokens.css';
import './app/global.css';

import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import { App } from './app/App';
import { ROOT_ELEMENT_ID } from './shared/constants/app';

const rootElement = document.getElementById(ROOT_ELEMENT_ID);
if (!rootElement) {
  throw new Error(`Element #${ROOT_ELEMENT_ID} is missing from index.html`);
}

createRoot(rootElement).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
