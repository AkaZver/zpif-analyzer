import React, { useCallback, useEffect, useState } from 'react';
import {
  Typography, Card, Button, Space, Form, Input, Select,
  message, Checkbox, Tooltip,
} from 'antd';
import { CloudDownloadOutlined, ReloadOutlined, CheckCircleOutlined, CloseCircleOutlined, LoadingOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import { apiClient } from '../../api/client';
import type { LLMSettings } from '../../types';

interface ModelTestResult {
  success: boolean;
  message: string;
}

const Settings: React.FC = () => {
  const [llmSettings, setLlmSettings] = useState<LLMSettings | null>(null);
  const [llmForm] = Form.useForm();
  const [testingLlm, setTestingLlm] = useState(false);
  const [savingLlm, setSavingLlm] = useState(false);
  const [models, setModels] = useState<string[]>([]);
  const [loadingModels, setLoadingModels] = useState(false);
  const [proxyEnabled, setProxyEnabled] = useState(false);
  const [testResults, setTestResults] = useState<{
    search: ModelTestResult | null;
    analysis: ModelTestResult | null;
  }>({ search: null, analysis: null });

  const loadLlmSettings = useCallback(async () => {
    try {
      const settings = await apiClient.getLLMSettings();
      setLlmSettings(settings);
      llmForm.setFieldsValue(settings);
      setProxyEnabled(settings.proxy_enabled || false);
    } catch {
    }
  }, [llmForm]);

  const loadModels = useCallback(async () => {
    setLoadingModels(true);
    try {
      const modelsList = await apiClient.getLLMModels();
      setModels([...modelsList].sort((a, b) => a.localeCompare(b)));
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
    setTestResults({ search: null, analysis: null });
    try {
      const result = await apiClient.testLLMConnection();
      setTestResults({
        search: result.search_model,
        analysis: result.analysis_model,
      });
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

      <Card title="Настройки LLM" className="mb-6 bg-white dark:bg-[#333333] border-0">
        <Form 
          form={llmForm} 
          layout="vertical" 
          initialValues={llmSettings || {}}
          onValuesChange={(changedValues) => {
            if ('search_model_name' in changedValues) {
              setTestResults(prev => ({ ...prev, search: null }));
            }
            if ('analysis_model_name' in changedValues) {
              setTestResults(prev => ({ ...prev, analysis: null }));
            }
          }}
        >
          <Form.Item name="api_key_encrypted" label="API Key">
            <Input.Password placeholder="sk-..." />
          </Form.Item>
          <Form.Item name="base_url" label="Base URL">
            <Input placeholder="https://api.openai.com/v1" />
          </Form.Item>
          <Form.Item 
            name="search_model_name" 
            label={
              <span>
                Модель для поиска
                <Tooltip title="Используется для поиска документов в интернете">
                  <QuestionCircleOutlined style={{ marginLeft: 4, color: '#888' }} />
                </Tooltip>
                {testingLlm && <LoadingOutlined style={{ marginLeft: 8 }} />}
                {!testingLlm && testResults.search && (
                  <Tooltip title={testResults.search.message}>
                    {testResults.search.success ? (
                      <CheckCircleOutlined style={{ color: '#52c41a', marginLeft: 8 }} />
                    ) : (
                      <CloseCircleOutlined style={{ color: '#ff4d4f', marginLeft: 8 }} />
                    )}
                  </Tooltip>
                )}
              </span>
            }
          >
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
          <Form.Item 
            name="analysis_model_name" 
            label={
              <span>
                Модель для анализа
                <Tooltip title="Используется для анализа документов и извлечения метрик">
                  <QuestionCircleOutlined style={{ marginLeft: 4, color: '#888' }} />
                </Tooltip>
                {testingLlm && <LoadingOutlined style={{ marginLeft: 8 }} />}
                {!testingLlm && testResults.analysis && (
                  <Tooltip title={testResults.analysis.message}>
                    {testResults.analysis.success ? (
                      <CheckCircleOutlined style={{ color: '#52c41a', marginLeft: 8 }} />
                    ) : (
                      <CloseCircleOutlined style={{ color: '#ff4d4f', marginLeft: 8 }} />
                    )}
                  </Tooltip>
                )}
              </span>
            }
          >
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
