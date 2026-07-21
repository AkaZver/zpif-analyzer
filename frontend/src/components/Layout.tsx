import React, { useState } from 'react';
import { Layout as AntLayout, Menu, Button, Space, Typography } from 'antd';
import {
  DashboardOutlined,
  SettingOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
  LoginOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

const { Header, Sider, Content } = AntLayout;

const Layout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { isAuthenticated, logout } = useAuth();

  const menuItems = [
    { key: '/', icon: <DashboardOutlined />, label: 'Сравнение' },
    { key: '/settings', icon: <SettingOutlined />, label: 'Настройки' },
  ];

  const handleMenuClick = (e: { key: string }) => {
    navigate(e.key);
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <AntLayout className="min-h-screen">
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        className="bg-[#0a0a1a]"
        width={240}
        trigger={null}
      >
        <div className="flex items-center justify-center h-16 px-4">
          <Typography.Title level={4} className="text-primary m-0">
            {collapsed ? 'ZA' : 'ZPIF Analyzer'}
          </Typography.Title>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
          className="bg-[#0a0a1a] border-r-0"
        />
      </Sider>
      <AntLayout>
        <Header className="bg-[#0f3460] px-4 flex items-center justify-between h-16">
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            className="text-text-primary text-lg"
          />
          <Space>
            {isAuthenticated ? (
              <Button
                type="text"
                icon={<LogoutOutlined />}
                onClick={handleLogout}
                className="text-text-primary"
              >
                Выйти
              </Button>
            ) : (
              <Button
                type="primary"
                icon={<LoginOutlined />}
                onClick={() => navigate('/login')}
              >
                Войти
              </Button>
            )}
          </Space>
        </Header>
        <Content className="p-6 bg-[#1a1a2e] min-h-[calc(100vh-64px)] overflow-auto">
          <Outlet />
        </Content>
      </AntLayout>
    </AntLayout>
  );
};

export default Layout;
