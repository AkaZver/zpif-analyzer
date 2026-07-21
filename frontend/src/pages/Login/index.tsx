import React, { useState } from 'react';
import { Form, Input, Button, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../../api/client';

const Login: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      const response = await apiClient.login(values);
      localStorage.setItem('token', response.token);
      message.success('Вход выполнен успешно');
      navigate('/');
    } catch (error) {
      message.error('Неверный логин или пароль');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-surface">
      <div className="card w-full max-w-md">
        <h1 className="text-3xl font-bold text-text-primary mb-6 text-center">
          Вход в ZPIF Analyzer
        </h1>
        
        <Form
          name="login"
          onFinish={handleSubmit}
          layout="vertical"
          size="large"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: 'Введите имя пользователя' }]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="Имя пользователя"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: 'Введите пароль' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="Пароль"
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              className="w-full"
              loading={loading}
            >
              Войти
            </Button>
          </Form.Item>
        </Form>
      </div>
    </div>
  );
};

export default Login;
