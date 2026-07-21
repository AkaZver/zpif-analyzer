import React, { useEffect, useState } from 'react';
import { Table, Button, Space, Tag, message, Spin, Select, Checkbox, Typography, Card } from 'antd';
import { DownloadOutlined, ReloadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { apiClient } from '../../api/client';
import type { Fund, FundFinancials } from '../../types';
import type { ColumnsType } from 'antd/es/table';

interface FundWithFinancials extends Fund {
  latest_financials?: FundFinancials | null;
}

const Dashboard: React.FC = () => {
  const [funds, setFunds] = useState<FundWithFinancials[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [filterSegment, setFilterSegment] = useState<string | undefined>();
  const [filterCompany, setFilterCompany] = useState<string | undefined>();
  const [filterQualified, setFilterQualified] = useState<boolean | undefined>();
  const navigate = useNavigate();

  useEffect(() => {
    loadFunds();
  }, []);

  const loadFunds = async () => {
    setLoading(true);
    try {
      const fundsData = await apiClient.getFunds();
      const fundsWithFinancials: FundWithFinancials[] = await Promise.all(
        fundsData.map(async (fund) => {
          try {
            const financials = await apiClient.getFinancials(fund.id);
            return {
              ...fund,
              latest_financials: financials.length > 0 ? financials[0] : null,
            };
          } catch {
            return { ...fund, latest_financials: null };
          }
        })
      );
      setFunds(fundsWithFinancials);
    } catch (error) {
      message.error('Не удалось загрузить фонды');
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await apiClient.discoverAll();
      message.success('Автопоиск документов запущен');
      setTimeout(() => loadFunds(), 2000);
    } catch (error) {
      message.error('Ошибка при обновлении');
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
      a.click();
      window.URL.revokeObjectURL(url);
      message.success('Экспорт завершён');
    } catch {
      message.error('Ошибка при экспорте');
    }
  };

  const segments = [...new Set(funds.map((f) => f.real_estate_segment).filter(Boolean))];
  const companies = [...new Set(funds.map((f) => f.management_company).filter(Boolean))];

  const filteredFunds = funds.filter((fund) => {
    if (filterSegment && fund.real_estate_segment !== filterSegment) return false;
    if (filterCompany && fund.management_company !== filterCompany) return false;
    if (filterQualified !== undefined && fund.qualified_required !== filterQualified) return false;
    return true;
  });

  const renderPctCell = (value: number) => {
    const color = value >= 0 ? '#52c41a' : '#ff4d4f';
    return <span style={{ color }}>{value?.toFixed(1)}%</span>;
  };

  const renderDiscountCell = (value: number) => {
    const color = value <= 0 ? '#52c41a' : '#ff4d4f';
    return <span style={{ color }}>{value?.toFixed(1)}%</span>;
  };

  const columns: ColumnsType<FundWithFinancials> = [
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
      sorter: (a, b) => a.name.localeCompare(b.name),
      fixed: 'left',
      width: 180,
    },
    { title: 'ISIN', dataIndex: 'isin', key: 'isin', width: 140 },
    {
      title: 'Тикер',
      dataIndex: 'ticker',
      key: 'ticker',
      width: 80,
      render: (v: string) => v || '—',
    },
    { title: 'УК', dataIndex: 'management_company', key: 'management_company', width: 150 },
    {
      title: 'Сегмент',
      dataIndex: 'real_estate_segment',
      key: 'real_estate_segment',
      width: 100,
      render: (v: string) => v || '—',
    },
    {
      title: 'Цена пая',
      key: 'unit_price',
      width: 100,
      render: (_, r) => r.latest_financials?.unit_price_rub?.toFixed(0) || '—',
      sorter: (a, b) => (a.latest_financials?.unit_price_rub || 0) - (b.latest_financials?.unit_price_rub || 0),
    },
    {
      title: 'РСП',
      key: 'nav',
      width: 100,
      render: (_, r) => r.latest_financials?.nav_per_unit_rub?.toFixed(0) || '—',
    },
    {
      title: 'Дисконт',
      key: 'discount',
      width: 90,
      render: (_, r) => r.latest_financials ? renderDiscountCell(r.latest_financials.discount_to_nav_pct) : '—',
      sorter: (a, b) => (a.latest_financials?.discount_to_nav_pct || 0) - (b.latest_financials?.discount_to_nav_pct || 0),
    },
    {
      title: 'Cap Rate',
      key: 'cap_rate',
      width: 90,
      render: (_, r) => r.latest_financials?.cap_rate_pct ? `${r.latest_financials.cap_rate_pct.toFixed(1)}%` : '—',
    },
    {
      title: 'P/NAV',
      key: 'p_nav',
      width: 80,
      render: (_, r) => r.latest_financials?.p_nav?.toFixed(2) || '—',
    },
    {
      title: 'Доходность выплат',
      key: 'payout_yield',
      width: 130,
      render: (_, r) => r.latest_financials ? renderPctCell(r.latest_financials.payout_yield_pct) : '—',
      sorter: (a, b) => (a.latest_financials?.payout_yield_pct || 0) - (b.latest_financials?.payout_yield_pct || 0),
    },
    {
      title: 'Полная доходность',
      key: 'total_return',
      width: 140,
      render: (_, r) => r.latest_financials ? renderPctCell(r.latest_financials.total_return_pct) : '—',
      sorter: (a, b) => (a.latest_financials?.total_return_pct || 0) - (b.latest_financials?.total_return_pct || 0),
    },
    {
      title: 'Долг/СЧА',
      key: 'debt',
      width: 90,
      render: (_, r) => r.latest_financials?.debt_to_nav_ratio?.toFixed(2) || '—',
    },
    {
      title: 'Квал',
      dataIndex: 'qualified_required',
      key: 'qualified',
      width: 60,
      render: (v: boolean) => <Tag color={v ? 'red' : 'green'}>{v ? 'Да' : 'Нет'}</Tag>,
    },
  ];

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <Typography.Title level={3} className="text-text-primary m-0">
          Сравнение ЗПИФ
        </Typography.Title>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh} loading={refreshing}>
            Обновить
          </Button>
          <Button icon={<DownloadOutlined />} onClick={handleExport}>
            Экспорт
          </Button>
        </Space>
      </div>

      <Card className="mb-4 bg-surface-card border-0">
        <Space wrap>
          <Select
            placeholder="Сегмент"
            allowClear
            style={{ width: 160 }}
            value={filterSegment}
            onChange={setFilterSegment}
            options={segments.map((s) => ({ value: s, label: s }))}
          />
          <Select
            placeholder="УК"
            allowClear
            style={{ width: 200 }}
            value={filterCompany}
            onChange={setFilterCompany}
            options={companies.map((c) => ({ value: c, label: c }))}
          />
          <Checkbox
            checked={filterQualified === true}
            onChange={(e) => setFilterQualified(e.target.checked ? true : undefined)}
          >
            Только для квалов
          </Checkbox>
        </Space>
      </Card>

      {loading ? (
        <div className="flex justify-center items-center h-64">
          <Spin size="large" />
        </div>
      ) : (
        <Table
          columns={columns}
          dataSource={filteredFunds}
          rowKey="id"
          pagination={false}
          scroll={{ x: 1600 }}
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
