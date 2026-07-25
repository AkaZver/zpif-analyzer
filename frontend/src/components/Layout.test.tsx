import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Layout from './Layout';
import { useAuth } from '../hooks/useAuth';
import { ThemeProvider } from '../hooks/ThemeProvider';

vi.mock('../hooks/useAuth');
vi.mock('../assets/building-icon.svg', () => ({
  default: 'mocked-icon.svg',
}));

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    Outlet: () => <div data-testid="outlet">Outlet Content</div>,
  };
});

describe('Layout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it('should render layout with outlet', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: false,
      token: null,
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    expect(screen.getByText('ZPIF Analyzer')).toBeInTheDocument();
    expect(screen.getByTestId('outlet')).toBeInTheDocument();
  });

  it('should show login button when not authenticated', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: false,
      token: null,
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    expect(screen.getByText('Войти')).toBeInTheDocument();
  });

  it('should show settings and logout buttons when authenticated', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: true,
      token: 'test-token',
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    expect(screen.queryByText('Войти')).not.toBeInTheDocument();
  });

  it('should navigate to home on logo click', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: false,
      token: null,
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    const logo = screen.getByText('ZPIF Analyzer').closest('[role="button"]');
    fireEvent.click(logo!);

    expect(mockNavigate).toHaveBeenCalledWith('/');
  });

  it('should navigate to home on Enter key press', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: false,
      token: null,
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    const logo = screen.getByText('ZPIF Analyzer').closest('[role="button"]');
    fireEvent.keyDown(logo!, { key: 'Enter' });

    expect(mockNavigate).toHaveBeenCalledWith('/');
  });

  it('should navigate to home on Space key press', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: false,
      token: null,
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    const logo = screen.getByText('ZPIF Analyzer').closest('[role="button"]');
    fireEvent.keyDown(logo!, { key: ' ' });

    expect(mockNavigate).toHaveBeenCalledWith('/');
  });

  it('should navigate to login on login button click', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: false,
      token: null,
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    const loginButton = screen.getByText('Войти');
    fireEvent.click(loginButton);

    expect(mockNavigate).toHaveBeenCalledWith('/login');
  });

  it('should call logout and navigate on logout button click', () => {
    const mockLogout = vi.fn();
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: true,
      token: 'test-token',
      login: vi.fn(),
      logout: mockLogout,
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    const buttons = screen.getAllByRole('button');
    const logoutButton = buttons[buttons.length - 1];
    fireEvent.click(logoutButton);

    expect(mockLogout).toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith('/login');
  });

  it('should show theme toggle button when authenticated', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: true,
      token: 'test-token',
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    const themeButton = screen.getByLabelText('Переключить тему');
    expect(themeButton).toBeInTheDocument();
  });

  it('should not show theme toggle button when not authenticated', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: false,
      token: null,
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    const themeButton = screen.queryByLabelText('Переключить тему');
    expect(themeButton).not.toBeInTheDocument();
  });

  it('should toggle theme when theme button is clicked', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: true,
      token: 'test-token',
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    const themeButton = screen.getByLabelText('Переключить тему');
    expect(document.documentElement.classList.contains('dark')).toBe(true);

    fireEvent.click(themeButton);

    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });

  it('should show sun icon in dark theme and moon icon in light theme', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: true,
      token: 'test-token',
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    const themeButton = screen.getByLabelText('Переключить тему');
    
    expect(themeButton.querySelector('.anticon-sun')).toBeInTheDocument();
    expect(themeButton.querySelector('.anticon-moon')).not.toBeInTheDocument();

    fireEvent.click(themeButton);

    expect(themeButton.querySelector('.anticon-moon')).toBeInTheDocument();
    expect(themeButton.querySelector('.anticon-sun')).not.toBeInTheDocument();
  });

  it('should position theme button before settings button', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: true,
      token: 'test-token',
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    const buttons = screen.getAllByRole('button');
    const themeButton = screen.getByLabelText('Переключить тему');
    const settingsButton = buttons.find(b => b.querySelector('.anticon-setting'));

    const themeIndex = buttons.indexOf(themeButton);
    const settingsIndex = buttons.indexOf(settingsButton!);

    expect(themeIndex).toBeLessThan(settingsIndex);
  });

  it('should navigate to settings on settings button click', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: true,
      token: 'test-token',
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <ThemeProvider>
        <MemoryRouter>
          <Layout />
        </MemoryRouter>
      </ThemeProvider>
    );

    const buttons = screen.getAllByRole('button');
    const settingsButton = buttons.find(b => b.querySelector('.anticon-setting'));
    fireEvent.click(settingsButton!);

    expect(mockNavigate).toHaveBeenCalledWith('/settings');
  });
});
