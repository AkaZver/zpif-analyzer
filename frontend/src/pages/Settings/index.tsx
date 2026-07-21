import React, { useEffect, useState } from 'react';
import {
  Typography, Card, Table, Button, Space, Modal, Form, Input, Select,
  Switch, message, Popconfirm, Tag,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, CloudDownloadOutlined, ReloadOutlined } from '@ant-design/icons';
import { apiClient } from '../../api/client';
import type { Fund, LLMSettings } from '../../types';
import type { ColumnsType } from 'antd/es/table';

const Settings: React.FC = () => {
  const [funds, setFunds] = useState<Fund[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingFund, setEditingFund] = useState<Fund | null>(null);
  const [llmSettings, setLlmSettings] = useState<LLMSettings | null>(null);
  const [form] = Form.useForm();
  const [llmForm] = Form.useForm();
  const [testingLlm, setTestingLlm] = useState(false);
  const [savingLlm, setSavingLlm] = useState(false);
  const [models, setModels] = useState<string[]>([]);
  const [loadingModels, setLoadingModels] = useState(false);

  useEffect(() => {
    loadFunds();
    loadLlmSettings();
    loadModels();
  }, []);

  const loadFunds = async () => {
    setLoading(true);
    try {
      const data = await apiClient.getFunds();
      setFunds(data);
    } catch {
      message.error('Не удалось загрузить фонды');
    } finally {
      setLoading(false);
    }
  };

  const loadLlmSettings = async () => {
    try {
      const settings = await apiClient.getLLMSettings();
      setLlmSettings(settings);
      llmForm.setFieldsValue(settings);
    } catch {
      // Settings not configured yet
    }
  };

  const handleAddFund = () => {
    setEditingFund(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEditFund = (fund: Fund) => {
    setEditingFund(fund);
    form.setFieldsValue(fund);
    setModalVisible(true);
  };

  const handleDeleteFund = async (id: number) => {
    try {
      await apiClient.deleteFund(id);
      message.success('Фонд удалён');
      await loadFunds();
    } catch {
      message.error('Ошибка при удалении');
    }
  };

  const handleSaveFund = async () => {
    try {
      const values = await form.validateFields();
      if (editingFund) {
        await apiClient.updateFund(editingFund.id, values);
        message.success('Фонд обновлён');
      } else {
        await apiClient.createFund(values);
        message.success('Фонд создан');
      }
      setModalVisible(false);
      await loadFunds();
    } catch {
      // Validation or API error
    }
  };

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

  const loadModels = async () => {
    setLoadingModels(true);
    try {
      const modelsList = await apiClient.getLLMModels();
      setModels(modelsList);
    } catch {
      setModels([]);
    } finally {
      setLoadingModels(false);
    }
  };

  const fundColumns: ColumnsType<Fund> = [
    { title: 'Название', dataIndex: 'name', key: 'name' },
    { title: 'ISIN', dataIndex: 'isin', key: 'isin' },
    { title: 'Тикер', dataIndex: 'ticker', key: 'ticker', render: (v: string) => v || '—' },
    { title: 'УК', dataIndex: 'management_company', key: 'management_company' },
    {
      title: 'Квал',
      dataIndex: 'qualified_required',
      key: 'qualified',
      render: (v: boolean) => <Tag color={v ? 'red' : 'green'}>{v ? 'Да' : 'Нет'}</Tag>,
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 120,
      render: (_, record) => (
        <Space>
          <Button type="text" icon={<EditOutlined />} onClick={() => handleEditFund(record)} />
          <Popconfirm title="Удалить фонд?" onConfirm={() => handleDeleteFund(record.id)}>
            <Button type="text" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Typography.Title level={3} className="text-text-primary mb-6">
        Настройки
      </Typography.Title>

      <Card
        title="Управление фондами"
        className="mb-6 bg-surface-card border-0"
        extra={<Button type="primary" icon={<PlusOutlined />} onClick={handleAddFund}>Добавить фонд</Button>}
      >
        <Table
          columns={fundColumns}
          dataSource={funds}
          rowKey="id"
          loading={loading}
          pagination={false}
        />
      </Card>

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

      <Modal
        title={editingFund ? 'Редактировать фонд' : 'Добавить фонд'}
        open={modalVisible}
        onOk={handleSaveFund}
        onCancel={() => setModalVisible(false)}
        okText="Сохранить"
        cancelText="Отмена"
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="Название" rules={[{ required: true, message: 'Введите название' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="isin" label="ISIN" rules={[{ required: true, message: 'Введите ISIN' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="ticker" label="Тикер">
            <Input />
          </Form.Item>
          <Form.Item name="management_company" label="Управляющая компания">
            <Input />
          </Form.Item>
          <Form.Item name="real_estate_segment" label="Сегмент недвижимости">
            <Select
              allowClear
              options={[
                { value: 'склады', label: 'Склады' },
                { value: 'офисы', label: 'Офисы' },
                { value: 'ТЦ', label: 'Торговые центры' },
                { value: 'ЦОД', label: 'ЦОД' },
                { value: 'жильё', label: 'Жильё' },
              ]}
            />
          </Form.Item>
          <Form.Item name="qualified_required" label="Требуется статус квала" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="has_market_maker" label="Маркет-мейкер" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Settings;
