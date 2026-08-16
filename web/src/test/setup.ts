import '@testing-library/jest-dom/vitest';
import { beforeAll, afterEach, afterAll } from 'vitest';
import { server } from './mocks/server';

// Start MSW server before all tests
beforeAll(() => server.listen({ onUnhandledRequest: 'bypass' }));

// Reset handlers after each test (so tests don't share state)
afterEach(() => server.resetHandlers());

// Clean up after all tests
afterAll(() => server.close());
