import React, { useCallback, useEffect, useState } from 'react';
import {
  Typography, Card, Button, Space, Form, Input, Select,
  message, Checkbox,
} from 'antd';
import { CloudDownloadOutlined, ReloadOutlined } from '@ant-design/icons';
import { apiClient } from '../../api/client';
import type { LLMSettings } from '../../types';

const Settings: React.FC = () => {
  const [llmSettings, setLlmSettings] = useState<LLMSettings | null>(null);
  const [llmForm] = Form.useForm();
  const [testingLlm, setTestingLlm] = useState(false);
  const [savingLlm, setSavingLlm] = useState(false);
  const [models, setModels] = useState<string[]>([]);
  const [loadingModels, setLoadingModels] = useState(false);
  const [proxyEnabled, setProxyEnabled] = useState(false);

  const loadLlmSettings = useCallback(async () => {
    try {
      const settings = await apiClient.getLLMSettings();
      setLlmSettings(settings);
      llmForm.setFieldsValue({
        ...settings,
        proxy_password: '',
      });
      setProxyEnabled(settings.proxy_enabled || false);
    } catch {
    }
  }, [llmForm]);

  const loadModels = useCallback(async () => {
    setLoadingModels(true);
    try {
      const modelsList = await apiClient.getLLMModels();
      setModels(modelsList);
    } catch {
      setModels([]);
    } finally {
      setLoadingModels(false);
    }
  }, []);

  useEffect(() => {
    loadLlmSettings();
    loadModels();
  }, [loadLlmSettings, loadModels]);

  const handleSaveLlmSettings = async () => {
    setSavingLlm(true);
    try {
      const values = await llmForm.validateFields();
      await apiClient.updateLLMSettings(values);
      message.success('Настройки сохранены');
    } catch {
      message.error('Ошибка при сохранении');
    } finally {
      setSavingLlm(false);
    }
  };

  const handleTestLlm = async () => {
    setTestingLlm(true);
    try {
      const result = await apiClient.testLLMConnection();
      message.success(result.message);
    } catch {
      message.error('Ошибка подключения к LLM');
    } finally {
      setTestingLlm(false);
    }
  };

  return (
    <div>
      <Typography.Title level={3} className="text-text-primary mb-6">
        Настройки
      </Typography.Title>

      <Card title="Настройки LLM" className="mb-6 bg-surface-card border-0">
        <Form form={llmForm} layout="vertical" initialValues={llmSettings || {}}>
          <Form.Item name="api_key_encrypted" label="API Key">
            <Input.Password placeholder="sk-..." />
          </Form.Item>
          <Form.Item name="base_url" label="Base URL">
            <Input placeholder="https://api.openai.com/v1" />
          </Form.Item>
          <Form.Item name="model_name" label="Модель">
            {models.length > 0 ? (
              <Select
                showSearch
                placeholder="Выберите модель"
                options={models.map((m) => ({ value: m, label: m }))}
                filterOption={(input, option) =>
                  (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                }
              />
            ) : (
              <Input placeholder="gpt-4o-mini" />
            )}
          </Form.Item>
          <Form.Item name="proxy_enabled" valuePropName="checked">
            <Checkbox onChange={(e) => setProxyEnabled(e.target.checked)}>
              Использовать прокси
            </Checkbox>
          </Form.Item>
          <Card title="Прокси" className="mb-4" size="small">
            <Form.Item name="proxy_url" label="URL прокси">
              <Input placeholder="http://proxy.example.com:8080" disabled={!proxyEnabled} />
            </Form.Item>
            <Form.Item name="proxy_username" label="Логин">
              <Input placeholder="username" disabled={!proxyEnabled} />
            </Form.Item>
            <Form.Item name="proxy_password" label="Пароль">
              <Input.Password placeholder="password" disabled={!proxyEnabled} />
            </Form.Item>
          </Card>
          <Space>
            <Button type="primary" onClick={handleSaveLlmSettings} loading={savingLlm}>
              Сохранить
            </Button>
            <Button icon={<CloudDownloadOutlined />} onClick={handleTestLlm} loading={testingLlm}>
              Тест LLM
            </Button>
            <Button icon={<ReloadOutlined />} onClick={loadModels} loading={loadingModels}>
              Загрузить модели
            </Button>
          </Space>
        </Form>
      </Card>
    </div>
  );
};

export default Settings;
