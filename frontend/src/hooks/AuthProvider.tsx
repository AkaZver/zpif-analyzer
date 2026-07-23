import React, { useState, useCallback } from 'react';
import type { ReactNode } from 'react';
import { apiClient } from '../api/client';
import { AuthContext } from './AuthContext';

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [token, setToken] = useState<string | null>(
    localStorage.getItem('token')
  );

  const login = useCallback(async (username: string, password: string) => {
    const response = await apiClient.login({ username, password });
    localStorage.setItem('token', response.token);
    setToken(response.token);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    setToken(null);
  }, []);

  return (
    <AuthContext.Provider
      value={{ isAuthenticated: !!token, token, login, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
};
