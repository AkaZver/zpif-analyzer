import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, theme } from 'antd';
import ruRU from 'antd/locale/ru_RU';
import { AuthProvider } from './hooks/AuthProvider';
import { ThemeProvider } from './hooks/ThemeProvider';
import { useTheme } from './hooks/useTheme';
import { ProtectedRoute, PublicRoute } from './components/ProtectedRoute';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import FundDetails from './pages/FundDetails';
import Settings from './pages/Settings';
import Login from './pages/Login';

const ThemedApp: React.FC = () => {
  const { theme: currentTheme } = useTheme();

  return (
    <ConfigProvider
      locale={ruRU}
      theme={{
        algorithm: currentTheme === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
        token: {
          colorPrimary: '#7c5cbf',
          colorBgContainer: currentTheme === 'dark' ? '#333333' : '#ffffff',
          colorBgElevated: currentTheme === 'dark' ? '#2a2a2a' : '#ffffff',
          colorBgLayout: currentTheme === 'dark' ? '#1a1a1a' : '#f5f5f5',
          colorText: currentTheme === 'dark' ? '#e0e0e0' : '#1f1f1f',
          colorTextSecondary: currentTheme === 'dark' ? '#a0a0a0' : '#666666',
          borderRadius: 8,
        },
      }}
    >
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route
              path="/login"
              element={
                <PublicRoute>
                  <Login />
                </PublicRoute>
              }
            />
            <Route
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route path="/" element={<Dashboard />} />
              <Route path="/funds/:id" element={<FundDetails />} />
              <Route path="/settings" element={<Settings />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </ConfigProvider>
  );
};

const App: React.FC = () => {
  return (
    <ThemeProvider>
      <ThemedApp />
    </ThemeProvider>
  );
};

export default App;
