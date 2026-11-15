/// <reference types="vitest/config" />
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

/** Path prefix of the backend API; the browser only ever talks to its own origin. */
const API_PATH = '/api';

/**
 * Where the dev server forwards API calls: the Go service started with `go run`.
 * Override with API_PROXY_TARGET when the API runs elsewhere.
 */
const DEFAULT_API_TARGET = 'http://localhost:8080';

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      [API_PATH]: process.env.API_PROXY_TARGET ?? DEFAULT_API_TARGET,
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/setupTests.ts'],
    coverage: {
      provider: 'v8',
      // text: table in the terminal; html: browsable report; json-summary: totals for CI.
      reporter: ['text', 'html', 'json-summary'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/main.tsx', 'src/setupTests.ts', 'src/mocks/**', 'src/**/*.d.ts'],
    },
  },
});
