import React from 'react';
import { Layout as AntLayout, Button, Typography } from 'antd';
import {
  SettingOutlined,
  LogoutOutlined,
  LoginOutlined,
  SunOutlined,
  MoonOutlined,
} from '@ant-design/icons';
import { useNavigate, Outlet } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { useTheme } from '../hooks/useTheme';
import buildingIcon from '../assets/building-icon.svg';

const { Header, Content } = AntLayout;

const Layout: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <AntLayout className="min-h-screen">
      <Header className="bg-white dark:bg-[#2a2a2a] px-6 flex items-center justify-between h-16 border-b border-gray-200 dark:border-[#3a3a3a]">
        <div
          className="inline-flex items-center cursor-pointer"
          onClick={() => navigate('/')}
          onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') navigate('/'); }}
          role="button"
          tabIndex={0}
        >
          <img src={buildingIcon} alt="ZPIF" className="h-8 mr-2 -translate-y-1" />
          <Typography.Title level={4} className="text-primary m-0 p-0 leading-none">
            ZPIF Analyzer
          </Typography.Title>
        </div>
        {isAuthenticated ? (
          <div className="flex items-center gap-2">
            <Button
              type="text"
              icon={theme === 'dark' ? <SunOutlined /> : <MoonOutlined />}
              onClick={toggleTheme}
              aria-label="Переключить тему"
            />
            <Button
              type="text"
              icon={<SettingOutlined />}
              onClick={() => navigate('/settings')}
            />
            <Button
              type="text"
              icon={<LogoutOutlined />}
              danger
              onClick={handleLogout}
            />
          </div>
        ) : (
          <Button
            type="primary"
            icon={<LoginOutlined />}
            onClick={() => navigate('/login')}
          >
            Войти
          </Button>
        )}
      </Header>
      <Content className="bg-gray-50 dark:bg-[#1a1a1a] min-h-[calc(100vh-64px)] overflow-auto">
        <div className="max-w-[1400px] mx-auto p-6">
          <Outlet />
        </div>
      </Content>
    </AntLayout>
  );
};

export default Layout;
