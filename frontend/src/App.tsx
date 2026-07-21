import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, theme } from 'antd';
import ruRU from 'antd/locale/ru_RU';
import Dashboard from './pages/Dashboard';
import FundDetails from './pages/FundDetails';
import Settings from './pages/Settings';
import Login from './pages/Login';

const App: React.FC = () => {
  return (
    <ConfigProvider
      locale={ruRU}
      theme={{
        algorithm: theme.darkAlgorithm,
        token: {
          colorPrimary: '#7c5cbf',
          colorBgContainer: '#0f3460',
          colorBgElevated: '#16213e',
          colorBgLayout: '#1a1a2e',
          colorText: '#e0e0e0',
          colorTextSecondary: '#a0a0a0',
          borderRadius: 8,
        },
      }}
    >
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/" element={<Dashboard />} />
          <Route path="/funds/:id" element={<FundDetails />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </ConfigProvider>
  );
};

export default App;
