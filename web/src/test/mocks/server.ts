import { setupServer } from 'msw/node';
import { handlers } from './handlers';

/**
 * MSW server for Vitest (Node.js / jsdom).
 * Import this in vitest setup or individual test files.
 */
export const server = setupServer(...handlers);
