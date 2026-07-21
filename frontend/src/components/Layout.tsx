import React from 'react';
import { Layout as AntLayout, Button, Typography, Dropdown } from 'antd';
import {
  DashboardOutlined,
  SettingOutlined,
  LogoutOutlined,
  LoginOutlined,
  MenuOutlined,
} from '@ant-design/icons';
import { useNavigate, Outlet } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import type { MenuProps } from 'antd';

const { Header, Content } = AntLayout;

const Layout: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated, logout } = useAuth();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const menuItems: MenuProps['items'] = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: 'Сравнение',
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: 'Настройки',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Выйти',
      danger: true,
    },
  ];

  const handleMenuClick: MenuProps['onClick'] = (e) => {
    if (e.key === 'logout') {
      handleLogout();
    } else {
      navigate(e.key);
    }
  };

  return (
    <AntLayout className="min-h-screen">
      <Header className="bg-[#2a2a2a] px-6 flex items-center justify-between h-16 border-b border-[#3a3a3a]">
        <Typography.Title
          level={4}
          className="text-primary m-0 cursor-pointer"
          onClick={() => navigate('/')}
        >
          ZPIF Analyzer
        </Typography.Title>
        {isAuthenticated ? (
          <Dropdown
            menu={{ items: menuItems, onClick: handleMenuClick }}
            trigger={['click']}
          >
            <Button type="text" icon={<MenuOutlined />} className="text-text-primary" />
          </Dropdown>
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
