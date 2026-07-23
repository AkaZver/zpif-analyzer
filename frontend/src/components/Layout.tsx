import React from 'react';
import { Layout as AntLayout, Button, Typography } from 'antd';
import {
  SettingOutlined,
  LogoutOutlined,
  LoginOutlined,
} from '@ant-design/icons';
import { useNavigate, Outlet } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import buildingIcon from '../assets/building-icon.svg';

const { Header, Content } = AntLayout;

const Layout: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated, logout } = useAuth();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <AntLayout className="min-h-screen">
      <Header className="bg-[#2a2a2a] px-6 flex items-center justify-between h-16 border-b border-[#3a3a3a]">
        <div className="flex items-center gap-2 cursor-pointer" onClick={() => navigate('/')}>
          <div className="h-8 flex items-center">
            <img src={buildingIcon} alt="ZPIF" className="h-full" />
          </div>
          <Typography.Title level={4} className="text-primary m-0 leading-tight">
            ZPIF Analyzer
          </Typography.Title>
        </div>
        {isAuthenticated ? (
          <div className="flex items-center gap-2">
            <Button
              type="text"
              icon={<SettingOutlined />}
              className="text-text-primary"
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
      <Content className="bg-[#1a1a1a] min-h-[calc(100vh-64px)] overflow-auto">
        <div className="max-w-[1400px] mx-auto p-6">
          <Outlet />
        </div>
      </Content>
    </AntLayout>
  );
};

export default Layout;
