import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Layout from './Layout';
import { useAuth } from '../hooks/useAuth';

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
  });

  it('should render layout with outlet', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: false,
      token: null,
      login: vi.fn(),
      logout: vi.fn(),
    });

    render(
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
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
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
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
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
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
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
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
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
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
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
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
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
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
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
    );

    const buttons = screen.getAllByRole('button');
    const logoutButton = buttons[buttons.length - 1];
    fireEvent.click(logoutButton);

    expect(mockLogout).toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith('/login');
  });
});
