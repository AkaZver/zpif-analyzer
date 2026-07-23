import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, act } from '@testing-library/react';
import { AuthProvider } from './AuthProvider';
import { AuthContext } from './AuthContext';
import { apiClient } from '../api/client';

vi.mock('../api/client');

describe('AuthProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it('should provide initial auth state when no token', () => {
    let contextValue: any;
    
    render(
      <AuthProvider>
        <AuthContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </AuthContext.Consumer>
      </AuthProvider>
    );

    expect(contextValue.isAuthenticated).toBe(false);
    expect(contextValue.token).toBe(null);
  });

  it('should provide initial auth state when token exists', () => {
    localStorage.setItem('token', 'existing-token');
    
    let contextValue: any;
    
    render(
      <AuthProvider>
        <AuthContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </AuthContext.Consumer>
      </AuthProvider>
    );

    expect(contextValue.isAuthenticated).toBe(true);
    expect(contextValue.token).toBe('existing-token');
  });

  it('should handle login', async () => {
    vi.mocked(apiClient.login).mockResolvedValue({
      token: 'new-token',
      user: {
        id: 1,
        username: 'admin',
        email: 'admin@test.com',
        is_active: true,
        created_at: '2024-01-01',
        updated_at: '2024-01-01',
      },
    });
    
    let contextValue: any;
    
    render(
      <AuthProvider>
        <AuthContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </AuthContext.Consumer>
      </AuthProvider>
    );

    await act(async () => {
      await contextValue.login('admin', 'admin');
    });

    expect(apiClient.login).toHaveBeenCalledWith({ username: 'admin', password: 'admin' });
    expect(localStorage.getItem('token')).toBe('new-token');
    expect(contextValue.isAuthenticated).toBe(true);
    expect(contextValue.token).toBe('new-token');
  });

  it('should handle logout', async () => {
    localStorage.setItem('token', 'test-token');
    
    let contextValue: any;
    
    render(
      <AuthProvider>
        <AuthContext.Consumer>
          {(value) => {
            contextValue = value;
            return <div>Test</div>;
          }}
        </AuthContext.Consumer>
      </AuthProvider>
    );

    act(() => {
      contextValue.logout();
    });

    expect(localStorage.getItem('token')).toBe(null);
    expect(contextValue.isAuthenticated).toBe(false);
    expect(contextValue.token).toBe(null);
  });
});
