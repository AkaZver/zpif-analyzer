import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, act } from '@testing-library/react';
import { ThemeProvider } from './ThemeProvider';
import { ThemeContext } from './ThemeContext';
import { useTheme } from './useTheme';

const originalMatchMedia = window.matchMedia;

describe('ThemeProvider', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove('dark');
    window.matchMedia = originalMatchMedia;
  });

  afterEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove('dark');
    window.matchMedia = originalMatchMedia;
  });

  it('should default to dark theme when no localStorage and no matchMedia preference', () => {
    let contextValue: any;

    render(
      <ThemeProvider>
        <ThemeContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </ThemeContext.Consumer>
      </ThemeProvider>
    );

    expect(contextValue.theme).toBe('dark');
  });

  it('should initialize from localStorage when set to light', () => {
    localStorage.setItem('theme', 'light');

    let contextValue: any;

    render(
      <ThemeProvider>
        <ThemeContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </ThemeContext.Consumer>
      </ThemeProvider>
    );

    expect(contextValue.theme).toBe('light');
  });

  it('should initialize from localStorage when set to dark', () => {
    localStorage.setItem('theme', 'dark');

    let contextValue: any;

    render(
      <ThemeProvider>
        <ThemeContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </ThemeContext.Consumer>
      </ThemeProvider>
    );

    expect(contextValue.theme).toBe('dark');
  });

  it('should initialize from prefers-color-scheme when no localStorage', () => {
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: (query: string) => ({
        matches: query === '(prefers-color-scheme: light)',
        media: query,
        onchange: null,
        addListener: () => {},
        removeListener: () => {},
        addEventListener: () => {},
        removeEventListener: () => {},
        dispatchEvent: () => false,
      }),
    });

    let contextValue: any;

    render(
      <ThemeProvider>
        <ThemeContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </ThemeContext.Consumer>
      </ThemeProvider>
    );

    expect(contextValue.theme).toBe('light');
  });

  it('should prioritize localStorage over system preference', () => {
    localStorage.setItem('theme', 'dark');

    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: (query: string) => ({
        matches: query === '(prefers-color-scheme: light)',
        media: query,
        onchange: null,
        addListener: () => {},
        removeListener: () => {},
        addEventListener: () => {},
        removeEventListener: () => {},
        dispatchEvent: () => false,
      }),
    });

    let contextValue: any;

    render(
      <ThemeProvider>
        <ThemeContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </ThemeContext.Consumer>
      </ThemeProvider>
    );

    expect(contextValue.theme).toBe('dark');
  });

  it('should toggle theme from dark to light', async () => {
    let contextValue: any;

    render(
      <ThemeProvider>
        <ThemeContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </ThemeContext.Consumer>
      </ThemeProvider>
    );

    const initialTheme = contextValue.theme;

    await act(async () => {
      contextValue.toggleTheme();
    });

    expect(contextValue.theme).toBe(initialTheme === 'dark' ? 'light' : 'dark');
  });

  it('should toggle theme from light to dark', async () => {
    localStorage.setItem('theme', 'light');

    let contextValue: any;

    render(
      <ThemeProvider>
        <ThemeContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </ThemeContext.Consumer>
      </ThemeProvider>
    );

    expect(contextValue.theme).toBe('light');

    await act(async () => {
      contextValue.toggleTheme();
    });

    expect(contextValue.theme).toBe('dark');
  });

  it('should save theme to localStorage on toggle', async () => {
    let contextValue: any;

    render(
      <ThemeProvider>
        <ThemeContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </ThemeContext.Consumer>
      </ThemeProvider>
    );

    const initialTheme = contextValue.theme;

    await act(async () => {
      contextValue.toggleTheme();
    });

    const expectedTheme = initialTheme === 'dark' ? 'light' : 'dark';
    expect(localStorage.getItem('theme')).toBe(expectedTheme);
  });

  it('should add dark class to documentElement when theme is dark', async () => {
    render(
      <ThemeProvider>
        <div>Test</div>
      </ThemeProvider>
    );

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 0));
    });

    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('should remove dark class from documentElement when theme is light', () => {
    localStorage.setItem('theme', 'light');

    render(
      <ThemeProvider>
        <div>Test</div>
      </ThemeProvider>
    );

    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });

  it('should toggle dark class on documentElement when toggling theme', async () => {
    let contextValue: any;

    render(
      <ThemeProvider>
        <ThemeContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </ThemeContext.Consumer>
      </ThemeProvider>
    );

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 0));
    });

    const hasDarkInitially = document.documentElement.classList.contains('dark');

    await act(async () => {
      contextValue.toggleTheme();
    });

    expect(document.documentElement.classList.contains('dark')).toBe(!hasDarkInitially);
  });
});

describe('useTheme', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove('dark');
  });

  afterEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove('dark');
  });

  it('should return theme context values when used inside ThemeProvider', () => {
    let themeValue: any;

    const TestComponent = () => {
      themeValue = useTheme();
      return <div>Test</div>;
    };

    render(
      <ThemeProvider>
        <TestComponent />
      </ThemeProvider>
    );

    expect(['dark', 'light']).toContain(themeValue.theme);
    expect(typeof themeValue.toggleTheme).toBe('function');
  });

  it('should throw error when used outside ThemeProvider', () => {
    const TestComponent = () => {
      useTheme();
      return <div>Test</div>;
    };

    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});

    expect(() => render(<TestComponent />)).toThrow(
      'useTheme must be used within ThemeProvider'
    );

    consoleError.mockRestore();
  });
});
