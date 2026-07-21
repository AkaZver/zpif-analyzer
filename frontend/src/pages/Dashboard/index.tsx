import React, { useEffect, useState } from 'react';
import { Table, Button, Space, Tag, message, Spin } from 'antd';
import { DownloadOutlined, UploadOutlined, ReloadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../../api/client';
import { Fund } from '../../types';

const Dashboard: React.FC = () => {
  const [funds, setFunds] = useState<Fund[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    loadFunds();
  }, []);

  const loadFunds = async () => {
    setLoading(true);
    try {
      const data = await apiClient.getFunds();
      setFunds(data);
    } catch (error) {
      message.error('Не удалось загрузить фонды');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await apiClient.discoverAll();
      message.success('Автопоиск документов запущен');
      await loadFunds();
    } catch (error) {
      message.error('Ошибка при обновлении');
      console.error(error);
    } finally {
      setRefreshing(false);
    }
  };

  const handleExport = async () => {
    try {
      const blob = await apiClient.exportExcel();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'zpif-analyzer-export.xlsx';
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      message.success('Экспорт завершён');
    } catch (error) {
      message.error('Ошибка при экспорте');
      console.error(error);
    }
  };

  const handleImport = () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.xlsx,.xls';
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;

      try {
        const result = await apiClient.importExcel(file);
        message.success(`Импортировано: ${result.imported} записей`);
        await loadFunds();
      } catch (error) {
        message.error('Ошибка при импорте');
        console.error(error);
      }
    };
    input.click();
  };

  const columns = [
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
      sorter: (a: Fund, b: Fund) => a.name.localeCompare(b.name),
    },
    {
      title: 'ISIN',
      dataIndex: 'isin',
      key: 'isin',
    },
    {
      title: 'Тикер',
      dataIndex: 'ticker',
      key: 'ticker',
      render: (ticker: string) => ticker || '—',
    },
    {
      title: 'УК',
      dataIndex: 'management_company',
      key: 'management_company',
    },
    {
      title: 'Сегмент',
      dataIndex: 'real_estate_segment',
      key: 'real_estate_segment',
      render: (segment: string) => segment || '—',
    },
    {
      title: 'Квал.',
      dataIndex: 'qualified_required',
      key: 'qualified_required',
      render: (required: boolean) => (
        <Tag color={required ? 'red' : 'green'}>
          {required ? 'Да' : 'Нет'}
        </Tag>
      ),
    },
    {
      title: 'ММ',
      dataIndex: 'has_market_maker',
      key: 'has_market_maker',
      render: (hasMM: boolean) => (
        <Tag color={hasMM ? 'green' : 'default'}>
          {hasMM ? 'Да' : 'Нет'}
        </Tag>
      ),
    },
  ];

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-text-primary">Сравнение ЗПИФ</h1>
        <Space>
          <Button
            icon={<ReloadOutlined />}
            onClick={handleRefresh}
            loading={refreshing}
          >
            Обновить данные
          </Button>
          <Button
            icon={<DownloadOutlined />}
            onClick={handleExport}
          >
            Экспорт в Excel
          </Button>
          <Button
            icon={<UploadOutlined />}
            onClick={handleImport}
          >
            Импорт из Excel
          </Button>
        </Space>
      </div>

      {loading ? (
        <div className="flex justify-center items-center h-64">
          <Spin size="large" />
        </div>
      ) : (
        <Table
          columns={columns}
          dataSource={funds}
          rowKey="id"
          pagination={false}
          onRow={(record) => ({
            onClick: () => navigate(`/funds/${record.id}`),
            style: { cursor: 'pointer' },
          })}
          className="bg-surface-card rounded-lg"
        />
      )}
    </div>
  );
};

export default Dashboard;
