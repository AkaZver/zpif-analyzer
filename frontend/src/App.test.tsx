import { describe, it, expect, beforeEach } from 'vitest';
import { render } from '@testing-library/react';
import App from './App';

vi.mock('./hooks/useAuth', () => ({
  useAuth: () => ({
    isAuthenticated: false,
    token: null,
    login: vi.fn(),
    logout: vi.fn(),
  }),
}));

describe('App', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove('dark');
  });

  it('should use darkAlgorithm when theme is dark', () => {
    render(<App />);

    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('should use defaultAlgorithm when theme is light', () => {
    localStorage.setItem('theme', 'light');

    render(<App />);

    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });
});
